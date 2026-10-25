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
  product_id    = anchor_product.example.id
  resource_type = data.anchor_product_resource_type.organization.slug
  name          = "flows:create"
  description   = "Create flows"
}
```

## Schema

### Required

- `name` (String) Permission name, for example `flows:create`. Changing this forces a new resource.
- `resource_type` (String) Slug of the [resource type](product_resource_type.md) the permission is scoped to, for example `organization`. Anchor never picks one. A permission's scope never changes; changing this forces a new resource. Anchor removes a deleted permission from every role, so the roles that bundled it lose it until the next apply adds it back.

### Optional

- `product_id` (String) Product KSUID. Defaults to the provider `product_id`. Changing this forces a new resource.
- `description` (String) Optional permission description.
- `scope_modifier` (String, Deprecated) Optional scope modifier (for example `own` or `team`). Anchor no longer evaluates it and will remove it; use `resource_type`. Changing this forces a new resource.

### Read-Only

- `id` (String) Composite identifier: `<product_id>:<name>`.
