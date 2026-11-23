# Copyright (c) Trifork

# Spread traffic 3:1 across two deployments, falling back to the other member
# when one fails.
resource "corax_model_pool" "gpt" {
  name          = "gpt-pool"
  fallback_mode = "try_next"

  # Keyed by model deployment ID. {} takes the defaults
  # (weight = 1, priority = 1, is_active = true).
  members = {
    (corax_model_deployment.primary.id)   = { weight = 3 }
    (corax_model_deployment.secondary.id) = {}
  }
}

resource "corax_completion_capability" "summary" {
  name              = "Summary"
  model_pool_id     = corax_model_pool.gpt.id # mutually exclusive with model_id
  system_prompt     = "You summarize text."
  completion_prompt = "Summarize: {{text}}"
  variables         = ["text"]
  output_type       = "text"
}
