---
page_title: "anchor_product_role Resource - terraform-provider-anchor"
subcategory: ""
description: |-
  Manages an Anchor product role.
---

# anchor_product_role (Resource)

Manages an Anchor product role.

## Example Usage

```terraform
resource "anchor_product_role" "admin" {
  product_id  = anchor_product.example.id
  name        = "admin"
  description = "Administrator role"
  permissions = [
    "flows:create",
    "flows:read",
  ]
}
```

## Schema

### Required

- `name` (String) Role name.

### Optional

- `product_id` (String) Product KSUID. Defaults to the provider `product_id`. Changing this forces a new resource.
- `description` (String) Optional role description.
- `permissions` (Set of String) Resource permission names assigned to this role.

### Read-Only

- `id` (String) Product role KSUID.
