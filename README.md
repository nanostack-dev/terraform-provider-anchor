# Terraform Provider: Anchor

Terraform provider for [Anchor](https://anchorapi.nanostack.dev) — manage products,
product roles, and product resource permissions as code.

Published to the public Terraform Registry as
[`nanostack-dev/anchor`](https://registry.terraform.io/providers/nanostack-dev/anchor/latest).

## Resources

- `anchor_product`
- `anchor_product_role`
- `anchor_product_permission`

## Usage

```hcl
terraform {
  required_providers {
    anchor = {
      source  = "nanostack-dev/anchor"
      version = "~> 0.1"
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

[FSL-1.1-ALv2](./LICENSE) (Functional Source License; converts to Apache-2.0).
