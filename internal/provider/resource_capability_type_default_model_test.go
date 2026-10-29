// Copyright (c) Trifork

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccCapabilityTypeDefaultModelDeploymentIDEnvVar = "CORAX_TEST_DEFAULT_MODEL_DEPLOYMENT_ID"
const testAccCapabilityTypeDefaultModelDeploymentIDEnvVar2 = "CORAX_TEST_DEFAULT_MODEL_DEPLOYMENT_ID_2" // For update test

func TestCapabilityTypeDefaultModelResourceCapabilityTypeValidator(t *testing.T) {
	ctx := context.Background()
	r := NewCapabilityTypeDefaultModelResource()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema method diagnostics: %v", schemaResp.Diagnostics)
	}
	if diags := schemaResp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("invalid schema implementation: %v", diags)
	}

	attribute, ok := schemaResp.Schema.Attributes["capability_type"].(resourceschema.StringAttribute)
	if !ok {
		t.Fatal("capability_type is not a string attribute")
	}
	if len(attribute.Validators) == 0 {
		t.Fatal("capability_type has no validators")
	}

	for _, capabilityType := range []string{"chat", "completion", "embedding", "speech_to_text", "image_generation"} {
		t.Run(capabilityType, func(t *testing.T) {
			for _, attributeValidator := range attribute.Validators {
				resp := &validator.StringResponse{}
				attributeValidator.ValidateString(ctx, validator.StringRequest{ConfigValue: types.StringValue(capabilityType)}, resp)
				if resp.Diagnostics.HasError() {
					t.Fatalf("capability type %q was rejected: %v", capabilityType, resp.Diagnostics)
				}
			}
		})
	}

	for _, attributeValidator := range attribute.Validators {
		resp := &validator.StringResponse{}
		attributeValidator.ValidateString(ctx, validator.StringRequest{ConfigValue: types.StringValue("not_a_capability_type")}, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("invalid capability type was accepted")
		}
	}
}

func TestAccCapabilityTypeDefaultModelResource_basic(t *testing.T) {
	if os.Getenv("CORAX_API_ENDPOINT") == "" || os.Getenv("CORAX_API_KEY") == "" {
		t.Skip("Skipping acceptance test: CORAX_API_ENDPOINT or CORAX_API_KEY not set")
	}
	testModelDeploymentID := os.Getenv(testAccCapabilityTypeDefaultModelDeploymentIDEnvVar)
	if testModelDeploymentID == "" {
		t.Skipf("Skipping acceptance test: %s must be set with a valid Model Deployment UUID", testAccCapabilityTypeDefaultModelDeploymentIDEnvVar)
	}
	testModelDeploymentID2 := os.Getenv(testAccCapabilityTypeDefaultModelDeploymentIDEnvVar2)
	if testModelDeploymentID2 == "" {
		t.Skipf("Skipping acceptance test: %s must be set with a second valid Model Deployment UUID for update test", testAccCapabilityTypeDefaultModelDeploymentIDEnvVar2)
	}
	if testModelDeploymentID == testModelDeploymentID2 {
		t.Skipf("Skipping acceptance test: %s and %s must be different for update test", testAccCapabilityTypeDefaultModelDeploymentIDEnvVar, testAccCapabilityTypeDefaultModelDeploymentIDEnvVar2)
	}

	resourceName := "corax_capability_type_default_model.chat_default"
	capabilityType := "chat" // Testing with "chat" type

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCapabilityTypeDefaultModelConfig(capabilityType, testModelDeploymentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "capability_type", capabilityType),
					resource.TestCheckResourceAttr(resourceName, "default_model_deployment_id", testModelDeploymentID),
					resource.TestCheckResourceAttrSet(resourceName, "name"), // Name is read-only from API
				),
			},
			// ImportState testing
			// The ID for this resource is the capability_type itself.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     capabilityType, // Import using the capability_type value
				ImportStateVerify: true,
			},
			// Update and Read testing (change the default_model_deployment_id)
			{
				Config: testAccCapabilityTypeDefaultModelConfig(capabilityType, testModelDeploymentID2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "capability_type", capabilityType),
					resource.TestCheckResourceAttr(resourceName, "default_model_deployment_id", testModelDeploymentID2),
				),
			},
			// Delete testing:
			// The resource's Delete method currently issues a warning and is a no-op regarding API calls.
			// So, after "deletion" by Terraform, the API will still have the last set default.
			// A Read after delete would find it.
			// To properly test "unsetting", the API would need to support it, and Delete method updated.
			// For now, we just test that Terraform removes it from its state.
		},
	})
}

func testAccCapabilityTypeDefaultModelConfig(capabilityType, modelDeploymentID string) string {
	return fmt.Sprintf(`
provider "corax" {
  # api_endpoint = "..." 
  # api_key      = "..."
}

resource "corax_capability_type_default_model" "chat_default" {
  capability_type              = "%s"
  default_model_deployment_id  = "%s"
}
`, capabilityType, modelDeploymentID)
}

// testAccPreCheck is defined in provider_test.go
