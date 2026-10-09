package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"
)

func TestLicenseTemplateArchiveResponsePolicy(t *testing.T) {
	for _, tt := range []struct {
		name       string
		status     int
		body       string
		cancel     bool
		wantDetail string
	}{
		{name: "archived", status: 200, body: `{"id":"template","product_id":"product","status":"ARCHIVED"}`},
		{name: "API failure", status: 400, body: `{"error":"not allowed"}`, wantDetail: "archive license template failed with status 400"},
		{name: "invalid response", status: 200, body: `{`, wantDetail: "unexpected end of JSON input"},
		{name: "canceled request", cancel: true, wantDetail: "context canceled"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/v1/products/product/licensing/templates/template/archive" {
					t.Errorf("unexpected archive path %q", req.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(server.Close)
			client, err := nanoclient.NewClientWithResponses(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if tt.cancel {
				cancel()
			}
			r := &licenseTemplateResource{client: client}
			result, diags := r.archive(ctx, "product", "template")
			if tt.wantDetail != "" {
				if result != nil || len(diags) != 1 || diags[0].Summary() != "Unable to Archive License Template" || !strings.Contains(diags[0].Detail(), tt.wantDetail) {
					t.Fatalf("result=%v diagnostics=%v", result, diags)
				}
				return
			}
			if diags.HasError() || result == nil || result.Id != "template" || result.Status != nanoclient.LicenseTemplateStatusARCHIVED {
				t.Fatalf("result=%v diagnostics=%v", result, diags)
			}
		})
	}
}
