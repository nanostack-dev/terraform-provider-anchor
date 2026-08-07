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

// TestAccLicenseTemplate covers the whole life of a license template: create it, edit its
// values, report drift after an edit made outside Terraform, correct that drift, and
// destroy it. Destroying a template archives it, because Anchor has no delete for one.
func TestAccLicenseTemplate(t *testing.T) {
	productName := acctest.RandomWithPrefix("tfacc-template")
	templateName := acctest.RandomWithPrefix("pro")

	var productID string
	var templateID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckLicenseTemplateArchived(t),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseTemplateConfig(productName, templateName, 100, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureAttr("anchor_license_template.test", "product_id", &productID),
					testAccCaptureAttr("anchor_license_template.test", "id", &templateID),
					resource.TestCheckResourceAttrSet("anchor_license_template.test", "id"),
					resource.TestCheckResourceAttrPair(
						"anchor_license_template.test", "product_id",
						"anchor_product.test", "id",
					),
					resource.TestCheckResourceAttr("anchor_license_template.test", "name", templateName),
					resource.TestCheckResourceAttr(
						"anchor_license_template.test", "description", "Acceptance test tier.",
					),
					resource.TestCheckResourceAttr(
						"anchor_license_template.test", "values",
						`{"max_flows":100,"sso_enabled":true}`,
					),
				),
			},
			{
				ResourceName:      "anchor_license_template.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLicenseTemplateImportID,
			},
			// Update: the values are replaced wholesale.
			{
				Config: testAccLicenseTemplateConfig(productName, templateName, 500, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"anchor_license_template.test", "values",
						`{"max_flows":500,"sso_enabled":false}`,
					),
				),
			},
			// Drift: the template is edited outside Terraform, so the next plan is not empty.
			{
				PreConfig:          testAccDriftLicenseTemplate(t, &productID, &templateID),
				Config:             testAccLicenseTemplateConfig(productName, templateName, 500, false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// The drift is corrected by the next apply.
			{
				Config: testAccLicenseTemplateConfig(productName, templateName, 500, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"anchor_license_template.test", "description", "Acceptance test tier.",
					),
					resource.TestCheckResourceAttr(
						"anchor_license_template.test", "values",
						`{"max_flows":500,"sso_enabled":false}`,
					),
				),
			},
		},
	})
}

// TestAccLicenseTemplateArchivedOutsideTerraform proves the recovery path of ADR-0010.
// An archived template can be neither edited nor instantiated, so Terraform treats it as
// gone and plans a replacement. Archiving frees the name, so the replacement keeps it.
func TestAccLicenseTemplateArchivedOutsideTerraform(t *testing.T) {
	productName := acctest.RandomWithPrefix("tfacc-archived")
	templateName := acctest.RandomWithPrefix("pro")

	var productID string
	var templateID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckLicenseTemplateArchived(t),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseTemplateConfig(productName, templateName, 100, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureAttr("anchor_license_template.test", "product_id", &productID),
					testAccCaptureAttr("anchor_license_template.test", "id", &templateID),
				),
			},
			{
				PreConfig:          testAccArchiveLicenseTemplate(t, &productID, &templateID),
				Config:             testAccLicenseTemplateConfig(productName, templateName, 100, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccLicenseTemplateConfig(productName, templateName, 100, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("anchor_license_template.test", "name", templateName),
					resource.TestCheckResourceAttr(
						"anchor_license_template.test", "values",
						`{"max_flows":100,"sso_enabled":true}`,
					),
				),
			},
		},
	})
}

func testAccLicenseTemplateConfig(productName, templateName string, maxFlows int, ssoEnabled bool) string {
	return fmt.Sprintf(`
resource "anchor_product" "test" {
  name        = %[1]q
  description = "Terraform provider acceptance test."
}

resource "anchor_license_schema" "test" {
  product_id = anchor_product.test.id

  fields = [
    {
      name = "max_flows"
      type = "LIMIT"
      rules = {
        min = 0
        max = 100000
      }
    },
    {
      name = "sso_enabled"
      type = "BOOLEAN"
    },
  ]
}

resource "anchor_license_template" "test" {
  product_id  = anchor_product.test.id
  name        = %[2]q
  description = "Acceptance test tier."

  values = jsonencode({
    max_flows   = %[3]d
    sso_enabled = %[4]t
  })

  depends_on = [anchor_license_schema.test]
}
`, productName, templateName, maxFlows, ssoEnabled)
}

func testAccLicenseTemplateImportID(state *terraform.State) (string, error) {
	res, ok := state.RootModule().Resources["anchor_license_template.test"]
	if !ok {
		return "", errors.New("anchor_license_template.test not found in state")
	}

	return res.Primary.Attributes["product_id"] + ":" + res.Primary.ID, nil
}

// testAccDriftLicenseTemplate edits the template behind Terraform's back.
func testAccDriftLicenseTemplate(t *testing.T, productID, templateID *string) func() {
	return func() {
		t.Helper()

		if *productID == "" || *templateID == "" {
			t.Fatal("no template captured from the Terraform state")
		}

		client := testAccClient(t)
		description := "Edited outside Terraform."
		values := nanoclient.LicenseTemplateValues{
			"max_flows":   float64(9000),
			"sso_enabled": true,
		}

		updateResp, err := client.UpdateLicenseTemplateWithResponse(
			testAccContext(t),
			*productID,
			*templateID,
			nanoclient.UpdateLicenseTemplateJSONRequestBody{
				Description: &description,
				Values:      &values,
			},
		)
		if err != nil {
			t.Fatalf("drift the license template: %v", err)
		}
		if updateResp.JSON200 == nil {
			t.Fatalf("drift the license template: status %d: %s", updateResp.StatusCode(), updateResp.Body)
		}
	}
}

// testAccArchiveLicenseTemplate withdraws the template behind Terraform's back.
func testAccArchiveLicenseTemplate(t *testing.T, productID, templateID *string) func() {
	return func() {
		t.Helper()

		if *productID == "" || *templateID == "" {
			t.Fatal("no template captured from the Terraform state")
		}

		client := testAccClient(t)

		archiveResp, err := client.ArchiveLicenseTemplateWithResponse(
			testAccContext(t),
			*productID,
			*templateID,
		)
		if err != nil {
			t.Fatalf("archive the license template: %v", err)
		}
		if archiveResp.StatusCode() != http.StatusOK {
			t.Fatalf("archive the license template: status %d: %s", archiveResp.StatusCode(), archiveResp.Body)
		}
	}
}

// testAccCheckLicenseTemplateArchived asserts that destroy archived every template. Anchor
// keeps the row for good, so the record still resolves and only the status changes.
func testAccCheckLicenseTemplateArchived(t *testing.T) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccClient(t)

		for name, res := range state.RootModule().Resources {
			if res.Type != "anchor_license_template" {
				continue
			}

			getResp, err := client.GetLicenseTemplateWithResponse(
				testAccContext(t),
				res.Primary.Attributes["product_id"],
				res.Primary.ID,
			)
			if err != nil {
				return fmt.Errorf("check %s was archived: %w", name, err)
			}

			// The product is destroyed in the same run, so a gone template is acceptable too.
			if getResp.StatusCode() == http.StatusNotFound {
				continue
			}

			if getResp.JSON200 == nil {
				return fmt.Errorf("check %s was archived: status %d", name, getResp.StatusCode())
			}

			if getResp.JSON200.Status != nanoclient.LicenseTemplateStatusARCHIVED {
				return fmt.Errorf("%s is still %s", name, getResp.JSON200.Status)
			}
		}

		return nil
	}
}
