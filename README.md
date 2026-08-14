# Terraform Provider: Anchor

Terraform provider for [Anchor](https://anchorapi.nanostack.dev) — manage products,
product roles, product resource permissions, license schemas, and license templates as
code.

Published to the public Terraform Registry as
[`nanostack-dev/anchor`](https://registry.terraform.io/providers/nanostack-dev/anchor/latest).

## Resources

- `anchor_product`
- `anchor_product_role`
- `anchor_product_permission`
- `anchor_license_schema`
- `anchor_license_template`

## What Terraform does not manage

The provider offers no resource and no data source over an organization's license.

An organization's license is runtime data. It is a copy of a template, and it carries the
bespoke per-customer adjustments an operator makes after the sale. A Terraform-managed
license would revert every one of them on the next apply. Write an organization's license
over the API or through `anchorsdk`.

Schemas and templates stay editable in the admin UI. There is no ownership marker, so a
conflict shows as ordinary Terraform drift and the operator decides.

## Withdrawing a license template

`terraform destroy` on `anchor_license_template` deletes it, but only if no organization
license names it — the destroy fails otherwise, with an error naming the reference, since
an organization's license is the statement of what it was sold and Terraform never manages
one. Withdraw a template that might already have customers in place instead, by setting
`archived = true` on the resource and applying: irreversible, but the resource and its
history stay in state. See
[the resource documentation](./docs/resources/license_template.md).

## Usage

```hcl
terraform {
  required_providers {
    anchor = {
      source  = "nanostack-dev/anchor"
      version = "~> 0.3"
    }
  }
}

provider "anchor" {
  base_url = "https://anchorapi.nanostack.dev"
  token    = var.anchor_token # or api_key + product_id
}

resource "anchor_product" "example" {
  name        = "echopoint"
  description = "Webhook testing platform"
}
```

`terraform init` downloads the provider from the registry — no local build required.

## Authentication

Provide one of:

- `token` — platform bearer token (`Authorization: Bearer ...`), env `ANCHOR_TOKEN`
- `api_key` + `product_id` — product API key (`X-Product-API-Key: ...`), env `ANCHOR_API_KEY` / `ANCHOR_PRODUCT_ID`

Base URL: `base_url` arg or env `ANCHOR_BASE_URL` (defaults to `https://anchorapi.nanostack.dev`).

## Acceptance tests

Acceptance tests create real resources against a real Anchor instance. They create their
own product, so they need a platform bearer token.

```bash
TF_ACC=1 ANCHOR_BASE_URL=https://apidev.tryanchor.dev ANCHOR_TOKEN=... go test ./... -v
```

Without `TF_ACC` the acceptance tests skip and `go test ./...` runs the unit tests only,
which is what CI does.

## Local development

Build and point Terraform at the local binary with a dev override.

```bash
go build -o terraform-provider-anchor
```

`~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "nanostack-dev/anchor" = "/ABSOLUTE/PATH/TO/terraform-provider-anchor"
  }
  direct {}
}
```

With a dev override active, skip `terraform init` and run `terraform plan` directly.

## Releasing

### One-time setup

1. **Generate a GPG signing key** (no expiry, RSA 4096):
   ```bash
   gpg --batch --full-generate-key <<EOF
   Key-Type: RSA
   Key-Length: 4096
   Name-Real: Nanostack Terraform Provider
   Name-Email: ops@nanostack.dev
   Expire-Date: 0
   %no-protection
   %commit
   EOF
   ```
   Get the fingerprint: `gpg --list-secret-keys --keyid-format=long`.

2. **Add repo secrets** (Settings → Secrets and variables → Actions):
   - `GPG_PRIVATE_KEY` — `gpg --armor --export-secret-keys <FINGERPRINT>`
   - `PASSPHRASE` — the key passphrase (empty if `%no-protection` was used)
   ```bash
   gh secret set GPG_PRIVATE_KEY < <(gpg --armor --export-secret-keys <FINGERPRINT>)
   gh secret set PASSPHRASE --body ""
   ```

3. **Connect the provider on the Terraform Registry**:
   - Sign in at https://registry.terraform.io with the GitHub org account.
   - https://registry.terraform.io/publish/provider → select `nanostack-dev/terraform-provider-anchor`.
   - Under the namespace's **GPG keys**, add the **public** key:
     `gpg --armor --export <FINGERPRINT>`.

### Cut a release

```bash
git tag v0.1.0
git push origin v0.1.0
```

The `Release` workflow runs GoReleaser, GPG-signs the checksums, and publishes the
zip archives + `SHA256SUMS` + `SHA256SUMS.sig` + `manifest.json` as a GitHub
release. The Terraform Registry ingests the tag automatically (once connected and
the GPG public key is registered), making `terraform init` work for consumers.

## License

[Apache License 2.0](./LICENSE).
