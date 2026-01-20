// Copyright (c) Trifork

package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

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
					// Use custom check function to validate schema_def contains expected keys
					// This is more robust than exact JSON string matching which is fragile to key ordering
					checkSchemaDefContainsKeys(resourceName, []string{"name", "age", "city", "is_student", "details"}),
				),
			},
			// Update: Change output_type to text and remove schema_def
			// Note: Uses dedicated config that targets the same resource "test_schema"
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

// checkSchemaDefContainsKeys returns a TestCheckFunc that validates schema_def
// contains expected keys as valid JSON. This is more robust than exact string matching
// because JSON key order can vary.
func checkSchemaDefContainsKeys(resourceName string, expectedKeys []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		schemaDef := rs.Primary.Attributes["schema_def"]
		if schemaDef == "" {
			return fmt.Errorf("schema_def is empty")
		}

		var got map[string]interface{}
		if err := json.Unmarshal([]byte(schemaDef), &got); err != nil {
			return fmt.Errorf("schema_def is not valid JSON: %v (value: %s)", err, schemaDef)
		}

		for _, key := range expectedKeys {
			if _, ok := got[key]; !ok {
				return fmt.Errorf("missing expected key %q in schema_def, got keys: %v", key, getMapKeys(got))
			}
		}
		return nil
	}
}

// getMapKeys returns the keys of a map as a slice for error reporting.
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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
	// Using jsonencode for schema_def values for easier HCL representation.
	// The provider needs to handle these stringified JSON values if DynamicType is used this way.
	return fmt.Sprintf(`
provider "corax" {}

resource "corax_completion_capability" "test_schema" {
  name               = "%s"
  system_prompt      = "%s"
  completion_prompt  = "%s"
  output_type        = "schema"
  
  variables = ["User", "Age", "City"]

  schema_def = {
    name = jsonencode({
      type        = "string"
      description = "The name of the user"
    })
    age = jsonencode({
      type        = "integer"
      description = "The age of the user"
    })
    city = jsonencode({
      type        = "string"
      description = "The city where the user lives"
    })
    is_student = jsonencode({
      type = "boolean"
      description = "Is the user a student"
    })
    details = jsonencode({
        type = "object"
        description = "Further details"
        properties = {
            hobby = { type = "string", description = "User's hobby"}
            occupation = { type = "string", description = "User's occupation"}
        }
    })
  }

  config = {
    temperature = 0.5
  }
}
`, name, sysPrompt, compPrompt)
}
