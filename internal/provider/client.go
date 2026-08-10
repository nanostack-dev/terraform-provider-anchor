package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
