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

## Withdrawing a template

A template with customers cannot be un-sold, so Anchor offers two ways to withdraw one and
they answer two different questions.

### Destroying the resource: delete if nobody was ever sold it

`terraform destroy` calls Anchor's `DELETE`, which removes the row outright — but only if
no organization license names the template. If one does, the destroy fails with a clear
error naming the reference, and the resource stays in state exactly as Terraform's destroy
semantics already promise on any failed destroy. Terraform cannot resolve that reference
itself: organization licenses are API-only, and this provider offers no resource or data
source over one. Release the reference outside Terraform, or archive the tier in place
instead (below) rather than destroying the resource.

This is the ordinary path for a template Terraform created and is now retiring in the same
breath — a tier drafted and abandoned before anyone bought it, or a fixture an acceptance
test tears down after itself.

### In place, with `archived`: withdraw a tier that has customers

```terraform
resource "anchor_license_template" "pro" {
  # ...
  archived = true
}
```

Set `archived = true` and apply. Archiving works whether or not the template is
referenced, and the resource stays in state, so its history and its values remain visible
in the configuration — the way to withdraw a tier without losing track of it in Terraform.
**This is irreversible.** Anchor has no route back from archived: `archived` can only move
from `false` to `true`, and a plan that would move it back is refused before any API call
is attempted.

If a template was archived outside Terraform, the next plan reports it as drift on this
attribute; if the configuration still declares `false`, that plan is refused too, and the
fix is to set `archived = true` to match reality (or `terraform state rm` this resource if
you no longer want to track it).

An operator who archives the wrong tier recreates it as a new template with the same name
— archiving frees the name — and the two rows are then distinguishable only by their
identifiers and dates.

## Drift

A template edited in the admin UI shows as ordinary drift on the next plan. There is no
ownership marker and the UI stays editable, so the operator decides whether to keep the
edit or let the next apply revert it. An archived template is read faithfully rather than
treated as gone — see `archived`, above.

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

An archived template imports fine — `archived` reads as `true`.

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
