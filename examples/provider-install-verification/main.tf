# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    corax = {
      source = "trifork/corax"
    }
  }
}

provider "corax" {
  api_endpoint = "https://api.corax.app"
  api_key      = "<api-key>"
}
