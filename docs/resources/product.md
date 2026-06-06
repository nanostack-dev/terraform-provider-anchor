---
page_title: "anchor_product Resource - terraform-provider-anchor"
subcategory: ""
description: |-
  Manages an Anchor product.
---

# anchor_product (Resource)

Manages an Anchor product.

## Example Usage

```terraform
resource "anchor_product" "example" {
  name        = "echopoint"
  description = "Webhook testing platform"
}
```

## Schema

### Required

- `name` (String) Product name.

### Optional

- `description` (String) Optional product description.

### Read-Only

- `id` (String) Product KSUID.
