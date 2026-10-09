package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nanoclient "github.com/nanostack-dev/anchor/clients/go"
)

func TestProductReadPreservesStateIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(nanoclient.ProductResponse{Id: "response-id", Name: "refreshed"})
	}))
	t.Cleanup(server.Close)
	client, err := nanoclient.NewClientWithResponses(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	r := &productResource{client: client}
	ctx := t.Context()
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	if diags := state.Set(ctx, &productResourceModel{
		ID: types.StringValue("state-id"), Name: types.StringValue("old"), Description: types.StringUnknown(),
	}); diags.HasError() {
		t.Fatal(diags)
	}
	response := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var got productResourceModel
	if diags := response.State.Get(ctx, &got); diags.HasError() {
		t.Fatal(diags)
	}
	if got.ID.ValueString() != "state-id" || got.Name.ValueString() != "refreshed" || !got.Description.IsNull() {
		t.Fatalf("unexpected refreshed state: %+v", got)
	}
}

func TestPermissionReadUsesResponseIdentityAndNullOptionals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(nanoclient.ProductResourcePermissionResponse{
			ProductId: "canonical-product", Name: "flows:read",
		})
	}))
	t.Cleanup(server.Close)
	client, err := nanoclient.NewClientWithResponses(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	r := &productPermissionResource{client: client}
	ctx := t.Context()
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	if diags := state.Set(ctx, &productPermissionResourceModel{
		ID: types.StringValue("old-id"), ProductID: types.StringValue("product"), Name: types.StringValue("flows:read"),
		Description: types.StringUnknown(), ScopeModifier: types.StringUnknown(),
	}); diags.HasError() {
		t.Fatal(diags)
	}
	response := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var got productPermissionResourceModel
	if diags := response.State.Get(ctx, &got); diags.HasError() {
		t.Fatal(diags)
	}
	if got.ID.ValueString() != "canonical-product:flows:read" || got.ProductID.ValueString() != "canonical-product" || !got.Description.IsNull() || !got.ScopeModifier.IsNull() {
		t.Fatalf("unexpected refreshed state: %+v", got)
	}
}
