# Copyright (c) Trifork

terraform {
  required_providers {
    corax = {
      source = "registry.terraform.io/trifork/corax"
    }
  }
}

# api_endpoint / api_key come from CORAX_API_ENDPOINT / CORAX_API_KEY.
provider "corax" {}

locals {
  prefix = "${var.name_prefix}-${var.run_id}"
}

resource "corax_project" "test" {
  name        = "${local.prefix}-project"
  description = "Created by the terraform-provider-corax integration harness (rev ${var.revision})."
  is_public   = false
}

resource "corax_mcp_server" "test" {
  name      = "${local.prefix}-mcp"
  url       = "https://example.invalid/mcp"
  type      = "streamablehttp"
  is_public = false

  config = {
    token = {
      type    = "header"
      label   = "Authorization"
      default = null
    }
    filters = {
      type     = "header"
      label    = "X-Filters"
      default  = null
      required = false
    }
  }
}

resource "corax_chat_capability" "test" {
  name          = "${local.prefix}-chat"
  semantic_id   = "${local.prefix}-chat"
  project_id    = corax_project.test.id
  system_prompt = "You are a terse assistant used by an integration test (rev ${var.revision})."
  is_public     = false

  config = {
    temperature = var.revision * 0.1
    # Must be false alongside timed data retention - the API forces it.
    content_tracing = false
    mcp_server_ids  = [corax_mcp_server.test.id]

    data_retention = {
      type  = "timed"
      hours = 24
    }

    blob_config = {
      max_blobs          = 2
      max_file_size_mb   = 5
      allowed_mime_types = ["application/pdf", "image/png"]
    }

    custom_parameters = {
      harness = "terraform-integration"
      retries = 3
      verbose = true
    }
  }
}

resource "corax_completion_capability" "test" {
  name              = "${local.prefix}-completion"
  semantic_id       = "${local.prefix}-completion"
  project_id        = corax_project.test.id
  system_prompt     = "Extract structured data from the supplied text (rev ${var.revision})."
  completion_prompt = "Summarize the following text: {{text}}"
  output_type       = "schema"
  variables         = ["text"]

  # schema_def is a flat map of field name -> property definition. Each property
  # needs "type" and "description"; arrays add "items", enums add "enum".
  schema_def = jsonencode({
    summary = {
      type        = "string"
      description = "One-sentence summary of the input text."
    }
    sentiment = {
      # Enum properties use type = "enum", not type = "string" + enum.
      type        = "enum"
      description = "Overall sentiment of the input text."
      enum        = ["positive", "neutral", "negative"]
    }
    topics = {
      type        = "array"
      description = "Key topics mentioned in the input text."
      items = {
        type        = "string"
        description = "A single topic."
      }
    }
  })

  config = {
    temperature     = 0.0
    content_tracing = true

    data_retention = {
      type = "infinite"
    }
  }
}

resource "corax_extraction_capability" "test" {
  name          = "${local.prefix}-extraction"
  semantic_id   = "${local.prefix}-extraction"
  project_id    = corax_project.test.id
  output_type   = "text"
  system_prompt = "Extract the invoice number and total from the supplied document (rev ${var.revision})."
  is_public     = false

  config = {
    temperature     = var.revision * 0.1
    content_tracing = false
    mcp_server_ids  = [corax_mcp_server.test.id]

    data_retention = {
      type  = "timed"
      hours = 48
    }

    blob_config = {
      max_blobs          = 3
      max_file_size_mb   = 10
      allowed_mime_types = ["application/pdf"]
    }

    custom_parameters = {
      harness = "terraform-integration"
      retries = 2
    }
  }
}

# semantic_id omitted on purpose: the API generates one, which exercises the
# Optional + Computed round-trip that has caused drift on other capabilities.
resource "corax_extraction_capability" "generated_semantic_id" {
  name       = "${local.prefix}-extraction-gen"
  project_id = corax_project.test.id
}

resource "corax_speech_to_text_capability" "test" {
  count = var.enable_speech_to_text ? 1 : 0

  name          = "${local.prefix}-stt"
  semantic_id   = "${local.prefix}-stt"
  project_id    = corax_project.test.id
  output_type   = "text"
  system_prompt = "Transcribe verbatim."
}

resource "corax_api_key" "test" {
  count = var.enable_api_key ? 1 : 0

  name       = "${local.prefix}-key"
  expires_at = timeadd(timestamp(), "24h")

  lifecycle {
    ignore_changes = [expires_at]
  }
}
