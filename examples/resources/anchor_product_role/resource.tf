resource "anchor_product_role" "admin" {
  product_id    = anchor_product.example.id
  resource_type = data.anchor_product_resource_type.organization.slug
  name          = "admin"
  description   = "Administrator role"
  permissions = [
    "flows:create",
    "flows:read",
  ]
}
