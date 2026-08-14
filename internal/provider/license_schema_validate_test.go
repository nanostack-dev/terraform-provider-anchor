package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestLicenseSchemaValidateConfig(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	r := &licenseSchemaResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema: %s", schemaResp.Diagnostics)
	}

	tests := []struct {
		name      string
		config    tftypes.Value
		wantError string
	}{
		{
			name: "a limit with a gauge shape is accepted",
			config: licenseSchemaConfigValue(t, []tftypes.Value{
				licenseFieldValue(t, "max_flows", "LIMIT", "GAUGE"),
			}),
		},
		{
			name: "a boolean with no usage shape is accepted",
			config: licenseSchemaConfigValue(t, []tftypes.Value{
				licenseFieldValue(t, "sso_enabled", "BOOLEAN", ""),
			}),
		},
		{
			name: "a limit with no usage shape is refused",
			config: licenseSchemaConfigValue(t, []tftypes.Value{
				licenseFieldValue(t, "max_flows", "LIMIT", ""),
			}),
			wantError: "Usage Shape Required",
		},
		{
			name: "a usage shape on a boolean is refused",
			config: licenseSchemaConfigValue(t, []tftypes.Value{
				licenseFieldValue(t, "sso_enabled", "BOOLEAN", "GAUGE"),
			}),
			wantError: "Usage Shape Not Allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := resource.ValidateConfigRequest{
				Config: tfsdk.Config{
					Raw:    tt.config,
					Schema: schemaResp.Schema,
				},
			}
			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(ctx, req, resp)

			if tt.wantError == "" {
				if resp.Diagnostics.HasError() {
					t.Fatalf("unexpected error: %s", resp.Diagnostics)
				}
				return
			}

			if !resp.Diagnostics.HasError() {
				t.Fatalf("expected error %q, got none", tt.wantError)
			}
			found := false
			for _, d := range resp.Diagnostics.Errors() {
				if d.Summary() == tt.wantError {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected summary %q, got: %s", tt.wantError, resp.Diagnostics)
			}
		})
	}
}

func licenseSchemaConfigValue(t *testing.T, fields []tftypes.Value) tftypes.Value {
	t.Helper()

	return tftypes.NewValue(licenseSchemaObjectType(), map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"product_id":  tftypes.NewValue(tftypes.String, "prd_test"),
		"description": tftypes.NewValue(tftypes.String, nil),
		"fields":      tftypes.NewValue(tftypes.Set{ElementType: licenseFieldObjectTfType()}, fields),
	})
}

func licenseFieldValue(t *testing.T, name, fieldType, usageShape string) tftypes.Value {
	t.Helper()

	var shape tftypes.Value
	if usageShape == "" {
		shape = tftypes.NewValue(tftypes.String, nil)
	} else {
		shape = tftypes.NewValue(tftypes.String, usageShape)
	}

	return tftypes.NewValue(licenseFieldObjectTfType(), map[string]tftypes.Value{
		"name":        tftypes.NewValue(tftypes.String, name),
		"type":        tftypes.NewValue(tftypes.String, fieldType),
		"description": tftypes.NewValue(tftypes.String, nil),
		"usage_shape": shape,
		"rules": tftypes.NewValue(tftypes.Object{
			AttributeTypes: licenseFieldRulesTfTypes(),
		}, nil),
	})
}

func licenseSchemaObjectType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":          tftypes.String,
			"product_id":  tftypes.String,
			"description": tftypes.String,
			"fields":      tftypes.Set{ElementType: licenseFieldObjectTfType()},
		},
	}
}

func licenseFieldObjectTfType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name":        tftypes.String,
			"type":        tftypes.String,
			"description": tftypes.String,
			"usage_shape": tftypes.String,
			"rules": tftypes.Object{
				AttributeTypes: licenseFieldRulesTfTypes(),
			},
		},
	}
}

func licenseFieldRulesTfTypes() map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"min":        tftypes.Number,
		"max":        tftypes.Number,
		"min_length": tftypes.Number,
		"max_length": tftypes.Number,
		"pattern":    tftypes.String,
		"values":     tftypes.List{ElementType: tftypes.String},
	}
}
