package provider_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccLicenseSchema covers the whole life of a license schema: create it, update its
// declaration, report drift after an edit made outside Terraform, correct that drift, and
// delete it.
func TestAccLicenseSchema(t *testing.T) {
	productName := acctest.RandomWithPrefix("tfacc-schema")

	var productID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckLicenseSchemaDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseSchemaConfig(productName, "Flows an organization can hold.", 100000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureAttr("anchor_license_schema.test", "product_id", &productID),
					resource.TestCheckResourceAttrSet("anchor_license_schema.test", "id"),
					resource.TestCheckResourceAttrPair(
						"anchor_license_schema.test", "product_id",
						"anchor_product.test", "id",
					),
					resource.TestCheckResourceAttr(
						"anchor_license_schema.test", "description", "Acceptance test schema.",
					),
					resource.TestCheckResourceAttr("anchor_license_schema.test", "fields.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"anchor_license_schema.test", "fields.*",
						map[string]string{
							"name":        "max_flows",
							"type":        "LIMIT",
							"usage_shape": "GAUGE",
							"description": "Flows an organization can hold.",
							"rules.min":   "0",
							"rules.max":   "100000",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"anchor_license_schema.test", "fields.*",
						map[string]string{
							"name": "sso_enabled",
							"type": "BOOLEAN",
						},
					),
				),
			},
			{
				ResourceName:      "anchor_license_schema.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLicenseSchemaImportID,
			},
			// Update: the declaration changes and the fields are replaced wholesale.
			{
				Config: testAccLicenseSchemaConfig(productName, "Flows this organization can hold.", 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"anchor_license_schema.test", "fields.*",
						map[string]string{
							"name":        "max_flows",
							"description": "Flows this organization can hold.",
							"rules.max":   "500",
						},
					),
				),
			},
			// Drift: the schema is edited outside Terraform, so the next plan is not empty.
			{
				PreConfig:          testAccDriftLicenseSchema(t, &productID),
				Config:             testAccLicenseSchemaConfig(productName, "Flows this organization can hold.", 500),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// The drift is corrected by the next apply.
			{
				Config: testAccLicenseSchemaConfig(productName, "Flows this organization can hold.", 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"anchor_license_schema.test", "description", "Acceptance test schema.",
					),
					resource.TestCheckResourceAttr("anchor_license_schema.test", "fields.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"anchor_license_schema.test", "fields.*",
						map[string]string{
							"name":      "max_flows",
							"rules.max": "500",
						},
					),
				),
			},
		},
	})
}

func testAccLicenseSchemaConfig(productName, flowsDescription string, maxFlows int) string {
	return fmt.Sprintf(`
resource "anchor_product" "test" {
  name        = %[1]q
  description = "Terraform provider acceptance test."
}

resource "anchor_license_schema" "test" {
  product_id  = anchor_product.test.id
  description = "Acceptance test schema."

  fields = [
    {
      name        = "max_flows"
      type        = "LIMIT"
      usage_shape = "GAUGE"
      description = %[2]q
      rules = {
        min = 0
        max = %[3]d
      }
    },
    {
      name        = "sso_enabled"
      type        = "BOOLEAN"
      description = "Whether single sign-on is granted."
    },
  ]
}
`, productName, flowsDescription, maxFlows)
}

func testAccLicenseSchemaImportID(state *terraform.State) (string, error) {
	res, ok := state.RootModule().Resources["anchor_license_schema.test"]
	if !ok {
		return "", errors.New("anchor_license_schema.test not found in state")
	}

	return res.Primary.Attributes["product_id"], nil
}

// testAccDriftLicenseSchema edits the schema behind Terraform's back: it rewrites the
// description and drops a declared field.
func testAccDriftLicenseSchema(t *testing.T, productID *string) func() {
	return func() {
		t.Helper()

		if *productID == "" {
			t.Fatal("no product ID captured from the Terraform state")
		}

		client := testAccClient(t)
		description := "Edited outside Terraform."
		shape := nanoclient.GAUGE
		fields := []nanoclient.LicenseFieldDeclaration{
			{
				Name:       "max_flows",
				Type:       nanoclient.LicenseFieldTypeLIMIT,
				UsageShape: &shape,
			},
		}

		updateResp, err := client.UpdateLicenseSchemaWithResponse(
			testAccContext(t),
			*productID,
			nanoclient.UpdateLicenseSchemaJSONRequestBody{
				Description: &description,
				Fields:      &fields,
			},
		)
		if err != nil {
			t.Fatalf("drift the license schema: %v", err)
		}
		if updateResp.JSON200 == nil {
			t.Fatalf("drift the license schema: status %d: %s", updateResp.StatusCode(), updateResp.Body)
		}
	}
}

func testAccCheckLicenseSchemaDestroyed(t *testing.T) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccClient(t)

		for name, res := range state.RootModule().Resources {
			if res.Type != "anchor_license_schema" {
				continue
			}

			getResp, err := client.GetLicenseSchemaWithResponse(
				testAccContext(t),
				res.Primary.Attributes["product_id"],
			)
			if err != nil {
				return fmt.Errorf("check %s was destroyed: %w", name, err)
			}
			if getResp.StatusCode() != http.StatusNotFound {
				return fmt.Errorf("%s still exists: status %d", name, getResp.StatusCode())
			}
		}

		return nil
	}
}
