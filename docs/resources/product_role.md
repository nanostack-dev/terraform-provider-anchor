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
  product_id    = anchor_product.example.id
  resource_type = data.anchor_product_resource_type.organization.slug
  name          = "admin"
  description   = "Administrator role"
  permissions = [
    "flows:create",
    "flows:read",
  ]
}
```

## Schema

### Required

- `name` (String) Role name.
- `resource_type` (String) Slug of the [resource type](product_resource_type.md) the role is scoped to, for example `organization`. Anchor never picks one. Every permission of the role must be scoped to this type or to a type below it. A role's scope never changes; changing this forces a new resource, which Anchor refuses while a member holds the role (`ROLE_IN_USE`).

### Optional

- `product_id` (String) Product KSUID. Defaults to the provider `product_id`. Changing this forces a new resource.
- `description` (String) Optional role description.
- `permissions` (Set of String) Resource permission names assigned to this role. Each must be scoped to the role's resource type or to a type below it.

### Read-Only

- `id` (String) Product role KSUID.
