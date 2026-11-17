# Copyright (c) Trifork

variable "azure_tenant_id" {
  type = string
}

variable "azure_client_id" {
  type = string
}

variable "azure_client_secret" {
  type      = string
  sensitive = true
}

variable "azure_certificate_password" {
  type      = string
  sensitive = true
  default   = ""
}

# The provider block authenticates to Corax as usual, independently of the
# identities configured below for connecting Corax to Azure model endpoints.
provider "corax" {}

resource "corax_model_provider" "foundry_client_secret" {
  name          = "Foundry client secret"
  provider_type = "azure_ai_service_principal"
  configuration = {
    api_endpoint  = "https://example.services.ai.azure.com/"
    tenant_id     = var.azure_tenant_id
    client_id     = var.azure_client_id
    client_secret = var.azure_client_secret
  }
}

# An administrator mounts the combined PEM certificate and RSA private key on
# each Corax backend as <client-id>.pem under CORAX_CERTIFICATES_DIR.
resource "corax_model_provider" "foundry_certificate" {
  name          = "Foundry certificate"
  provider_type = "azure_ai_service_principal_certificate"
  configuration = {
    api_endpoint         = "https://example.services.ai.azure.com/"
    tenant_id            = var.azure_tenant_id
    client_id            = var.azure_client_id
    certificate_password = var.azure_certificate_password
  }
}
