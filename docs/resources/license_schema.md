---
page_title: "anchor_license_schema Resource - terraform-provider-anchor"
subcategory: ""
description: |-
  Manages the license schema of an Anchor product.
---

# anchor_license_schema (Resource)

Manages the license schema of an Anchor product: every field a license can carry, its
type, and its validation rules.

A product has at most one license schema, so the resource is addressed by product and has
no identifier of its own in the configuration.

## Example Usage

```terraform
resource "anchor_license_schema" "echopoint" {
  product_id  = anchor_product.echopoint.id
  description = "What an Echopoint license can carry."

  fields = [
    {
      name        = "max_flows"
      type        = "LIMIT"
      description = "Flows an organization can hold."
      rules = {
        min = 0
        max = 100000
      }
    },
    {
      name        = "support_tier"
      type        = "ENUM"
      description = "The support the organization is entitled to."
      rules = {
        values = ["none", "standard", "priority"]
      }
    },
    {
      name        = "sso_enabled"
      type        = "BOOLEAN"
      description = "Whether single sign-on is granted."
    },
  ]
}
```

## Every field is mandatory in every template

There is no `required` flag. Every field the schema declares must be set by every license
template of the product. A template that omits one is refused with `LICENSE_FIELD_MISSING`.

A field with an "off" state says so in its own type: a `BOOLEAN` set to `false`, a `LIMIT`
set to `0`, an `ENUM` that declares a `none` value.

Adding a field to the schema is accepted even while the templates still lack it. Anchor
validates but never gates, so a schema edit is never held hostage by a template. The next
edit to each template is refused until the new field is set, so widen a schema in two
steps: add the field, then update every template.

## Fields are replaced wholesale

`fields` is the whole declaration, not a patch. A field you remove from the configuration
is removed from the schema on the next apply.

## Field types

| type | meaning |
| --- | --- |
| `LIMIT` | A ceiling that usage is reported against. Only this type carries usage and a status. |
| `NUMBER` | A plain number that is not a ceiling. |
| `BOOLEAN` | A feature toggle. |
| `ENUM` | One value out of a declared list. |
| `STRING` | Free text. |

## Rules

Rules constrain a license value when it is set. They never apply to a usage report.

| rule | applies to |
| --- | --- |
| `min`, `max` | numeric fields |
| `min_length`, `max_length` | string fields |
| `pattern` | string fields |
| `values` | enum fields |

Omit the `rules` block for a field with no rules. An empty `rules = {}` is refused,
because Anchor gives an absent rule set and an empty one the same answer.

## Import

Import a license schema with the product KSUID. The schema's own KSUID is read back from
the API.

```shell
terraform import anchor_license_schema.echopoint prd_2ikcVW44U7UtqJHCOTqHuwkgrBb
```

## Schema

### Required

- `fields` (Attributes Set) Every license field the schema declares. See [below](#nestedatt--fields).

### Optional

- `product_id` (String) Product KSUID. Defaults to the provider `product_id`. Changing this forces a new resource.
- `description` (String) Optional schema description.

### Read-Only

- `id` (String) License schema KSUID.

<a id="nestedatt--fields"></a>
### Nested Schema for `fields`

#### Required

- `name` (String) Stable identifier used by product code, unique within the schema.
- `type` (String) License field type. One of `BOOLEAN`, `ENUM`, `LIMIT`, `NUMBER`, `STRING`.

#### Optional

- `description` (String) Optional field description.
- `rules` (Attributes) Validation rules applied when a license value is set. See [below](#nestedatt--fields--rules).

<a id="nestedatt--fields--rules"></a>
### Nested Schema for `fields.rules`

#### Optional

- `min` (Number) Inclusive lower bound. Numeric fields only.
- `max` (Number) Inclusive upper bound. Numeric fields only.
- `min_length` (Number) Inclusive minimum length in runes. String fields only.
- `max_length` (Number) Inclusive maximum length in runes. String fields only.
- `pattern` (String) Regular expression the value must match. String fields only.
- `values` (List of String) The list the value must be drawn from. Enum fields only.
