package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	nanoclient "github.com/nanostack-dev/anchor/clients/go"
)

func TestResourceConfiguration(t *testing.T) {
	client := &nanoclient.ClientWithResponses{}
	data := &providerData{client: client, productID: "product"}
	for _, newResource := range []func() resource.Resource{
		NewProductResource, NewProductRoleResource, NewProductPermissionResource,
		NewLicenseSchemaResource, NewLicenseTemplateResource,
	} {
		r := newResource().(resource.ResourceWithConfigure)
		for _, value := range []any{nil, (*providerData)(nil), "wrong", data} {
			var response resource.ConfigureResponse
			r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: value}, &response)
			if value == "wrong" {
				if len(response.Diagnostics) != 1 || response.Diagnostics[0].Summary() != "Unexpected Provider Data Type" || response.Diagnostics[0].Detail() != "Expected *providerData, got: string" {
					t.Fatalf("%T: unexpected diagnostics: %v", r, response.Diagnostics)
				}
				continue
			}
			if response.Diagnostics.HasError() {
				t.Fatalf("%T: configure: %v", r, response.Diagnostics)
			}
		}
		switch configured := r.(type) {
		case *productResource:
			if configured.client != client {
				t.Fatal("product client not configured")
			}
		case *productRoleResource:
			if configured.client != client || configured.defaultProductID != "product" {
				t.Fatal("role defaults not configured")
			}
		case *productPermissionResource:
			if configured.client != client || configured.defaultProductID != "product" {
				t.Fatal("permission defaults not configured")
			}
		case *licenseSchemaResource:
			if configured.client != client || configured.defaultProductID != "product" {
				t.Fatal("schema defaults not configured")
			}
		case *licenseTemplateResource:
			if configured.client != client || configured.defaultProductID != "product" {
				t.Fatal("template defaults not configured")
			}
		}
	}
}
