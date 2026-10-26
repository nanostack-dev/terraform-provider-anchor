package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccProductResourceTypes covers a product's catalog as code: a tree of resource types
// under the built-in organization type, a permission and a role scoped through it, an
// update of the tree, and an import.
func TestAccProductResourceTypes(t *testing.T) {
	productName := acctest.RandomWithPrefix("tfacc-rtype")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProductResourceTypesConfig(productName, "Folders", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.anchor_product_resource_type.organization", "built_in", "true"),
					resource.TestCheckResourceAttr("data.anchor_product_resource_type.organization", "parent_types.#", "0"),
					resource.TestCheckResourceAttrSet("anchor_product_resource_type.project", "id"),
					resource.TestCheckTypeSetElemAttr("anchor_product_resource_type.project", "parent_types.*", "organization"),
					resource.TestCheckTypeSetElemAttr("anchor_product_resource_type.folder", "parent_types.*", "project"),
					resource.TestCheckResourceAttr("anchor_product_resource_type.folder", "nests_in_itself", "true"),
					resource.TestCheckResourceAttr("anchor_product_permission.view", "resource_type", "folder"),
					resource.TestCheckResourceAttr("anchor_product_role.lead", "resource_type", "project"),
					resource.TestCheckResourceAttr("anchor_product_resource_type.project", "description", "Work grouped together."),
				),
			},
			{
				ResourceName:      "anchor_product_permission.view",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccProductPermissionImportID("anchor_product_permission.view"),
			},
			{
				ResourceName:      "anchor_product_role.lead",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccProductResourceTypeImportID("anchor_product_role.lead"),
			},
			{
				ResourceName:      "anchor_product_resource_type.folder",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccProductResourceTypeImportID("anchor_product_resource_type.folder"),
			},
			{
				Config: testAccProductResourceTypesConfig(productName, "Folder", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("anchor_product_resource_type.folder", "name", "Folder"),
					resource.TestCheckNoResourceAttr("anchor_product_resource_type.project", "description"),
				),
			},
		},
	})
}

func testAccProductResourceTypesConfig(productName, folderName string, described bool) string {
	description := ""
	if described {
		description = `description  = "Work grouped together."`
	}
	return fmt.Sprintf(`
resource "anchor_product" "test" {
  name        = %[1]q
  description = "Terraform provider acceptance test."
}

data "anchor_product_resource_type" "organization" {
  product_id = anchor_product.test.id
  slug       = "organization"
}

resource "anchor_product_resource_type" "project" {
  product_id   = anchor_product.test.id
  slug         = "project"
  name         = "Project"
  parent_types = [data.anchor_product_resource_type.organization.slug]
  %[3]s
}

resource "anchor_product_resource_type" "folder" {
  product_id      = anchor_product.test.id
  slug            = "folder"
  name            = %[2]q
  parent_types    = [anchor_product_resource_type.project.slug]
  nests_in_itself = true
}

resource "anchor_product_permission" "view" {
  product_id    = anchor_product.test.id
  name          = "document:view"
  resource_type = anchor_product_resource_type.folder.slug
}

resource "anchor_product_role" "lead" {
  product_id    = anchor_product.test.id
  name          = "project-lead"
  resource_type = anchor_product_resource_type.project.slug
  permissions   = [anchor_product_permission.view.name]
}
`, productName, folderName, description)
}

func testAccProductResourceTypeImportID(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		found, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		return found.Primary.Attributes["product_id"] + ":" + found.Primary.ID, nil
	}
}

func testAccProductPermissionImportID(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		found, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		return found.Primary.Attributes["product_id"] + ":" + found.Primary.Attributes["name"], nil
	}
}
