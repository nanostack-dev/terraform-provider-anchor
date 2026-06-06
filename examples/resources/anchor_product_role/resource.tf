resource "anchor_product_role" "admin" {
  product_id  = anchor_product.example.id
  name        = "admin"
  description = "Administrator role"
  permissions = [
    "flows:create",
    "flows:read",
  ]
}
