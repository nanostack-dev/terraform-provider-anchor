package provider

import (
	"context"
	"fmt"
	"strings"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.String = (*licenseFieldTypeValidator)(nil)
	_ validator.Object = (*nonEmptyObjectValidator)(nil)
)

// licenseFieldTypeNames lists the license field types the Anchor contract declares.
func licenseFieldTypeNames() []string {
	return []string{
		string(nanoclient.LicenseFieldTypeBOOLEAN),
		string(nanoclient.LicenseFieldTypeENUM),
		string(nanoclient.LicenseFieldTypeLIMIT),
		string(nanoclient.LicenseFieldTypeNUMBER),
		string(nanoclient.LicenseFieldTypeSTRING),
	}
}

type licenseFieldTypeValidator struct{}

func (v licenseFieldTypeValidator) Description(_ context.Context) string {
	return "value must be one of: " + strings.Join(licenseFieldTypeNames(), ", ")
}

func (v licenseFieldTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v licenseFieldTypeValidator) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if nanoclient.LicenseFieldType(req.ConfigValue.ValueString()).Valid() {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid License Field Type",
		fmt.Sprintf(
			"Expected one of %s, got: %q.",
			strings.Join(licenseFieldTypeNames(), ", "),
			req.ConfigValue.ValueString(),
		),
	)
}

// nonEmptyObjectValidator refuses an object whose every attribute is null.
//
// An empty rules object and an absent one mean the same thing to Anchor, which returns
// neither. Refusing the empty form keeps the configuration and the API response in step,
// so a plan does not report drift that no write can settle.
type nonEmptyObjectValidator struct {
	message string
}

func (v nonEmptyObjectValidator) Description(_ context.Context) string {
	return "object must set at least one attribute"
}

func (v nonEmptyObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v nonEmptyObjectValidator) ValidateObject(
	_ context.Context,
	req validator.ObjectRequest,
	resp *validator.ObjectResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	for _, attribute := range req.ConfigValue.Attributes() {
		if !attribute.IsNull() {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(req.Path, "Empty Object", v.message)
}
