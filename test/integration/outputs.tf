# Copyright (c) Trifork

output "project_id" {
  value = corax_project.test.id
}

output "mcp_server" {
  value = {
    id   = corax_mcp_server.test.id
    slug = corax_mcp_server.test.slug
  }
}

output "chat_capability" {
  value = {
    id          = corax_chat_capability.test.id
    semantic_id = corax_chat_capability.test.semantic_id
    model_id    = corax_chat_capability.test.model_id
    type        = corax_chat_capability.test.type
  }
}

output "completion_capability" {
  value = {
    id          = corax_completion_capability.test.id
    semantic_id = corax_completion_capability.test.semantic_id
    type        = corax_completion_capability.test.type
  }
}

output "speech_to_text_capability_id" {
  value = var.enable_speech_to_text ? corax_speech_to_text_capability.test[0].id : null
}

output "extraction_capability" {
  value = {
    id          = corax_extraction_capability.test.id
    semantic_id = corax_extraction_capability.test.semantic_id
    output_type = corax_extraction_capability.test.output_type
    type        = corax_extraction_capability.test.type
  }
}

# The API-generated semantic_id, to confirm it is populated and stable.
output "extraction_capability_generated_semantic_id" {
  value = corax_extraction_capability.generated_semantic_id.semantic_id
}
