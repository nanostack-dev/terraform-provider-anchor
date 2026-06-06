---
page_title: "anchor_product_permission Resource - terraform-provider-anchor"
subcategory: ""
description: |-
  Manages an Anchor product resource permission.
---

# anchor_product_permission (Resource)

Manages an Anchor product resource permission (for example `flows:create`).

## Example Usage

```terraform
resource "anchor_product_permission" "flows_create" {
  product_id  = anchor_product.example.id
  name        = "flows:create"
  description = "Create flows"
}
```

## Schema

### Required

- `name` (String) Permission name, for example `flows:create`. Changing this forces a new resource.

### Optional

- `product_id` (String) Product KSUID. Defaults to the provider `product_id`. Changing this forces a new resource.
- `description` (String) Optional permission description.
- `scope_modifier` (String) Optional scope modifier (for example `own` or `team`). Changing this forces a new resource.

### Read-Only

- `id` (String) Composite identifier: `<product_id>:<name>`.
