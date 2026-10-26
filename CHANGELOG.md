# Changelog

## Unreleased

### Breaking changes

- `anchor_product_permission` and `anchor_product_role` require `resource_type`, the slug
  of the resource type they are scoped to. Anchor no longer creates either without one
  and never picks a scope. Add `resource_type = data.anchor_product_resource_type.organization.slug`
  to an existing configuration to keep its current scope; Anchor scoped every existing
  permission and role to `organization`, so the next plan shows no change. State written
  by the previous version holds no `resource_type`, and filling it in never replaces the
  resource, even with `-refresh=false`. A different
  value replaces the resource, since a scope never changes. Anchor removes a replaced
  permission from every role, and refuses to replace a role a member holds.
- Needs an Anchor release with resource types. Upgrade Anchor first: against an older
  Anchor, creating a permission or a role fails. Against a newer Anchor, the previous
  provider can no longer create permissions or roles, so upgrade the provider before
  the next apply that creates one.
- After one apply with this version, the previous provider can't read the state, which
  now holds `resource_type`.

### Features

- `anchor_product_resource_type` — manage a product's resource types: the tree of kinds
  of things the product protects, under the built-in `organization` type. Parent types
  are referenced by slug, so Terraform orders creates and destroys through the tree.
- `anchor_product_resource_type` data source — read one type by slug, such as the
  built-in `organization` type that no resource manages.

### Fixes

- `anchor_product_role` keeps an unset `description` null, so a role created without one
  no longer fails the apply with an inconsistent result, and removing a description
  clears it in Anchor.

### Deprecations

- `anchor_product_permission.scope_modifier`: Anchor no longer evaluates it and will
  remove it. Use `resource_type`.

### Notes

- The Anchor Go client moves to a version that has resource types.

## 0.3.0 (2026-08-14)

### Features

- `anchor_license_schema` fields gain `usage_shape` — required when `type` is `LIMIT`,
  refused for every other type. `GAUGE` is a point-in-time reading.
  `WINDOWED_COUNTER` is a count over an explicit window. Anchor refuses a limit
  whose shape is undeclared (ADR-0013), so a schema written without this attribute
  fails against current Anchor. The provider now catches the pairing at plan time.

### Notes

- The Anchor Go client moves to `v0.13.0`, which carries `usage_shape` on
  `LicenseFieldDeclaration`.

## 0.2.0 (2026-08-10)

### Features

- `anchor_license_schema` — manage the license schema of a product: every field a license
  can carry, its type, and its validation rules. A product has at most one schema, so the
  resource is addressed by product. Fields are replaced wholesale on update.
- `anchor_license_template` — manage a named set of license values that satisfies the
  schema. `values` is a JSON object written with `jsonencode` and compared semantically,
  so whitespace and key order never show as drift.
- `anchor_license_template` gains `archived` — set it to `true` and apply to withdraw the
  tier in place, without destroying the resource. It can only move from `false` to `true`;
  a plan that would move it back is refused before any API call is attempted.
- `terraform destroy` on `anchor_license_template` deletes it outright — but only if no
  organization license names it. Anchor refuses otherwise, and the refusal surfaces as a
  clear diagnostic pointing at `archived` as the alternative, rather than the raw API error.

### Notes

- **A template with customers cannot be un-sold; `archived` is how you withdraw it without
  destroying the resource, and it cannot be undone.** An organization's license names the
  template as the statement of what it was sold, so a template Terraform cannot delete
  (because it is referenced) is still withdrawable in place. A template archived outside
  Terraform shows up as drift on the `archived` attribute on the next plan; a configuration
  that still declares `false` is refused, since there is nothing an apply could do to
  satisfy it.
- The provider offers no resource and no data source over an organization's license.
  Licenses are runtime data and carry per-customer adjustments that Terraform would revert.
- Schemas and templates carry no ownership marker and stay editable in the admin UI. A
  conflict shows as ordinary Terraform drift.
- The Anchor Go client moves to `v0.10.1`, which carries the delete route
  `Delete()` on `anchor_license_template` uses, and the `license_template:update`
  scope archiving requires (`nanostack-dev/anchor#86`, `nanostack-dev/anchor#87`).

## 0.1.0 (2026-06-06)

Initial release.

### Features

- `anchor_product` — manage Anchor products.
- `anchor_product_role` — manage product roles and their assigned permissions.
- `anchor_product_permission` — manage product resource permissions (e.g. `flows:create`).
- Provider authentication via `token` or `api_key` + `product_id` (with `ANCHOR_*` env fallbacks).
