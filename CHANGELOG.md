# Changelog

## Unreleased

### Features

- `anchor_license_schema` — manage the license schema of a product: every field a license
  can carry, its type, and its validation rules. A product has at most one schema, so the
  resource is addressed by product. Fields are replaced wholesale on update.
- `anchor_license_template` — manage a named set of license values that satisfies the
  schema. `values` is a JSON object written with `jsonencode` and compared semantically,
  so whitespace and key order never show as drift.

### Notes

- **Destroying an `anchor_license_template` archives it, and archiving cannot be undone.**
  Anchor has no delete for a template, because an organization's license names it as the
  statement of what it was sold. A template archived outside Terraform is treated as gone,
  so the next plan proposes a replacement, which keeps the name because archiving frees it.
- The provider offers no resource and no data source over an organization's license.
  Licenses are runtime data and carry per-customer adjustments that Terraform would revert.
- Schemas and templates carry no ownership marker and stay editable in the admin UI. A
  conflict shows as ordinary Terraform drift.
- The Anchor Go client moves to `v0.9.0`, which carries the licensing routes.

## 0.1.0 (2026-06-06)

Initial release.

### Features

- `anchor_product` — manage Anchor products.
- `anchor_product_role` — manage product roles and their assigned permissions.
- `anchor_product_permission` — manage product resource permissions (e.g. `flows:create`).
- Provider authentication via `token` or `api_key` + `product_id` (with `ANCHOR_*` env fallbacks).
