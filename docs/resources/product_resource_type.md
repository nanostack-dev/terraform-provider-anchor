---
page_title: "anchor_product_resource_type Resource - terraform-provider-anchor"
subcategory: ""
description: |-
  Manages a resource type of an Anchor product.
---

# anchor_product_resource_type (Resource)

Manages a resource type of an Anchor product: a kind of thing the product protects, such as a project or a folder. Types form a tree under the built-in `organization` type, which every product has; read it with the [`anchor_product_resource_type` data source](../data-sources/product_resource_type.md). Permissions and roles are scoped to a type through their `resource_type`.

Reference parent types by their resource or data source, as below, so Terraform creates a type after its parents and destroys it before them. Anchor refuses a change that would break the tree (a cycle, a tree deeper than five levels) or leave a role bundling a permission outside its scope, and refuses to delete a type that another type, a role or a permission still uses; the apply then fails with Anchor's error.

## Example Usage

```terraform
data "anchor_product_resource_type" "organization" {
  product_id = anchor_product.example.id
  slug       = "organization"
}

resource "anchor_product_resource_type" "project" {
  product_id   = anchor_product.example.id
  slug         = "project"
  name         = "Project"
  parent_types = [data.anchor_product_resource_type.organization.slug]
}

resource "anchor_product_resource_type" "folder" {
  product_id      = anchor_product.example.id
  slug            = "folder"
  name            = "Folder"
  parent_types    = [anchor_product_resource_type.project.slug]
  nests_in_itself = true
}
```

## Schema

### Required

- `slug` (String) Lowercase identifier, for example `project`. It never changes; changing this forces a new resource. Anchor refuses to delete a type that other types sit under, or that roles or permissions are scoped to, so move them first.
- `name` (String) Display name.
- `parent_types` (Set of String) Slugs of the types an instance may sit under, for example `organization`.

### Optional

- `product_id` (String) Product KSUID. Defaults to the provider `product_id`. Changing this forces a new resource.
- `description` (String) Optional description.
- `nests_in_itself` (Boolean) Whether an instance may sit under another instance of the same type. Defaults to `false`.

### Read-Only

- `id` (String) Resource type KSUID.

## Import

```shell
terraform import anchor_product_resource_type.folder <product_id>:<resource_type_id>
```

The built-in `organization` type can't be imported; read it with the data source. A change made in the Anchor admin UI shows up as drift on the next `terraform plan`.
