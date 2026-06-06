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
