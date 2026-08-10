# terraform destroy deletes a template outright, but only if no organization license
# names it — a fixture like this one that nothing was ever sold from is destroyable.
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

# A withdrawn tier that is still worth tracking in Terraform: set archived = true and
# apply, rather than removing the resource. There is no route back from archived — a
# plan that tried to set this to false again would be refused.
resource "anchor_license_template" "legacy" {
  product_id  = anchor_product.echopoint.id
  name        = "Legacy"
  description = "Withdrawn tier, kept for the organizations still licensed from it."
  archived    = true

  values = jsonencode({
    max_flows    = 50
    support_tier = "standard"
    sso_enabled  = false
  })

  depends_on = [anchor_license_schema.echopoint]
}
