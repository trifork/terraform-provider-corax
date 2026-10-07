# Copyright (c) Trifork

resource "corax_extraction_capability" "invoice" {
  name          = "Invoice Extractor"
  semantic_id   = "invoice-extractor"
  system_prompt = "Extract the invoice number, total amount and due date from the document."

  # "text" (default) returns the extracted markdown as a string; "schema" also
  # stores the extracted images and returns {markdown, file_ids, files}.
  output_type = "schema"

  # Either model_id (a specific model deployment) or model_pool_id (load
  # balanced across a pool). The two are mutually exclusive; when both are
  # omitted the default model for the extraction capability type is used.
  # model_id = "11111111-1111-1111-1111-111111111111"

  config = {
    temperature = 0.2
  }
}
