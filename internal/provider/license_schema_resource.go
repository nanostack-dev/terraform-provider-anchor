package provider

import (
	"context"
	"fmt"
	"net/http"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = (*licenseSchemaResource)(nil)
	_ resource.ResourceWithConfigure   = (*licenseSchemaResource)(nil)
	_ resource.ResourceWithImportState = (*licenseSchemaResource)(nil)
)

func NewLicenseSchemaResource() resource.Resource {
	return &licenseSchemaResource{}
}

type licenseSchemaResource struct {
	client           *nanoclient.ClientWithResponses
	defaultProductID string
}

type licenseSchemaResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProductID   types.String `tfsdk:"product_id"`
	Description types.String `tfsdk:"description"`
	Fields      types.Set    `tfsdk:"fields"`
}

type licenseFieldModel struct {
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Rules       types.Object `tfsdk:"rules"`
}

type licenseFieldRulesModel struct {
	Min       types.Float64 `tfsdk:"min"`
	Max       types.Float64 `tfsdk:"max"`
	MinLength types.Int64   `tfsdk:"min_length"`
	MaxLength types.Int64   `tfsdk:"max_length"`
	Pattern   types.String  `tfsdk:"pattern"`
	Values    types.List    `tfsdk:"values"`
}

func licenseFieldRulesAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"min":        types.Float64Type,
		"max":        types.Float64Type,
		"min_length": types.Int64Type,
		"max_length": types.Int64Type,
		"pattern":    types.StringType,
		"values":     types.ListType{ElemType: types.StringType},
	}
}

func licenseFieldAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"type":        types.StringType,
		"description": types.StringType,
		"rules":       types.ObjectType{AttrTypes: licenseFieldRulesAttrTypes()},
	}
}

func licenseFieldObjectType() attr.Type {
	return types.ObjectType{AttrTypes: licenseFieldAttrTypes()}
}

func (r *licenseSchemaResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_license_schema"
}

func (r *licenseSchemaResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages the license schema of an Anchor product. A product has at most one schema, " +
			"so this resource is addressed by product. Every field the schema declares must be set by " +
			"every license template of that product.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "License schema KSUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"product_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Product KSUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional schema description.",
			},
			"fields": schema.SetNestedAttribute{
				Required: true,
				Description: "Every license field the schema declares. The set is replaced wholesale on " +
					"update: a field you remove here is removed from the schema.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Stable identifier used by product code, unique within the schema.",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "License field type. One of BOOLEAN, ENUM, LIMIT, NUMBER, STRING.",
							Validators: []validator.String{
								licenseFieldTypeValidator{},
							},
						},
						"description": schema.StringAttribute{
							Optional:    true,
							Description: "Optional field description.",
						},
						"rules": schema.SingleNestedAttribute{
							Optional: true,
							Description: "Validation rules applied when a license value is set. Omit the " +
								"block to declare a field with no rules.",
							Validators: []validator.Object{
								nonEmptyObjectValidator{
									message: "Set at least one rule, or remove the rules block. " +
										"Anchor treats an empty rule set and an absent one as the same thing.",
								},
							},
							Attributes: map[string]schema.Attribute{
								"min": schema.Float64Attribute{
									Optional:    true,
									Description: "Inclusive lower bound. Numeric fields only.",
								},
								"max": schema.Float64Attribute{
									Optional:    true,
									Description: "Inclusive upper bound. Numeric fields only.",
								},
								"min_length": schema.Int64Attribute{
									Optional:    true,
									Description: "Inclusive minimum length in runes. String fields only.",
								},
								"max_length": schema.Int64Attribute{
									Optional:    true,
									Description: "Inclusive maximum length in runes. String fields only.",
								},
								"pattern": schema.StringAttribute{
									Optional:    true,
									Description: "Regular expression the value must match. String fields only.",
								},
								"values": schema.ListAttribute{
									Optional:    true,
									ElementType: types.StringType,
									Description: "The list the value must be drawn from. Enum fields only.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *licenseSchemaResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *providerData, got: %T", req.ProviderData),
		)
		return
	}

	r.client = data.client
	r.defaultProductID = data.productID
}

func (r *licenseSchemaResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan licenseSchemaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	productID, diags := resolveProductID(plan.ProductID, r.defaultProductID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fields, diags := licenseFieldDeclarationsFromPlan(ctx, plan.Fields)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createResp, err := r.client.CreateLicenseSchemaWithResponse(
		ctx,
		productID,
		nanoclient.CreateLicenseSchemaJSONRequestBody{
			Description: plan.Description.ValueStringPointer(),
			Fields:      fields,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create License Schema", err.Error())
		return
	}

	if createResp.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Unable to Create License Schema",
			formatAPIError("create license schema", createResp.StatusCode(), createResp.Body),
		)
		return
	}

	// The written values are kept verbatim from the plan. Anchor validates the declaration
	// and refuses it, or stores it as sent, so only the identifiers come from the response.
	plan.ID = types.StringValue(createResp.JSON201.Id)
	plan.ProductID = types.StringValue(createResp.JSON201.ProductId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *licenseSchemaResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state licenseSchemaResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.client.GetLicenseSchemaWithResponse(ctx, state.ProductID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read License Schema", err.Error())
		return
	}

	if getResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Read License Schema",
			formatAPIError("get license schema", getResp.StatusCode(), getResp.Body),
		)
		return
	}

	updatedState, diags := licenseSchemaStateFromAPI(ctx, getResp.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *licenseSchemaResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan licenseSchemaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	productID, diags := resolveProductID(plan.ProductID, r.defaultProductID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fields, diags := licenseFieldDeclarationsFromPlan(ctx, plan.Fields)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateResp, err := r.client.UpdateLicenseSchemaWithResponse(
		ctx,
		productID,
		nanoclient.UpdateLicenseSchemaJSONRequestBody{
			Description: plan.Description.ValueStringPointer(),
			Fields:      &fields,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update License Schema", err.Error())
		return
	}

	if updateResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Update License Schema",
			formatAPIError("update license schema", updateResp.StatusCode(), updateResp.Body),
		)
		return
	}

	plan.ID = types.StringValue(updateResp.JSON200.Id)
	plan.ProductID = types.StringValue(updateResp.JSON200.ProductId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *licenseSchemaResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state licenseSchemaResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResp, err := r.client.DeleteLicenseSchemaWithResponse(ctx, state.ProductID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Delete License Schema", err.Error())
		return
	}

	if deleteResp.StatusCode() != http.StatusNoContent && deleteResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to Delete License Schema",
			formatAPIError("delete license schema", deleteResp.StatusCode(), deleteResp.Body),
		)
	}
}

// ImportState takes the product KSUID. A product has at most one license schema, so the
// product identifies the schema and the schema's own KSUID is read back from the API.
func (r *licenseSchemaResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			"Expected import identifier with format product_id, got an empty string.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("product_id"), req.ID)...)
}

func licenseFieldDeclarationsFromPlan(
	ctx context.Context,
	fields types.Set,
) ([]nanoclient.LicenseFieldDeclaration, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if fields.IsNull() || fields.IsUnknown() {
		return []nanoclient.LicenseFieldDeclaration{}, diags
	}

	models := make([]licenseFieldModel, 0, len(fields.Elements()))
	diags.Append(fields.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	declarations := make([]nanoclient.LicenseFieldDeclaration, 0, len(models))
	for _, model := range models {
		rules, rulesDiags := licenseFieldRulesFromPlan(ctx, model.Rules)
		diags.Append(rulesDiags...)
		if diags.HasError() {
			return nil, diags
		}

		declarations = append(declarations, nanoclient.LicenseFieldDeclaration{
			Name:        model.Name.ValueString(),
			Type:        nanoclient.LicenseFieldType(model.Type.ValueString()),
			Description: model.Description.ValueStringPointer(),
			Rules:       rules,
		})
	}

	return declarations, diags
}

func licenseFieldRulesFromPlan(
	ctx context.Context,
	rules types.Object,
) (*nanoclient.LicenseFieldRules, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if rules.IsNull() || rules.IsUnknown() {
		return nil, diags
	}

	var model licenseFieldRulesModel
	diags.Append(rules.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	declared := &nanoclient.LicenseFieldRules{
		Min:       model.Min.ValueFloat64Pointer(),
		Max:       model.Max.ValueFloat64Pointer(),
		MinLength: int64PtrToIntPtr(model.MinLength.ValueInt64Pointer()),
		MaxLength: int64PtrToIntPtr(model.MaxLength.ValueInt64Pointer()),
		Pattern:   model.Pattern.ValueStringPointer(),
	}

	if !model.Values.IsNull() && !model.Values.IsUnknown() {
		values := make([]string, 0, len(model.Values.Elements()))
		diags.Append(model.Values.ElementsAs(ctx, &values, false)...)
		if diags.HasError() {
			return nil, diags
		}
		declared.Values = &values
	}

	return declared, diags
}

func licenseSchemaStateFromAPI(
	ctx context.Context,
	licenseSchema *nanoclient.LicenseSchemaResponse,
) (licenseSchemaResourceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	models := make([]licenseFieldModel, 0, len(licenseSchema.Fields))
	for _, field := range licenseSchema.Fields {
		rules, rulesDiags := licenseFieldRulesToState(ctx, field.Rules)
		diags.Append(rulesDiags...)

		models = append(models, licenseFieldModel{
			Name:        types.StringValue(field.Name),
			Type:        types.StringValue(string(field.Type)),
			Description: types.StringPointerValue(field.Description),
			Rules:       rules,
		})
	}

	fields, fieldsDiags := types.SetValueFrom(ctx, licenseFieldObjectType(), models)
	diags.Append(fieldsDiags...)

	return licenseSchemaResourceModel{
		ID:          types.StringValue(licenseSchema.Id),
		ProductID:   types.StringValue(licenseSchema.ProductId),
		Description: types.StringPointerValue(licenseSchema.Description),
		Fields:      fields,
	}, diags
}

// licenseFieldRulesToState maps an API rule set to state. Anchor answers a field with no
// rules as a rule set whose members are all absent, which is the null object in Terraform.
func licenseFieldRulesToState(
	ctx context.Context,
	rules nanoclient.LicenseFieldRules,
) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if rules.Min == nil && rules.Max == nil && rules.MinLength == nil &&
		rules.MaxLength == nil && rules.Pattern == nil && rules.Values == nil {
		return types.ObjectNull(licenseFieldRulesAttrTypes()), diags
	}

	values := types.ListNull(types.StringType)
	if rules.Values != nil {
		list, listDiags := types.ListValueFrom(ctx, types.StringType, *rules.Values)
		diags.Append(listDiags...)
		values = list
	}

	object, objectDiags := types.ObjectValueFrom(ctx, licenseFieldRulesAttrTypes(), licenseFieldRulesModel{
		Min:       types.Float64PointerValue(rules.Min),
		Max:       types.Float64PointerValue(rules.Max),
		MinLength: intPtrToInt64Value(rules.MinLength),
		MaxLength: intPtrToInt64Value(rules.MaxLength),
		Pattern:   types.StringPointerValue(rules.Pattern),
		Values:    values,
	})
	diags.Append(objectDiags...)

	return object, diags
}

func int64PtrToIntPtr(value *int64) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

func intPtrToInt64Value(value *int) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}
