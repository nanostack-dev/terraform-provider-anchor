---
page_title: "anchor_product_resource_type Data Source - terraform-provider-anchor"
subcategory: ""
description: |-
  Reads one resource type of an Anchor product by slug.
---

# anchor_product_resource_type (Data Source)

Reads one resource type of an Anchor product by slug. Its main use is the built-in `organization` type that every product has, which no resource manages, so that permissions, roles and other types can reference it.

## Example Usage

```terraform
data "anchor_product_resource_type" "organization" {
  product_id = anchor_product.example.id
  slug       = "organization"
}
```

## Schema

### Required

- `slug` (String) Slug of the type to read, for example `organization`.

### Optional

- `product_id` (String) Product KSUID. Defaults to the provider `product_id`.

### Read-Only

- `id` (String) Resource type KSUID.
- `name` (String) Display name.
- `description` (String) Description.
- `parent_types` (Set of String) Slugs of the types an instance may sit under. Empty for `organization`.
- `nests_in_itself` (Boolean) Whether an instance may sit under another instance of the same type.
- `built_in` (Boolean) `true` for the `organization` type every product has.
