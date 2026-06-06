---
page_title: "Anchor Provider"
description: |-
  The Anchor provider manages products, product roles, and product resource permissions.
---

# Anchor Provider

The Anchor provider manages [Anchor](https://anchorapi.nanostack.dev) products,
product roles, and product resource permissions.

## Example Usage

```terraform
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
  token    = var.anchor_token
}
```

## Authentication

Provide one of:

- `token` — platform bearer token, or env `ANCHOR_TOKEN`.
- `api_key` together with `product_id` — product API key, or env `ANCHOR_API_KEY` and `ANCHOR_PRODUCT_ID`.

## Schema

### Optional

- `base_url` (String) Anchor API base URL. Can also be set with `ANCHOR_BASE_URL`. Defaults to `https://anchorapi.nanostack.dev`.
- `token` (String, Sensitive) Platform bearer token. Can also be set with `ANCHOR_TOKEN`.
- `api_key` (String, Sensitive) Product API key sent as `X-Product-API-Key`. Can also be set with `ANCHOR_API_KEY`.
- `product_id` (String) Default product ID for product-scoped resources. Can also be set with `ANCHOR_PRODUCT_ID`.
