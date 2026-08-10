package provider_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"
	"github.com/nanostack-dev/terraform-provider-anchor/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories serves the provider in-process to the acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"anchor": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck stops an acceptance test that has no Anchor instance to run against.
//
// The licensing tests create their own product, which needs a platform bearer token.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("ANCHOR_TOKEN") == "" {
		t.Fatal("ANCHOR_TOKEN must be set for acceptance tests")
	}
}

// testAccClient builds an API client from the same environment the provider reads, so a
// test can change a resource behind Terraform's back and prove the next plan reports it.
func testAccClient(t *testing.T) *nanoclient.ClientWithResponses {
	t.Helper()

	baseURL := os.Getenv("ANCHOR_BASE_URL")
	if baseURL == "" {
		baseURL = "https://anchorapi.nanostack.dev"
	}

	editors := make([]nanoclient.RequestEditorFn, 0, 2)
	if token := os.Getenv("ANCHOR_TOKEN"); token != "" {
		editors = append(editors, func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+token)
			return nil
		})
	}
	if apiKey := os.Getenv("ANCHOR_API_KEY"); apiKey != "" {
		editors = append(editors, func(_ context.Context, req *http.Request) error {
			req.Header.Set("X-Product-Api-Key", apiKey)
			return nil
		})
	}

	client, err := nanoclient.NewClientWithConfig(nanoclient.Config{
		BaseURL:        baseURL,
		RequestEditors: editors,
	})
	if err != nil {
		t.Fatalf("build acceptance test client: %v", err)
	}

	return client
}

func testAccContext(t *testing.T) context.Context {
	t.Helper()
	return t.Context()
}

// testAccProductAPIKeyClient mints a fresh product API key scoped to permissions and
// returns a client authenticated with it.
//
// Creating an organization and instantiating a license both require a product API key —
// platform bearer auth answers 401, not merely 403, for either route — so a test whose
// subject is one of them cannot go through testAccClient the way every other helper does.
func testAccProductAPIKeyClient(
	t *testing.T, productID string, permissions []string,
) *nanoclient.ClientWithResponses {
	t.Helper()

	keyResp, err := testAccClient(t).CreateProductAPIKeyWithResponse(
		testAccContext(t), productID,
		nanoclient.CreateProductAPIKeyJSONRequestBody{
			Name:        acctest.RandomWithPrefix("tfacc-key"),
			Permissions: permissions,
		},
	)
	if err != nil {
		t.Fatalf("create product API key: %v", err)
	}
	if keyResp.JSON201 == nil {
		t.Fatalf("create product API key: status %d: %s", keyResp.StatusCode(), keyResp.Body)
	}

	baseURL := os.Getenv("ANCHOR_BASE_URL")
	if baseURL == "" {
		baseURL = "https://anchorapi.nanostack.dev"
	}

	client, err := nanoclient.NewClientWithConfig(nanoclient.Config{
		BaseURL: baseURL,
		RequestEditors: []nanoclient.RequestEditorFn{
			func(_ context.Context, req *http.Request) error {
				req.Header.Set("X-Product-Api-Key", keyResp.JSON201.Value)
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("build product API key client: %v", err)
	}

	return client
}

// testAccCaptureAttr copies an attribute out of the Terraform state into target, so a
// later step can reach the live resource over the API and change it behind Terraform.
func testAccCaptureAttr(resourceName, attribute string, target *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", resourceName)
		}

		value, ok := res.Primary.Attributes[attribute]
		if !ok {
			return fmt.Errorf("%s has no attribute %q", resourceName, attribute)
		}

		*target = value

		return nil
	}
}
