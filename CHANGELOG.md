# Changelog

## Unreleased

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
- The Anchor Go client moves to `v0.9.0`, which carries the licensing routes, then to a
  pseudo-version pinned at `nanostack-dev/anchor#86`'s commit, which carries the delete
  route `Delete()` on `anchor_license_template` now uses. **That pin must be replaced
  with a real tagged version before this is released** — see the note in `go.mod`.

## 0.1.0 (2026-06-06)

Initial release.

### Features

- `anchor_product` — manage Anchor products.
- `anchor_product_role` — manage product roles and their assigned permissions.
- `anchor_product_permission` — manage product resource permissions (e.g. `flows:create`).
- Provider authentication via `token` or `api_key` + `product_id` (with `ANCHOR_*` env fallbacks).
