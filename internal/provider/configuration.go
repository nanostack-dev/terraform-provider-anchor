package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func configuredProviderData(value any) (*providerData, diag.Diagnostics) {
	if value == nil {
		return nil, nil
	}
	data, ok := value.(*providerData)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Unexpected Provider Data Type", fmt.Sprintf("Expected *providerData, got: %T", value))
		return nil, diags
	}
	return data, nil
}
