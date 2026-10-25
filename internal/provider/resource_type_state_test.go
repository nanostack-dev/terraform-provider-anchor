package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	nanoclient "github.com/nanostack-dev/anchor/clients/go"
)

func TestResourceTypeStateKeepsAnEmptyDescriptionNull(t *testing.T) {
	empty := ""
	state, diags := productResourceTypeStateFromAPI(t.Context(), &nanoclient.ProductResourceTypeResponse{
		Id: "rtype_1", ProductId: "prd_1", Slug: "folder", Name: "Folder", Description: &empty,
		ParentTypes: []string{"project"}, NestsInItself: true,
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !state.Description.IsNull() || state.Slug.ValueString() != "folder" || !state.NestsInItself.ValueBool() {
		t.Fatalf("unexpected state: %+v", state)
	}
	if len(state.ParentTypes.Elements()) != 1 {
		t.Fatalf("parent types not kept: %v", state.ParentTypes)
	}
}

func TestClearableStringPtrSendsEmptyToClearARemovedValue(t *testing.T) {
	cleared := clearableStringPtr(types.StringNull(), types.StringValue("old"))
	if cleared == nil || *cleared != "" {
		t.Fatalf("a removed description must send an empty string, got %v", cleared)
	}
	if kept := clearableStringPtr(types.StringNull(), types.StringNull()); kept != nil {
		t.Fatalf("a description never set must stay omitted, got %q", *kept)
	}
	if set := clearableStringPtr(types.StringValue("new"), types.StringValue("old")); set == nil || *set != "new" {
		t.Fatalf("a planned description must be sent, got %v", set)
	}
}

func TestResourceTypeChangeFromNullDoesNotReplace(t *testing.T) {
	cases := []struct {
		name    string
		prior   types.String
		replace bool
	}{
		{"state written before resource_type existed", types.StringNull(), false},
		{"a new scope", types.StringValue("organization"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			existing := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{})
			resp := &planmodifier.StringResponse{PlanValue: types.StringValue("project")}
			requiresReplaceOnceSet().PlanModifyString(t.Context(), planmodifier.StringRequest{
				State:      tfsdk.State{Raw: existing},
				Plan:       tfsdk.Plan{Raw: existing},
				StateValue: tc.prior, PlanValue: types.StringValue("project"), ConfigValue: types.StringValue("project"),
			}, resp)
			if resp.RequiresReplace != tc.replace {
				t.Fatalf("requires replace = %v, want %v", resp.RequiresReplace, tc.replace)
			}
		})
	}
}
