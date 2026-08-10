package provider_test

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	nanoclient "github.com/nanostack-dev/anchor/clients/go"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccLicenseTemplate covers the whole life of a license template: create it, edit its
// values, report drift after an edit made outside Terraform, correct that drift, and
// destroy it. This template is never referenced by an organization license, so destroy
// removes it outright — see TestAccLicenseTemplateDeleteBlockedWhenReferenced for the
// other case, and TestAccLicenseTemplateArchiveInPlace for withdrawing one without
// destroying the resource.
func TestAccLicenseTemplate(t *testing.T) {
	productName := acctest.RandomWithPrefix("tfacc-template")
	templateName := acctest.RandomWithPrefix("pro")

	var productID string
	var templateID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckLicenseTemplateDestroyed(t),
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
					resource.TestCheckResourceAttr("anchor_license_template.test", "archived", "false"),
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

// TestAccLicenseTemplateArchivedOutsideTerraform proves the recovery path for a template
// archived behind Terraform's back. archived is a real, drift-checked attribute: the
// out-of-band archive is caught as ordinary drift, and a config that still declares
// archived = false cannot be applied — there is nothing the apply could do to satisfy it,
// since Anchor has no route back from archived. Setting archived = true in config is what
// converges.
func TestAccLicenseTemplateArchivedOutsideTerraform(t *testing.T) {
	productName := acctest.RandomWithPrefix("tfacc-archived")
	templateName := acctest.RandomWithPrefix("pro")

	var productID string
	var templateID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckLicenseTemplateDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseTemplateConfigWithArchived(productName, templateName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureAttr("anchor_license_template.test", "product_id", &productID),
					testAccCaptureAttr("anchor_license_template.test", "id", &templateID),
					resource.TestCheckResourceAttr("anchor_license_template.test", "archived", "false"),
				),
			},
			{
				// The config still says archived = false, so the refreshed drift and the
				// config disagree. There is no way to satisfy "false" against a template
				// Anchor already archived, so the plan itself is refused.
				PreConfig: testAccArchiveLicenseTemplate(t, &productID, &templateID),
				Config:    testAccLicenseTemplateConfigWithArchived(productName, templateName, false),
				PlanOnly:  true,
				ExpectError: regexp.MustCompile(
					"Cannot Un-Archive a License Template",
				),
			},
			{
				// Updating the configuration to match reality is what converges.
				Config: testAccLicenseTemplateConfigWithArchived(productName, templateName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("anchor_license_template.test", "name", templateName),
					resource.TestCheckResourceAttr(
						"anchor_license_template.test", "values",
						`{"max_flows":100,"sso_enabled":true}`,
					),
					resource.TestCheckResourceAttr("anchor_license_template.test", "archived", "true"),
				),
			},
		},
	})
}

// TestAccLicenseTemplateArchiveInPlace archives a template by declaring archived = true and
// applying, without destroying the resource — the alternative to a Terraform-destroy
// archive when the operator wants to keep managing the record. Setting archived back to
// false is refused: Anchor has no route back from archived.
func TestAccLicenseTemplateArchiveInPlace(t *testing.T) {
	productName := acctest.RandomWithPrefix("tfacc-archive-inplace")
	templateName := acctest.RandomWithPrefix("pro")

	var productID string
	var templateID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckLicenseTemplateDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseTemplateConfigWithArchived(productName, templateName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureAttr("anchor_license_template.test", "product_id", &productID),
					testAccCaptureAttr("anchor_license_template.test", "id", &templateID),
					resource.TestCheckResourceAttr("anchor_license_template.test", "archived", "false"),
				),
			},
			{
				// Declaring archived = true and applying withdraws the tier without
				// destroying the resource — the record stays in state.
				Config: testAccLicenseTemplateConfigWithArchived(productName, templateName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("anchor_license_template.test", "archived", "true"),
					testAccCheckLicenseTemplateStatus(
						t, &productID, &templateID, nanoclient.LicenseTemplateStatusARCHIVED,
					),
				),
			},
			{
				// There is no way back. The plan modifier refuses this before any API
				// call is attempted.
				Config: testAccLicenseTemplateConfigWithArchived(productName, templateName, false),
				ExpectError: regexp.MustCompile(
					"Cannot Un-Archive a License Template",
				),
			},
		},
	})
}

// TestAccLicenseTemplateDeleteBlockedWhenReferenced proves the guard ADR-0011 (in the
// anchor repository) added to DELETE: Terraform's destroy is refused, with a clear
// diagnostic, while an organization license names the template. Organization licenses
// are API-only (ADR-0006) — Terraform cannot resolve the reference itself, so the test
// releases it directly through the client, the same way an operator would.
func TestAccLicenseTemplateDeleteBlockedWhenReferenced(t *testing.T) {
	productName := acctest.RandomWithPrefix("tfacc-template-inuse")
	templateName := acctest.RandomWithPrefix("pro")

	var productID string
	var templateID string
	var organizationID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckLicenseTemplateDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseTemplateConfigWithArchived(productName, templateName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureAttr("anchor_license_template.test", "product_id", &productID),
					testAccCaptureAttr("anchor_license_template.test", "id", &templateID),
					testAccInstantiateOrganizationLicense(t, &productID, &templateID, &organizationID),
				),
			},
			{
				// The configuration no longer declares the template, so applying plans
				// its destruction. Anchor refuses: the organization above still holds a
				// license naming it, and Terraform surfaces that refusal rather than a
				// raw API error dump.
				Config: testAccProductOnlyConfig(productName),
				ExpectError: regexp.MustCompile(
					"(?s)License Template Is Still In Use.*organization license",
				),
			},
			{
				// Releasing the reference is what unblocks the same destroy. Deleting the
				// organization cascades its license away — organization_licenses has
				// ON DELETE CASCADE on the organization foreign key — which is the
				// simplest release available; there is no direct "unlicense" route.
				PreConfig: testAccDeleteOrganization(t, &productID, &organizationID),
				Config:    testAccProductOnlyConfig(productName),
			},
		},
	})
}

// testAccInstantiateOrganizationLicense creates an organization and instantiates the
// template onto it, so the template becomes one a real customer was sold — outside
// Terraform, since organization licenses are never managed here.
func testAccInstantiateOrganizationLicense(
	t *testing.T, productID, templateID, organizationID *string,
) resource.TestCheckFunc {
	return func(*terraform.State) error {
		t.Helper()

		if *productID == "" || *templateID == "" {
			return errors.New("no template captured from the Terraform state")
		}

		// Platform bearer auth answers 401 for organization and license routes; both
		// need a product API key, minted here rather than through testAccClient.
		client := testAccProductAPIKeyClient(
			t, *productID, []string{"organization:create", "organization_license:create"},
		)
		ctx := testAccContext(t)

		orgResp, err := client.CreateProductOrganizationWithResponse(
			ctx, *productID,
			nanoclient.CreateProductOrganizationJSONRequestBody{Name: acctest.RandomWithPrefix("tfacc-org")},
		)
		if err != nil {
			return fmt.Errorf("create organization: %w", err)
		}
		if orgResp.JSON201 == nil {
			return fmt.Errorf("create organization: status %d: %s", orgResp.StatusCode(), orgResp.Body)
		}
		*organizationID = orgResp.JSON201.Id

		instResp, err := client.InstantiateOrganizationLicenseWithResponse(
			ctx, *productID, *organizationID,
			nanoclient.InstantiateOrganizationLicenseJSONRequestBody{TemplateId: *templateID},
		)
		if err != nil {
			return fmt.Errorf("instantiate organization license: %w", err)
		}
		if instResp.JSON201 == nil {
			return fmt.Errorf(
				"instantiate organization license: status %d: %s", instResp.StatusCode(), instResp.Body,
			)
		}

		return nil
	}
}

// testAccDeleteOrganization releases a template's only reference by removing the
// organization licensed from it.
func testAccDeleteOrganization(t *testing.T, productID, organizationID *string) func() {
	return func() {
		t.Helper()

		if *productID == "" || *organizationID == "" {
			t.Fatal("no organization captured to release")
		}

		client := testAccProductAPIKeyClient(t, *productID, []string{"organization:delete"})
		resp, err := client.DeleteProductOrganizationWithResponse(
			testAccContext(t), *productID, *organizationID,
		)
		if err != nil {
			t.Fatalf("delete organization: %v", err)
		}
		if resp.StatusCode() != http.StatusNoContent {
			t.Fatalf("delete organization: status %d: %s", resp.StatusCode(), resp.Body)
		}
	}
}

// testAccProductOnlyConfig is testAccLicenseTemplateConfigWithArchived's product and
// schema, without the template — declaring this after a template has been created is
// what plans the template's destruction, since it drops out of the configuration.
func testAccProductOnlyConfig(productName string) string {
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
`, productName)
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

// testAccLicenseTemplateConfigWithArchived is testAccLicenseTemplateConfig plus an explicit
// archived argument, for the tests whose subject is that attribute. Values are fixed: these
// tests are about the archived transition, not about what the template holds.
func testAccLicenseTemplateConfigWithArchived(productName, templateName string, archived bool) string {
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
  archived    = %[3]t

  values = jsonencode({
    max_flows   = 100
    sso_enabled = true
  })

  depends_on = [anchor_license_schema.test]
}
`, productName, templateName, archived)
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

// testAccCheckLicenseTemplateStatus asserts the template's status directly against the API,
// independent of what Terraform's own state says.
func testAccCheckLicenseTemplateStatus(
	t *testing.T, productID, templateID *string, want nanoclient.LicenseTemplateStatus,
) resource.TestCheckFunc {
	return func(*terraform.State) error {
		t.Helper()

		client := testAccClient(t)
		getResp, err := client.GetLicenseTemplateWithResponse(testAccContext(t), *productID, *templateID)
		if err != nil {
			return fmt.Errorf("read license template status: %w", err)
		}
		if getResp.JSON200 == nil {
			return fmt.Errorf("read license template status: status %d: %s", getResp.StatusCode(), getResp.Body)
		}
		if getResp.JSON200.Status != want {
			return fmt.Errorf("license template status is %s, want %s", getResp.JSON200.Status, want)
		}

		return nil
	}
}

// testAccCheckLicenseTemplateDestroyed asserts that destroy actually removed every
// template none of these tests ever reference from an organization: with the real
// DELETE route wired, an unreferenced template's row is gone, not merely archived.
func testAccCheckLicenseTemplateDestroyed(t *testing.T) resource.TestCheckFunc {
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
				return fmt.Errorf("check %s was destroyed: %w", name, err)
			}

			if getResp.StatusCode() != http.StatusNotFound {
				return fmt.Errorf("%s still exists: status %d: %s", name, getResp.StatusCode(), getResp.Body)
			}
		}

		return nil
	}
}
