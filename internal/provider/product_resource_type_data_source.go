package provider

import (
	"context"
	"fmt"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*productResourceTypeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*productResourceTypeDataSource)(nil)
)

func NewProductResourceTypeDataSource() datasource.DataSource {
	return &productResourceTypeDataSource{}
}

type productResourceTypeDataSource struct {
	client           *nanoclient.ClientWithResponses
	defaultProductID string
}

type productResourceTypeDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProductID     types.String `tfsdk:"product_id"`
	Slug          types.String `tfsdk:"slug"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	ParentTypes   types.Set    `tfsdk:"parent_types"`
	NestsInItself types.Bool   `tfsdk:"nests_in_itself"`
	BuiltIn       types.Bool   `tfsdk:"built_in"`
}

func (d *productResourceTypeDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_product_resource_type"
}

func (d *productResourceTypeDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Reads one resource type of an Anchor product by slug, such as the built-in organization type " +
			"that every product has, so permissions, roles and other types can reference it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource type KSUID.",
			},
			"product_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Product KSUID. Defaults to the provider product_id.",
			},
			"slug": schema.StringAttribute{
				Required:    true,
				Description: "Slug of the type to read, for example organization.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Display name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description.",
			},
			"parent_types": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Slugs of the types an instance may sit under. Empty for organization.",
			},
			"nests_in_itself": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether an instance may sit under another instance of the same type.",
			},
			"built_in": schema.BoolAttribute{
				Computed:    true,
				Description: "True for the organization type every product has.",
			},
		},
	}
}

func (d *productResourceTypeDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	data, diags := configuredProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() || data == nil {
		return
	}

	d.client = data.client
	d.defaultProductID = data.productID
}

func (d *productResourceTypeDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config productResourceTypeDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	productID, diags := resolveProductID(config.ProductID, d.defaultProductID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := config.Slug.ValueString()
	searchResp, err := d.client.SearchProductResourceTypesWithResponse(
		ctx,
		productID,
		nanoclient.SearchProductResourceTypesJSONRequestBody{
			Filter: &nanoclient.ProductResourceTypeFilter{Slugs: []nanoclient.ResourceTypeSlug{slug}},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Product Resource Type", err.Error())
		return
	}

	if searchResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to Read Product Resource Type",
			formatAPIError("search product resource types", searchResp.StatusCode(), searchResp.Body),
		)
		return
	}

	if len(searchResp.JSON200.Items) != 1 {
		resp.Diagnostics.AddError(
			"Product Resource Type Not Found",
			fmt.Sprintf("Product %s has no resource type with slug %q.", productID, slug),
		)
		return
	}

	found := searchResp.JSON200.Items[0]
	parentTypes, setDiags := types.SetValueFrom(ctx, types.StringType, found.ParentTypes)
	resp.Diagnostics.Append(setDiags...)

	state := productResourceTypeDataSourceModel{
		ID:            types.StringValue(found.Id),
		ProductID:     types.StringValue(found.ProductId),
		Slug:          types.StringValue(found.Slug),
		Name:          types.StringValue(found.Name),
		Description:   stringValueFromAPI(found.Description),
		ParentTypes:   parentTypes,
		NestsInItself: types.BoolValue(found.NestsInItself),
		BuiltIn:       types.BoolValue(found.BuiltIn),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
