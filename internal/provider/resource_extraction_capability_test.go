// Copyright (c) Trifork

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	api "terraform-provider-corax/internal/generated"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestMapExtractionCapabilityRepresentationToModel verifies that a capability
// read back from the API (GET/PUT shape) is mapped onto the Terraform model,
// including the fields that live inside the flattened `configuration` map.
func TestMapExtractionCapabilityRepresentationToModel(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	apiCap := api.NewCapabilityRepresentation(
		"Invoice Extractor", // name
		"extraction",        // type
		"invoice-extractor", // semanticId
		"cap-1",             // id
		"user-1",            // createdBy
		"user-2",            // updatedBy
		now,                 // createdAt
		now,                 // updatedAt
		"owner-1",           // owner
		map[string]interface{}{},
		map[string]interface{}{"type": "text"},
		map[string]interface{}{
			"system_prompt": "Extract the invoice fields.",
			"output_type":   "text",
		},
		true, // isDefaultVersion
	)
	apiCap.SetIsPublic(true)
	apiCap.SetModelId("11111111-1111-1111-1111-111111111111")

	model := &ExtractionCapabilityResourceModel{}
	var diags diag.Diagnostics

	mapExtractionCapabilityRepresentationToModel(apiCap, model, &diags, ctx)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags)
	}

	if got, want := model.ID.ValueString(), "cap-1"; got != want {
		t.Errorf("ID = %q, want %q", got, want)
	}
	if got, want := model.Name.ValueString(), "Invoice Extractor"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if got, want := model.Type.ValueString(), "extraction"; got != want {
		t.Errorf("Type = %q, want %q", got, want)
	}
	if !model.IsPublic.ValueBool() {
		t.Error("IsPublic = false, want true")
	}
	if got, want := model.ModelID.ValueString(), "11111111-1111-1111-1111-111111111111"; got != want {
		t.Errorf("ModelID = %q, want %q", got, want)
	}
	if !model.ModelPoolID.IsNull() {
		t.Errorf("ModelPoolID = %v, want null", model.ModelPoolID)
	}
	if got, want := model.SemanticID.ValueString(), "invoice-extractor"; got != want {
		t.Errorf("SemanticID = %q, want %q", got, want)
	}
	if got, want := model.SystemPrompt.ValueString(), "Extract the invoice fields."; got != want {
		t.Errorf("SystemPrompt = %q, want %q", got, want)
	}
	if got, want := model.OutputType.ValueString(), "text"; got != want {
		t.Errorf("OutputType = %q, want %q", got, want)
	}
	if got, want := model.Owner.ValueString(), "owner-1"; got != want {
		t.Errorf("Owner = %q, want %q", got, want)
	}
	if got, want := model.CreatedBy.ValueString(), "user-1"; got != want {
		t.Errorf("CreatedBy = %q, want %q", got, want)
	}
	if got, want := model.UpdatedBy.ValueString(), "user-2"; got != want {
		t.Errorf("UpdatedBy = %q, want %q", got, want)
	}
	if got, want := model.CreatedAt.ValueString(), now.Format(time.RFC3339); got != want {
		t.Errorf("CreatedAt = %q, want %q", got, want)
	}
}

// TestMapExtractionCapabilityRepresentationToModel_NullOptionals verifies that
// absent optional fields become Terraform nulls rather than empty strings, so
// that Terraform does not report spurious drift.
func TestMapExtractionCapabilityRepresentationToModel_NullOptionals(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	apiCap := api.NewCapabilityRepresentation(
		"Minimal", "extraction", "minimal-extractor", "cap-2",
		"user-1", "user-1", now, now, "owner-1",
		map[string]interface{}{},
		map[string]interface{}{},
		map[string]interface{}{},
		true,
	)

	model := &ExtractionCapabilityResourceModel{}
	var diags diag.Diagnostics

	mapExtractionCapabilityRepresentationToModel(apiCap, model, &diags, ctx)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags)
	}
	if !model.ModelID.IsNull() {
		t.Errorf("ModelID = %v, want null", model.ModelID)
	}
	if !model.ProjectID.IsNull() {
		t.Errorf("ProjectID = %v, want null", model.ProjectID)
	}
	if !model.SystemPrompt.IsNull() {
		t.Errorf("SystemPrompt = %v, want null", model.SystemPrompt)
	}
	// output_type is absent from configuration; it must fall back to the default.
	if got, want := model.OutputType.ValueString(), "text"; got != want {
		t.Errorf("OutputType = %q, want %q", got, want)
	}
}

// TestMapExtractionCapabilityCreateResponseToModel verifies the POST response
// shape (a typed ExtractionCapability) maps onto the Terraform model.
func TestMapExtractionCapabilityCreateResponseToModel(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	apiCap := api.NewExtractionCapability(
		"Invoice Extractor", // name
		"cap-3",             // id
		"user-1",            // createdBy
		"user-1",            // updatedBy
		now,                 // createdAt
		now,                 // updatedAt
		"owner-1",           // owner
	)
	apiCap.SetIsPublic(false)
	apiCap.SetSemanticId("invoice-extractor")
	apiCap.SetSystemPrompt("Extract the invoice fields.")
	apiCap.SetModelPoolId("22222222-2222-2222-2222-222222222222")
	apiCap.SetProjectId("33333333-3333-3333-3333-333333333333")

	model := &ExtractionCapabilityResourceModel{}
	var diags diag.Diagnostics

	mapExtractionCapabilityCreateResponseToModel(apiCap, model, &diags, ctx)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags)
	}
	if got, want := model.ID.ValueString(), "cap-3"; got != want {
		t.Errorf("ID = %q, want %q", got, want)
	}
	// The generated model defaults `type` to "extraction".
	if got, want := model.Type.ValueString(), "extraction"; got != want {
		t.Errorf("Type = %q, want %q", got, want)
	}
	if got, want := model.OutputType.ValueString(), "text"; got != want {
		t.Errorf("OutputType = %q, want %q", got, want)
	}
	if got, want := model.SemanticID.ValueString(), "invoice-extractor"; got != want {
		t.Errorf("SemanticID = %q, want %q", got, want)
	}
	if got, want := model.ModelPoolID.ValueString(), "22222222-2222-2222-2222-222222222222"; got != want {
		t.Errorf("ModelPoolID = %q, want %q", got, want)
	}
	if got, want := model.ProjectID.ValueString(), "33333333-3333-3333-3333-333333333333"; got != want {
		t.Errorf("ProjectID = %q, want %q", got, want)
	}
	if !model.ModelID.IsNull() {
		t.Errorf("ModelID = %v, want null", model.ModelID)
	}
	if model.IsPublic.ValueBool() {
		t.Error("IsPublic = true, want false")
	}
}

// TestExtractionCapabilityResourceSchema verifies the resource schema is valid
// and exposes the attributes the API supports.
func TestExtractionCapabilityResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewExtractionCapabilityResource()

	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema method diagnostics: %v", schemaResp.Diagnostics)
	}
	if diags := schemaResp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("invalid schema implementation: %v", diags)
	}

	for _, attr := range []string{
		"id", "name", "is_public", "model_id", "model_pool_id", "project_id",
		"semantic_id", "system_prompt", "output_type", "config", "type",
		"owner", "created_at", "updated_at", "created_by", "updated_by",
	} {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema is missing attribute %q", attr)
		}
	}
}

func TestAccExtractionCapabilityResource_basic(t *testing.T) {
	if os.Getenv("CORAX_API_ENDPOINT") == "" || os.Getenv("CORAX_API_KEY") == "" {
		t.Skip("Skipping acceptance test: CORAX_API_ENDPOINT or CORAX_API_KEY not set")
	}

	resourceName := "corax_extraction_capability.test_basic"
	capabilityName := "tf-acc-test-extraction-basic"
	systemPrompt := "Extract the key fields from the document."

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccExtractionCapabilityResourceBasicConfig(capabilityName, systemPrompt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName),
					resource.TestCheckResourceAttr(resourceName, "system_prompt", systemPrompt),
					resource.TestCheckResourceAttr(resourceName, "output_type", "text"),
					resource.TestCheckResourceAttr(resourceName, "type", "extraction"),
					resource.TestCheckResourceAttr(resourceName, "is_public", "false"),
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
				Config: testAccExtractionCapabilityResourceBasicConfig(capabilityName+"-upd", systemPrompt+" Be concise."),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "system_prompt", systemPrompt+" Be concise."),
				),
			},
		},
	})
}

func TestAccExtractionCapabilityResource_withConfig(t *testing.T) {
	if os.Getenv("CORAX_API_ENDPOINT") == "" || os.Getenv("CORAX_API_KEY") == "" {
		t.Skip("Skipping acceptance test: CORAX_API_ENDPOINT or CORAX_API_KEY not set")
	}

	resourceName := "corax_extraction_capability.test_config"
	capabilityName := "tf-acc-test-extraction-config"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExtractionCapabilityResourceConfigWithConfig(capabilityName, 0.2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName),
					resource.TestCheckResourceAttr(resourceName, "semantic_id", "tf-acc-extraction-config"),
					resource.TestCheckResourceAttr(resourceName, "config.temperature", "0.2"),
				),
			},
			{
				Config: testAccExtractionCapabilityResourceConfigWithConfig(capabilityName, 0.7),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "config.temperature", "0.7"),
				),
			},
		},
	})
}

func testAccExtractionCapabilityResourceBasicConfig(name, sysPrompt string) string {
	return fmt.Sprintf(`
provider "corax" {}

resource "corax_extraction_capability" "test_basic" {
  name          = %[1]q
  system_prompt = %[2]q
  output_type   = "text"
}
`, name, sysPrompt)
}

func testAccExtractionCapabilityResourceConfigWithConfig(name string, temperature float64) string {
	return fmt.Sprintf(`
provider "corax" {}

resource "corax_extraction_capability" "test_config" {
  name          = %[1]q
  semantic_id   = "tf-acc-extraction-config"
  system_prompt = "Extract structured data from the document."

  config = {
    temperature = %[2]g
  }
}
`, name, temperature)
}

// newFakeCapabilityAPI returns an httptest server that implements the subset of
// /v1/capabilities the extraction capability resource uses, backed by an
// in-memory store. It lets the full Terraform lifecycle (create, read, update,
// import, destroy) run against real HTTP without requiring a live Corax API.
func newFakeCapabilityAPI(t *testing.T) *httptest.Server {
	t.Helper()

	var (
		mu    sync.Mutex
		store = map[string]map[string]interface{}{}
		next  = 0
	)

	// representation renders a stored capability in the CapabilityRepresentation
	// shape returned by GET and PUT, where system_prompt and output_type are
	// flattened into `configuration`.
	representation := func(stored map[string]interface{}) map[string]interface{} {
		rep := map[string]interface{}{}
		for k, v := range stored {
			rep[k] = v
		}
		configuration := map[string]interface{}{}
		if v, ok := stored["system_prompt"]; ok && v != nil {
			configuration["system_prompt"] = v
		}
		if v, ok := stored["output_type"]; ok && v != nil {
			configuration["output_type"] = v
		}
		rep["configuration"] = configuration
		if _, ok := rep["config"]; !ok {
			// The API always materialises the capability config, applying
			// server-side defaults when the client omits it.
			rep["config"] = map[string]interface{}{"content_tracing": true}
		}
		rep["input"] = map[string]interface{}{}
		rep["output"] = map[string]interface{}{}
		rep["is_default_version"] = true
		if _, ok := rep["semantic_id"]; !ok || rep["semantic_id"] == nil {
			rep["semantic_id"] = "generated-semantic-id"
		}
		return rep
	}

	// writeJSON sets Content-Type before writing the status line: the generated
	// client picks its decoder from that header, and a response without it
	// fails to decode.
	writeJSON := func(w http.ResponseWriter, status int, body interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Errorf("failed to encode fake API response: %v", err)
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/capabilities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if got := body["type"]; got != "extraction" {
			t.Errorf("create payload type = %v, want \"extraction\"", got)
		}

		mu.Lock()
		defer mu.Unlock()
		next++
		id := fmt.Sprintf("00000000-0000-0000-0000-%012d", next)
		now := time.Now().UTC().Format(time.RFC3339)
		body["id"] = id
		body["owner"] = "fake-owner"
		body["created_by"] = "fake-user"
		body["updated_by"] = "fake-user"
		body["created_at"] = now
		body["updated_at"] = now
		if _, ok := body["semantic_id"]; !ok {
			body["semantic_id"] = "generated-semantic-id"
		}
		if _, ok := body["config"]; !ok {
			body["config"] = map[string]interface{}{"content_tracing": true}
		}
		store[id] = body

		writeJSON(w, http.StatusCreated, body)
	})

	mux.HandleFunc("/v1/capabilities/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/capabilities/")

		mu.Lock()
		defer mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			stored, ok := store[id]
			if !ok {
				http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, representation(stored))
		case http.MethodPut:
			if _, ok := store[id]; !ok {
				http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
				return
			}
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if got := body["type"]; got != "extraction" {
				t.Errorf("update payload type = %v, want \"extraction\"", got)
			}
			existing := store[id]
			body["id"] = id
			body["owner"] = existing["owner"]
			body["created_by"] = existing["created_by"]
			body["created_at"] = existing["created_at"]
			body["updated_by"] = "fake-user"
			body["updated_at"] = time.Now().UTC().Format(time.RFC3339)
			store[id] = body
			writeJSON(w, http.StatusOK, representation(body))
		case http.MethodDelete:
			if _, ok := store[id]; !ok {
				http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
				return
			}
			delete(store, id)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// TestAccExtractionCapabilityResource_againstFakeAPI drives the whole resource
// lifecycle through Terraform against an in-memory Corax API, so it runs in CI
// without credentials.
func TestAccExtractionCapabilityResource_againstFakeAPI(t *testing.T) {
	server := newFakeCapabilityAPI(t)

	t.Setenv("TF_ACC", "1")
	t.Setenv("CORAX_API_ENDPOINT", server.URL)
	t.Setenv("CORAX_API_KEY", "fake-api-key")

	resourceName := "corax_extraction_capability.test_basic"
	capabilityName := "tf-unit-test-extraction"
	systemPrompt := "Extract the key fields from the document."

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccExtractionCapabilityResourceBasicConfig(capabilityName, systemPrompt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName),
					resource.TestCheckResourceAttr(resourceName, "system_prompt", systemPrompt),
					resource.TestCheckResourceAttr(resourceName, "output_type", "text"),
					resource.TestCheckResourceAttr(resourceName, "type", "extraction"),
					resource.TestCheckResourceAttr(resourceName, "is_public", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
				),
			},
			// Import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update in place
			{
				Config: testAccExtractionCapabilityResourceBasicConfig(capabilityName+"-upd", systemPrompt+" Be concise."),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", capabilityName+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "system_prompt", systemPrompt+" Be concise."),
				),
			},
		},
	})
}

// TestAccExtractionCapabilityResource_configAgainstFakeAPI covers the explicit
// config and semantic_id path, including an in-place config update.
func TestAccExtractionCapabilityResource_configAgainstFakeAPI(t *testing.T) {
	server := newFakeCapabilityAPI(t)

	t.Setenv("TF_ACC", "1")
	t.Setenv("CORAX_API_ENDPOINT", server.URL)
	t.Setenv("CORAX_API_KEY", "fake-api-key")

	resourceName := "corax_extraction_capability.test_config"
	capabilityName := "tf-unit-test-extraction-config"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExtractionCapabilityResourceConfigWithConfig(capabilityName, 0.2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "semantic_id", "tf-acc-extraction-config"),
					resource.TestCheckResourceAttr(resourceName, "config.temperature", "0.2"),
				),
			},
			{
				Config: testAccExtractionCapabilityResourceConfigWithConfig(capabilityName, 0.7),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "config.temperature", "0.7"),
				),
			},
		},
	})
}
