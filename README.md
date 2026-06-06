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

Releases are cut by pushing a semver tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The `Release` workflow runs GoReleaser, GPG-signs the checksums, and publishes the
archives + `SHA256SUMS` + `SHA256SUMS.sig` + manifest as a GitHub release. The
Terraform Registry ingests the tag automatically once the provider is connected
and the GPG public key is registered.

Requires repo secrets `GPG_PRIVATE_KEY` and `PASSPHRASE`.

## License

[FSL-1.1-ALv2](./LICENSE) (Functional Source License; converts to Apache-2.0).
