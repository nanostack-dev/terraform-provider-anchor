package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*licenseTemplateResource)(nil)
	_ resource.ResourceWithConfigure   = (*licenseTemplateResource)(nil)
	_ resource.ResourceWithImportState = (*licenseTemplateResource)(nil)
)

func NewLicenseTemplateResource() resource.Resource {
	return &licenseTemplateResource{}
}

type licenseTemplateResource struct {
	client           *nanoclient.ClientWithResponses
	defaultProductID string
}

type licenseTemplateResourceModel struct {
	ID          types.String         `tfsdk:"id"`
	ProductID   types.String         `tfsdk:"product_id"`
	Name        types.String         `tfsdk:"name"`
	Description types.String         `tfsdk:"description"`
	Values      jsontypes.Normalized `tfsdk:"values"`
}

func (r *licenseTemplateResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_license_template"
}

func (r *licenseTemplateResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages an Anchor license template: a named set of values that satisfies the " +
			"product's license schema. Every field the schema declares must be set. " +
			"Destroying this resource archives the template, which cannot be undone — " +
			"Anchor has no delete for a template, because an organization's license names it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "License template KSUID.",
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
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Operator-facing name, unique among the product's active templates.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional template description.",
			},
			"values": schema.StringAttribute{
				Required:   true,
				CustomType: jsontypes.NormalizedType{},
				Description: "A JSON object holding one value per license field the schema declares. " +
					"Write it with jsonencode. The values are replaced wholesale on update, and a " +
					"template that omits any declared field is refused.",
			},
		},
	}
}

func (r *licenseTemplateResource) Configure(
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

func (r *licenseTemplateResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan licenseTemplateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	productID, diags := resolveProductID(plan.ProductID, r.defaultProductID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	values, diags := licenseTemplateValuesFromPlan(plan.Values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createResp, err := r.client.CreateLicenseTemplateWithResponse(
		ctx,
		productID,
		nanoclient.CreateLicenseTemplateJSONRequestBody{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueStringPointer(),
			Values:      values,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create License Template", err.Error())
		return
	}

	if createResp.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Unable to Create License Template",
			formatAPIError("create license template", createResp.StatusCode(), createResp.Body),
		)
		return
	}

	// The written values are kept verbatim from the plan. Anchor validates the template
	// and refuses it, or stores it as sent, so only the identifiers come from the response.
	plan.ID = types.StringValue(createResp.JSON201.Id)
	plan.ProductID = types.StringValue(createResp.JSON201.ProductId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read removes an archived template from state. An archived tier can be neither edited nor
// instantiated, so Terraform treats it as gone and the next plan proposes a replacement.
// Archiving frees the name, so the replacement can keep it.
func (r *licenseTemplateResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state licenseTemplateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.client.GetLicenseTemplateWithResponse(
		ctx,
		state.ProductID.ValueString(),
		state.ID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read License Template", err.Error())
		return
	}

	if getResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Read License Template",
			formatAPIError("get license template", getResp.StatusCode(), getResp.Body),
		)
		return
	}

	if getResp.JSON200.Status == nanoclient.LicenseTemplateStatusARCHIVED {
		resp.State.RemoveResource(ctx)
		return
	}

	updatedState, diags := licenseTemplateStateFromAPI(getResp.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *licenseTemplateResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan licenseTemplateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	productID, diags := resolveProductID(plan.ProductID, r.defaultProductID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	values, diags := licenseTemplateValuesFromPlan(plan.Values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	updateResp, err := r.client.UpdateLicenseTemplateWithResponse(
		ctx,
		productID,
		plan.ID.ValueString(),
		nanoclient.UpdateLicenseTemplateJSONRequestBody{
			Name:        &name,
			Description: plan.Description.ValueStringPointer(),
			Values:      &values,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update License Template", err.Error())
		return
	}

	if updateResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Update License Template",
			formatAPIError("update license template", updateResp.StatusCode(), updateResp.Body),
		)
		return
	}

	plan.ID = types.StringValue(updateResp.JSON200.Id)
	plan.ProductID = types.StringValue(updateResp.JSON200.ProductId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete archives the template. Anchor keeps the row for good, because the organizations
// licensed from it name it as the statement of what they were sold, so there is no delete
// to call. Archiving is the one withdrawal the API offers and it cannot be undone.
func (r *licenseTemplateResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state licenseTemplateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	archiveResp, err := r.client.ArchiveLicenseTemplateWithResponse(
		ctx,
		state.ProductID.ValueString(),
		state.ID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Archive License Template", err.Error())
		return
	}

	if archiveResp.StatusCode() != http.StatusOK && archiveResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to Archive License Template",
			formatAPIError("archive license template", archiveResp.StatusCode(), archiveResp.Body),
		)
	}
}

func (r *licenseTemplateResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format product_id:license_template_id, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("product_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func licenseTemplateValuesFromPlan(
	values jsontypes.Normalized,
) (nanoclient.LicenseTemplateValues, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	decoded := nanoclient.LicenseTemplateValues{}
	diags.Append(values.Unmarshal(&decoded)...)
	if diags.HasError() {
		return nil, diags
	}

	return decoded, diags
}

func licenseTemplateStateFromAPI(
	template *nanoclient.LicenseTemplateResponse,
) (licenseTemplateResourceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	encoded, err := json.Marshal(template.Values)
	if err != nil {
		diags.AddError("Unable to Encode License Template Values", err.Error())
		return licenseTemplateResourceModel{}, diags
	}

	return licenseTemplateResourceModel{
		ID:          types.StringValue(template.Id),
		ProductID:   types.StringValue(template.ProductId),
		Name:        types.StringValue(template.Name),
		Description: types.StringPointerValue(template.Description),
		Values:      jsontypes.NewNormalizedValue(string(encoded)),
	}, diags
}
