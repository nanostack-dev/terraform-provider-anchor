package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"
)

func bearerTokenEditor(token string) nanoclient.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

func productAPIKeyEditor(apiKey string) nanoclient.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("X-Product-Api-Key", apiKey)
		return nil
	}
}

func formatAPIError(operation string, statusCode int, body []byte) string {
	bodyText := strings.TrimSpace(string(body))
	if bodyText == "" {
		return fmt.Sprintf("%s failed with status %d", operation, statusCode)
	}

	return fmt.Sprintf("%s failed with status %d: %s", operation, statusCode, bodyText)
}

// apiErrorMessages joins the human-readable half of a decoded ApiErrorResponse, for a
// diagnostic that wants Anchor's own explanation rather than the raw response body.
// Anchor's contract permits several details on one response; this reads all of them.
func apiErrorMessages(errResp *nanoclient.ApiErrorResponse) string {
	if errResp == nil || len(errResp.Errors) == 0 {
		return "no further detail was returned"
	}

	messages := make([]string, 0, len(errResp.Errors))
	for _, apiErr := range errResp.Errors {
		messages = append(messages, apiErr.Message)
	}

	return strings.Join(messages, "; ")
}

func stringPtrFromTFValue(value string) *string {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	return &v
}

// clearableStringPtr is the value an update sends for an optional string. An
// attribute removed from configuration sends an empty string, since Anchor keeps
// the stored value of an omitted field.
func clearableStringPtr(planned, prior types.String) *string {
	if value := stringPtrFromTFValue(planned.ValueString()); value != nil {
		return value
	}
	if prior.IsNull() || prior.ValueString() == "" {
		return nil
	}
	empty := ""
	return &empty
}

// requiresReplaceOnceSet replaces the resource when the value changes, except from
// null: state written before the attribute existed holds null, and Anchor already
// stores the scope it now reports.
func requiresReplaceOnceSet() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.RequiresReplace = !req.StateValue.IsNull()
		},
		"A change replaces the resource, unless the prior state predates the attribute.",
		"A change replaces the resource, unless the prior state predates the attribute.",
	)
}

// stringValueFromAPI keeps an empty optional string null in state, since Anchor returns
// an empty description for one that was never set.
func stringValueFromAPI(value *string) types.String {
	if value == nil || *value == "" {
		return types.StringNull()
	}
	return types.StringValue(*value)
}
