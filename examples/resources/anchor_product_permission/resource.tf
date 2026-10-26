resource "anchor_product_permission" "flows_create" {
  product_id    = anchor_product.example.id
  resource_type = data.anchor_product_resource_type.organization.slug
  name          = "flows:create"
  description   = "Create flows"
}
