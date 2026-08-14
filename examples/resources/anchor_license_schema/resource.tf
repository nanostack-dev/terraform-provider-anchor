resource "anchor_license_schema" "echopoint" {
  product_id  = anchor_product.echopoint.id
  description = "What an Echopoint license can carry."

  fields = [
    {
      name        = "max_flows"
      type        = "LIMIT"
      usage_shape = "GAUGE"
      description = "Flows an organization can hold."
      rules = {
        min = 0
        max = 100000
      }
    },
    {
      name        = "support_tier"
      type        = "ENUM"
      description = "The support the organization is entitled to."
      rules = {
        values = ["none", "standard", "priority"]
      }
    },
    {
      name        = "sso_enabled"
      type        = "BOOLEAN"
      description = "Whether single sign-on is granted."
    },
  ]
}
