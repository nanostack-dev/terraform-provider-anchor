resource "anchor_product_permission" "flows_create" {
  product_id  = anchor_product.example.id
  name        = "flows:create"
  description = "Create flows"
}
