# Implemented architecture

License-template create and update retain their own archive decisions. They
share the archive request and response validation after the initial write
succeeds; neither path replaces plan values with the archive response's values.

Product and permission resources each own one API-to-state conversion used by
create, read and update. Product reads retain the existing state ID; create and
update use the response ID. Permission IDs are reconstructed from the response's
product and permission name. Nil optional response fields become Terraform null.

All six resources and the data source share the provider-data decoder. Absent data leaves resource
configuration untouched; a different type produces the same configuration
diagnostic. Each resource then selects its client and, where applicable, its
default product ID.

`main.go` starts the Terraform Plugin Framework provider. `internal/provider/provider.go` registers six resource types: product, product role, product permission, product resource type, license schema and license template. It registers one data source, product resource type, which reads a type by slug, such as the built-in `organization` type that no resource manages. Each resource maps Terraform plan/configuration/state to Anchor API operations through the generated Go client pinned in [go.mod](../../go.mod).

Provider attributes override environment configuration. Authentication supports platform bearer `token` or `api_key` with `product_id`; environment alternatives are `ANCHOR_TOKEN`, `ANCHOR_API_KEY` and `ANCHOR_PRODUCT_ID`. `base_url` or `ANCHOR_BASE_URL` chooses the target, defaulting to the API in [provider.go](../../internal/provider/provider.go). Credentials are sensitive schema fields; keep them out of docs and committed configuration.

Products, resource types, roles, permissions, schemas and templates are declarative product configuration. Permissions and roles name their resource type by slug; a scope never changes in Anchor, so a new `resource_type` replaces them, except over state written before the attribute existed, which holds null. Organization licenses are runtime customer data with bespoke adjustments, so Terraform offers neither resources nor data sources for them. [Provider surface tests](../../internal/provider/provider_surface_test.go) enforce this boundary; the original ADR-0006 is owned by [Anchor](https://github.com/nanostack-dev/anchor).

Templates and schemas can also be edited in the admin UI; there is no ownership marker, so conflicts appear as ordinary Terraform drift. Referenced license templates cannot be destroyed. Archiving withdraws one irreversibly while preserving history; [license template docs](../resources/license_template.md) own exact semantics.

This repository publishes a provider binary and Registry documentation. The Anchor service owns API behavior, authorization, tenancy and deployment. Ordinary provider unit tests and builds need no sibling checkout; acceptance tests need the explicitly selected API.
