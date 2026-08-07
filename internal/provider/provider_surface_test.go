package provider_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	anchorprovider "github.com/nanostack-dev/terraform-provider-anchor/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// TestProviderOffersNoOrganizationLicenseSurface guards ADR-0006 in the anchor repository.
//
// An organization's license is runtime data. It carries per-customer adjustments, so a
// Terraform resource or data source over it would revert every one of them on the next
// apply. This test fails if such a surface is ever added.
func TestProviderOffersNoOrganizationLicenseSurface(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	anchor := anchorprovider.New("test")()

	for _, name := range resourceTypeNames(ctx, t, anchor) {
		if strings.Contains(name, "organization_license") || strings.Contains(name, "license_usage") {
			t.Errorf("resource %q manages an organization license, which Terraform must not own", name)
		}
	}

	for _, name := range dataSourceTypeNames(ctx, t, anchor) {
		if strings.Contains(name, "organization_license") || strings.Contains(name, "license_usage") {
			t.Errorf("data source %q reads an organization license, which Terraform must not own", name)
		}
	}
}

func TestProviderRegistersLicensingResources(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	names := resourceTypeNames(ctx, t, anchorprovider.New("test")())

	for _, want := range []string{"anchor_license_schema", "anchor_license_template"} {
		if !slices.Contains(names, want) {
			t.Errorf("resource %q is not registered, got: %v", want, names)
		}
	}
}

func resourceTypeNames(ctx context.Context, t *testing.T, anchor provider.Provider) []string {
	t.Helper()

	names := make([]string, 0)
	for _, newResource := range anchor.Resources(ctx) {
		resp := &resource.MetadataResponse{}
		newResource().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "anchor"}, resp)
		names = append(names, resp.TypeName)
	}

	return names
}

func dataSourceTypeNames(ctx context.Context, t *testing.T, anchor provider.Provider) []string {
	t.Helper()

	names := make([]string, 0)
	for _, newDataSource := range anchor.DataSources(ctx) {
		resp := &datasource.MetadataResponse{}
		newDataSource().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "anchor"}, resp)
		names = append(names, resp.TypeName)
	}

	return names
}
