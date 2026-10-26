package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*productResourceTypeResource)(nil)
	_ resource.ResourceWithConfigure   = (*productResourceTypeResource)(nil)
	_ resource.ResourceWithImportState = (*productResourceTypeResource)(nil)
)

func NewProductResourceTypeResource() resource.Resource {
	return &productResourceTypeResource{}
}

type productResourceTypeResource struct {
	client           *nanoclient.ClientWithResponses
	defaultProductID string
}

type productResourceTypeResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProductID     types.String `tfsdk:"product_id"`
	Slug          types.String `tfsdk:"slug"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	ParentTypes   types.Set    `tfsdk:"parent_types"`
	NestsInItself types.Bool   `tfsdk:"nests_in_itself"`
}

func (r *productResourceTypeResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_product_resource_type"
}

func (r *productResourceTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a resource type of an Anchor product: a kind of thing the product protects, " +
			"such as a project or a folder, placed under other types in the product's tree.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource type KSUID.",
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
			"slug": schema.StringAttribute{
				Required: true,
				Description: "Lowercase identifier, for example project. It never changes, so a new value " +
					"replaces the resource type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description.",
			},
			"parent_types": schema.SetAttribute{
				Required: true,
				Description: "Slugs of the types an instance may sit under, for example organization. " +
					"Reference another resource type's slug so Terraform creates it first.",
				ElementType: types.StringType,
			},
			"nests_in_itself": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether an instance may sit under another instance of the same type.",
			},
		},
	}
}

func (r *productResourceTypeResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	data, diags := configuredProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() || data == nil {
		return
	}

	r.client = data.client
	r.defaultProductID = data.productID
}

func (r *productResourceTypeResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan productResourceTypeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	productID, diags := resolveProductID(plan.ProductID, r.defaultProductID)
	resp.Diagnostics.Append(diags...)
	parentTypes, setDiags := setToStringSlice(ctx, plan.ParentTypes)
	resp.Diagnostics.Append(setDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nestsInItself := plan.NestsInItself.ValueBool()
	createResp, err := r.client.CreateProductResourceTypeWithResponse(
		ctx,
		productID,
		nanoclient.CreateProductResourceTypeJSONRequestBody{
			Slug:          plan.Slug.ValueString(),
			Name:          plan.Name.ValueString(),
			Description:   stringPtrFromTFValue(plan.Description.ValueString()),
			ParentTypes:   parentTypes,
			NestsInItself: &nestsInItself,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Product Resource Type", err.Error())
		return
	}

	if createResp.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Unable to Create Product Resource Type",
			formatAPIError("create product resource type", createResp.StatusCode(), createResp.Body),
		)
		return
	}

	state, diags := productResourceTypeStateFromAPI(ctx, createResp.JSON201)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *productResourceTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state productResourceTypeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.client.GetProductResourceTypeWithResponse(
		ctx,
		state.ProductID.ValueString(),
		state.ID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Product Resource Type", err.Error())
		return
	}

	if getResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Read Product Resource Type",
			formatAPIError("get product resource type", getResp.StatusCode(), getResp.Body),
		)
		return
	}

	if getResp.JSON200.BuiltIn {
		resp.Diagnostics.AddError(
			"Built-In Resource Type",
			"The organization resource type is built into every product and can't be managed. "+
				"Read it with the anchor_product_resource_type data source instead.",
		)
		return
	}

	state, diags := productResourceTypeStateFromAPI(ctx, getResp.JSON200)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *productResourceTypeResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, prior productResourceTypeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentTypes, diags := setToStringSlice(ctx, plan.ParentTypes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	nestsInItself := plan.NestsInItself.ValueBool()
	updateResp, err := r.client.UpdateProductResourceTypeWithResponse(
		ctx,
		prior.ProductID.ValueString(),
		prior.ID.ValueString(),
		nanoclient.UpdateProductResourceTypeJSONRequestBody{
			Name:          &name,
			Description:   clearableStringPtr(plan.Description, prior.Description),
			ParentTypes:   &parentTypes,
			NestsInItself: &nestsInItself,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Product Resource Type", err.Error())
		return
	}

	if updateResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Update Product Resource Type",
			formatAPIError("update product resource type", updateResp.StatusCode(), updateResp.Body),
		)
		return
	}

	state, diags := productResourceTypeStateFromAPI(ctx, updateResp.JSON200)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *productResourceTypeResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state productResourceTypeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResp, err := r.client.DeleteProductResourceTypeWithResponse(
		ctx,
		state.ProductID.ValueString(),
		state.ID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Delete Product Resource Type", err.Error())
		return
	}

	if deleteResp.StatusCode() != http.StatusNoContent && deleteResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to Delete Product Resource Type",
			formatAPIError("delete product resource type", deleteResp.StatusCode(), deleteResp.Body),
		)
	}
}

func (r *productResourceTypeResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	productID, resourceTypeID, found := strings.Cut(req.ID, ":")
	if !found || productID == "" || resourceTypeID == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format product_id:resource_type_id, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("product_id"), productID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), resourceTypeID)...)
}

func productResourceTypeStateFromAPI(
	ctx context.Context,
	resourceType *nanoclient.ProductResourceTypeResponse,
) (productResourceTypeResourceModel, diag.Diagnostics) {
	parentTypes, diags := types.SetValueFrom(ctx, types.StringType, resourceType.ParentTypes)

	return productResourceTypeResourceModel{
		ID:            types.StringValue(resourceType.Id),
		ProductID:     types.StringValue(resourceType.ProductId),
		Slug:          types.StringValue(resourceType.Slug),
		Name:          types.StringValue(resourceType.Name),
		Description:   stringValueFromAPI(resourceType.Description),
		ParentTypes:   parentTypes,
		NestsInItself: types.BoolValue(resourceType.NestsInItself),
	}, diags
}
