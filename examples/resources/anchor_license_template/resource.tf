# Destroying a license template archives it. Archiving cannot be undone.
resource "anchor_license_template" "free" {
  product_id  = anchor_product.echopoint.id
  name        = "Free"
  description = "The tier every new organization starts on."

  values = jsonencode({
    max_flows    = 5
    support_tier = "none"
    sso_enabled  = false
  })

  depends_on = [anchor_license_schema.echopoint]
}

resource "anchor_license_template" "pro" {
  product_id  = anchor_product.echopoint.id
  name        = "Pro"
  description = "The paid tier."

  values = jsonencode({
    max_flows    = 500
    support_tier = "priority"
    sso_enabled  = true
  })

  depends_on = [anchor_license_schema.echopoint]
}
