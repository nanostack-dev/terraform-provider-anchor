data "anchor_product_resource_type" "organization" {
  product_id = anchor_product.example.id
  slug       = "organization"
}

resource "anchor_product_resource_type" "project" {
  product_id   = anchor_product.example.id
  slug         = "project"
  name         = "Project"
  parent_types = [data.anchor_product_resource_type.organization.slug]
}

resource "anchor_product_resource_type" "folder" {
  product_id      = anchor_product.example.id
  slug            = "folder"
  name            = "Folder"
  parent_types    = [anchor_product_resource_type.project.slug]
  nests_in_itself = true
}
