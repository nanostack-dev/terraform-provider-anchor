---
page_title: "anchor_license_template Resource - terraform-provider-anchor"
subcategory: ""
description: |-
  Manages an Anchor license template.
---

# anchor_license_template (Resource)

Manages an Anchor license template: a named set of values that satisfies the product's
license schema. "Free" and "Pro" are license templates.

An organization's license is a copy of a template, taken once when the license is
instantiated. Editing a template later never changes a live organization.

## Example Usage

```terraform
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
```

Declare the dependency on the schema. Anchor validates a template against the schema on
every write, so the schema must exist first.

## Archiving a template

**Withdrawing a tier is irreversible. Anchor has no route back from archived.**

Every organization licensed from a template names it as the statement of what they were
sold, so a template is never deleted — it is archived, `POST .../templates/{id}/archive`.
What archiving means:

- The template can no longer be edited or instantiated.
- It still resolves by identifier and still appears in the product's template listing.
- Its name is freed, so a replacement template can take it.
- Organizations already licensed from it keep their own copy of the values, unchanged.

An operator who archives the wrong tier recreates it as a new template with the same name.
The two rows are then distinguishable only by their identifiers and dates.

There are two ways to archive a template through this provider:

### In place, with `archived`

```terraform
resource "anchor_license_template" "pro" {
  # ...
  archived = true
}
```

Set `archived = true` and apply. The resource stays in state, so its history and its
values remain visible in the configuration. This is the way to withdraw a tier without
losing track of it in Terraform.

`archived` can only move from `false` to `true`. A plan that would move it back is refused
before any API call is attempted — there is nothing an apply could do to satisfy it. If a
template was archived outside Terraform, the next plan reports it as drift on this
attribute; if the configuration still declares `false`, that plan is refused too, and the
fix is to set `archived = true` to match reality (or `terraform state rm` this resource if
you no longer want to track it).

### By destroying the resource

**`terraform destroy` also archives the template**, since archiving is the only withdrawal
the API offers. Destroying removes the resource from Terraform's state the way `destroy`
normally does; the row itself is kept archived in Anchor, unaffected by the resource
leaving state.

## Drift

A template edited in the admin UI shows as ordinary drift on the next plan. There is no
ownership marker and the UI stays editable, so the operator decides whether to keep the
edit or let the next apply revert it.

A template **archived** outside Terraform is treated as gone: it can be neither edited nor
instantiated, so the next plan proposes creating a replacement. Because archiving frees the
name, the replacement keeps the name the configuration declares.

## Values

`values` is a JSON object, one value per license field the schema declares. Write it with
`jsonencode`.

- Every declared field must be set. A template that omits one is refused with
  `LICENSE_FIELD_MISSING`.
- A key the schema does not declare is refused with `LICENSE_FIELD_UNKNOWN`.
- Values are replaced wholesale on update: a field absent from the configuration is unset.
- The value is compared semantically, so whitespace and key order never show as drift.

## Import

Import a license template with the product KSUID and the template KSUID, separated by a
colon.

```shell
terraform import anchor_license_template.pro prd_2ikcVW44U7UtqJHCOTqHuwkgrBb:license_template_5mNOPqRsTuVwXyZ
```

An archived template cannot be imported.

## Schema

### Required

- `name` (String) Operator-facing name, unique among the product's active templates.
- `values` (String) A JSON object holding one value per license field the schema declares.

### Optional

- `product_id` (String) Product KSUID. Defaults to the provider `product_id`. Changing this forces a new resource.
- `description` (String) Optional template description.
- `archived` (Boolean) Whether the template is archived. Set to true and apply to withdraw the tier in place. Can only move from `false` to `true`. Defaults to `false`.

### Read-Only

- `id` (String) License template KSUID.
