// Copyright (c) Trifork

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestSchemaDefToAPI(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		input       types.String
		expectNil   bool
		expectKeys  []string
		expectError bool
	}{
		{
			name:      "null string returns nil",
			input:     types.StringNull(),
			expectNil: true,
		},
		{
			name:      "empty string returns nil",
			input:     types.StringValue(""),
			expectNil: true,
		},
		{
			name:       "valid JSON object",
			input:      types.StringValue(`{"field1":{"type":"string","description":"A field"},"field2":{"type":"integer","description":"Another field"}}`),
			expectNil:  false,
			expectKeys: []string{"field1", "field2"},
		},
		{
			name:        "invalid JSON returns error",
			input:       types.StringValue(`not valid json`),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := schemaDefToAPI(ctx, tt.input, &diags)

			if tt.expectError {
				if !diags.HasError() {
					t.Errorf("expected error but got none")
				}
				return
			}

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags)
				return
			}

			if tt.expectNil {
				if result != nil {
					t.Errorf("expected nil but got %v", result)
				}
				return
			}

			for _, key := range tt.expectKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("expected key %q in result, got keys: %v", key, result)
				}
			}
		})
	}
}

func TestSchemaDefAPIToString(t *testing.T) {
	tests := []struct {
		name       string
		input      map[string]interface{}
		expectNull bool
		expectKeys []string
	}{
		{
			name:       "nil map returns null",
			input:      nil,
			expectNull: true,
		},
		{
			name:       "empty map returns null",
			input:      map[string]interface{}{},
			expectNull: true,
		},
		{
			name: "valid map converts to JSON string",
			input: map[string]interface{}{
				"field1": map[string]interface{}{"type": "string", "description": "A field"},
				"field2": map[string]interface{}{"type": "integer", "description": "Another field"},
			},
			expectNull: false,
			expectKeys: []string{"field1", "field2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := schemaDefAPIToString(tt.input, &diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags)
				return
			}

			if tt.expectNull {
				if !result.IsNull() {
					t.Errorf("expected null but got %v", result)
				}
				return
			}

			if result.IsNull() {
				t.Errorf("expected non-null result")
				return
			}

			// Parse the JSON string and verify keys
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(result.ValueString()), &parsed); err != nil {
				t.Errorf("result is not valid JSON: %v", err)
				return
			}

			for _, key := range tt.expectKeys {
				if _, ok := parsed[key]; !ok {
					t.Errorf("expected key %q in result", key)
				}
			}
		})
	}
}

func TestAccCompletionCapabilityResource_basic(t *testing.T) {
	if os.Getenv("CORAX_API_ENDPOINT") == "" || os.Getenv("CORAX_API_KEY") == "" {
		t.Skip("Skipping acceptance test: CORAX_API_ENDPOINT or CORAX_API_KEY not set")
	}

	resourceName := "corax_completion_capability.test_basic"
	capabilityName := "tf-acc-test-completion-basic"
	systemPrompt := "You are a text completion model."
	completionPrompt := "Once upon a time, "

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCompletionCapabilityResourceBasicConfig(capabilityName, systemPrompt, completionPrompt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName),
					resource.TestCheckResourceAttr(resourceName, "system_prompt", systemPrompt),
					resource.TestCheckResourceAttr(resourceName, "completion_prompt", completionPrompt),
					resource.TestCheckResourceAttr(resourceName, "output_type", "text"),
					resource.TestCheckResourceAttr(resourceName, "type", "completion"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccCompletionCapabilityResourceBasicConfig(capabilityName+"-upd", systemPrompt+" upd", completionPrompt+"there was a..."),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "system_prompt", systemPrompt+" upd"),
					resource.TestCheckResourceAttr(resourceName, "completion_prompt", completionPrompt+"there was a..."),
				),
			},
		},
	})
}

func TestAccCompletionCapabilityResource_withSchemaOutput(t *testing.T) {
	if os.Getenv("CORAX_API_ENDPOINT") == "" || os.Getenv("CORAX_API_KEY") == "" {
		t.Skip("Skipping acceptance test: CORAX_API_ENDPOINT or CORAX_API_KEY not set")
	}

	resourceName := "corax_completion_capability.test_schema"
	capabilityName := "tf-acc-test-completion-schema"
	systemPrompt := "Extract structured data."
	completionPrompt := "User: John Doe, Age: 30, City: New York."

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with schema output type
			{
				Config: testAccCompletionCapabilityResourceSchemaOutputConfig(capabilityName, systemPrompt, completionPrompt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName),
					resource.TestCheckResourceAttr(resourceName, "output_type", "schema"),
					resource.TestCheckResourceAttr(resourceName, "variables.#", "3"), // ["User", "Age", "City"]
					// Check that schema_def is set and contains expected keys
					checkSchemaDefContainsKeys(resourceName, []string{"name", "age", "city", "is_student", "details"}),
				),
			},
			// Update: Change output_type to text and remove schema_def
			{
				Config: testAccCompletionCapabilityResourceSchemaToTextConfig(capabilityName, systemPrompt, completionPrompt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName),
					resource.TestCheckResourceAttr(resourceName, "output_type", "text"),
					// When output_type is "text", schema_def should be null
					resource.TestCheckNoResourceAttr(resourceName, "schema_def"),
				),
			},
		},
	})
}

func TestAccCompletionCapabilityResource_schemaDefFieldUpdate(t *testing.T) {
	if os.Getenv("CORAX_API_ENDPOINT") == "" || os.Getenv("CORAX_API_KEY") == "" {
		t.Skip("Skipping acceptance test: CORAX_API_ENDPOINT or CORAX_API_KEY not set")
	}

	resourceName := "corax_completion_capability.test_field_update"
	capabilityName := "tf-acc-test-completion-field-update"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with initial schema_def
			{
				Config: testAccCompletionCapabilityResourceFieldUpdateConfig(capabilityName, "OPTION_A", "OPTION_B"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName),
					resource.TestCheckResourceAttr(resourceName, "output_type", "schema"),
					checkSchemaDefContainsKeys(resourceName, []string{"status", "notes"}),
				),
			},
			// Update: Change only the enum values in the status field
			// This tests that field-level updates work correctly
			{
				Config: testAccCompletionCapabilityResourceFieldUpdateConfig(capabilityName, "OPTION_A", "OPTION_C"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName),
					resource.TestCheckResourceAttr(resourceName, "output_type", "schema"),
					checkSchemaDefContainsKeys(resourceName, []string{"status", "notes"}),
				),
			},
		},
	})
}

// checkSchemaDefContainsKeys returns a TestCheckFunc that validates schema_def
// contains expected keys. With StringAttribute, schema_def is a JSON string.
func checkSchemaDefContainsKeys(resourceName string, expectedKeys []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// With StringAttribute, schema_def is stored directly as a JSON string
		schemaDefJSON := rs.Primary.Attributes["schema_def"]
		if schemaDefJSON == "" {
			return fmt.Errorf("schema_def is empty or missing")
		}

		// Parse the JSON to check for keys
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(schemaDefJSON), &parsed); err != nil {
			return fmt.Errorf("schema_def is not valid JSON: %v (value: %s)", err, schemaDefJSON)
		}

		for _, key := range expectedKeys {
			if _, ok := parsed[key]; !ok {
				// Collect existing keys for error message
				var existingKeys []string
				for k := range parsed {
					existingKeys = append(existingKeys, k)
				}
				return fmt.Errorf("missing expected key %q in schema_def, got keys: %v", key, existingKeys)
			}
		}
		return nil
	}
}

func testAccCompletionCapabilityResourceBasicConfig(name, sysPrompt, compPrompt string) string {
	return fmt.Sprintf(`
provider "corax" {}

resource "corax_completion_capability" "test_basic" {
  name               = "%s"
  system_prompt      = "%s"
  completion_prompt  = "%s"
  output_type        = "text"
}
`, name, sysPrompt, compPrompt)
}

// testAccCompletionCapabilityResourceSchemaToTextConfig creates a config for the "test_schema" resource
// with output_type = "text". This is used to test transitioning from schema to text output.
func testAccCompletionCapabilityResourceSchemaToTextConfig(name, sysPrompt, compPrompt string) string {
	return fmt.Sprintf(`
provider "corax" {}

resource "corax_completion_capability" "test_schema" {
  name               = "%s"
  system_prompt      = "%s"
  completion_prompt  = "%s"
  output_type        = "text"
}
`, name, sysPrompt, compPrompt)
}

func testAccCompletionCapabilityResourceSchemaOutputConfig(name, sysPrompt, compPrompt string) string {
	// Using jsonencode for the entire schema_def as a single JSON string
	return fmt.Sprintf(`
provider "corax" {}

resource "corax_completion_capability" "test_schema" {
  name               = "%s"
  system_prompt      = "%s"
  completion_prompt  = "%s"
  output_type        = "schema"
  
  variables = ["User", "Age", "City"]

  schema_def = jsonencode({
    name = {
      type        = "string"
      description = "The name of the user"
    }
    age = {
      type        = "integer"
      description = "The age of the user"
    }
    city = {
      type        = "string"
      description = "The city where the user lives"
    }
    is_student = {
      type        = "boolean"
      description = "Is the user a student"
    }
    details = {
      type        = "object"
      description = "Further details"
      properties = {
        hobby      = { type = "string", description = "User's hobby" }
        occupation = { type = "string", description = "User's occupation" }
      }
    }
  })

  config = {
    temperature = 0.5
  }
}
`, name, sysPrompt, compPrompt)
}

// testAccCompletionCapabilityResourceFieldUpdateConfig creates a config for testing
// field-level updates to schema_def. The enum values can be changed between test steps.
func testAccCompletionCapabilityResourceFieldUpdateConfig(name, enumVal1, enumVal2 string) string {
	return fmt.Sprintf(`
provider "corax" {}

resource "corax_completion_capability" "test_field_update" {
  name               = "%s"
  system_prompt      = "Evaluate status"
  completion_prompt  = "Determine the status"
  output_type        = "schema"

  schema_def = jsonencode({
    status = {
      type        = "enum"
      description = "The status value"
      enum        = ["%s", "%s"]
    }
    notes = {
      type        = "string"
      description = "Optional notes"
    }
  })
}
`, name, enumVal1, enumVal2)
}
