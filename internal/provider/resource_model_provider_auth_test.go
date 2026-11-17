// Copyright (c) Trifork

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-corax/internal/coraxclient"
)

func requireModelProviderDiagnostics(t *testing.T, diagnostics diag.Diagnostics) {
	t.Helper()
	if diagnostics.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diagnostics)
	}
}

func testModelProviderMap(t *testing.T, configuration map[string]string) types.Map {
	t.Helper()
	result, diagnostics := types.MapValueFrom(context.Background(), types.StringType, configuration)
	requireModelProviderDiagnostics(t, diagnostics)
	return result
}

func testModelProviderState(t *testing.T, state tfsdk.State) ModelProviderResourceModel {
	t.Helper()
	var model ModelProviderResourceModel
	requireModelProviderDiagnostics(t, state.Get(context.Background(), &model))
	return model
}

func TestModelProviderCredentialLifecycle(t *testing.T) {
	for _, tc := range []struct{ providerType, secret string }{
		{"azure_ai_service_principal", "client_secret"},
		{"azure_service_principal", "client_secret"},
		{"azure_ai_service_principal_certificate", "certificate_password"},
		{"azure_service_principal_certificate", "certificate_password"},
		{"azure_ai", "api_key"},
	} {
		t.Run(tc.providerType, func(t *testing.T) {
			ctx := context.Background()
			requests := make(chan coraxclient.ModelProviderUpdate, 3)
			var mu sync.Mutex
			stored := coraxclient.ModelProvider{
				ID:        "00000000-0000-0000-0000-000000000001",
				CreatedAt: "2026-01-01T00:00:00Z", CreatedBy: "terraform-test",
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				switch req.Method {
				case http.MethodPost, http.MethodPut:
					var body coraxclient.ModelProviderUpdate
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Errorf("Decode request: %v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					requests <- body
					changedType := stored.ProviderType != body.ProviderType
					stored.Name, stored.ProviderType = body.Name, body.ProviderType
					if req.Method == http.MethodPost || changedType || stored.Configuration == nil {
						stored.Configuration = make(map[string]string)
					}
					for key, value := range body.Configuration {
						stored.Configuration[key] = value
					}
				case http.MethodGet:
				default:
					t.Errorf("Unexpected method: %s", req.Method)
				}
				response := stored
				response.Configuration = make(map[string]string)
				for key, value := range stored.Configuration {
					if key == tc.secret && value != "" {
						response.Configuration[key] = "te...cret"
					} else {
						response.Configuration[key] = value
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(response); err != nil {
					t.Errorf("Encode response: %v", err)
				}
			}))
			t.Cleanup(server.Close)
			client, err := coraxclient.NewClient(server.URL, "test-api-key")
			if err != nil {
				t.Fatal(err)
			}
			subject := ModelProviderResource{client: client}
			var schemaResponse resource.SchemaResponse
			subject.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			configuration := map[string]string{
				"api_endpoint": "https://example.services.ai.azure.com", "tenant_id": "test-tenant", "client_id": "test-client",
				tc.secret: "test-original-secret",
			}
			model := ModelProviderResourceModel{
				ID: types.StringUnknown(), Name: types.StringValue("Test connection"), ProviderType: types.StringValue(tc.providerType),
				Configuration: testModelProviderMap(t, configuration),
			}
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			requireModelProviderDiagnostics(t, plan.Set(ctx, &model))
			created := resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
			subject.Create(ctx, resource.CreateRequest{Plan: plan}, &created)
			requireModelProviderDiagnostics(t, created.Diagnostics)
			createBody := <-requests
			if !reflect.DeepEqual(createBody.Configuration, configuration) {
				t.Fatal("Create request changed the planned credentials")
			}
			if !testModelProviderState(t, created.State).Configuration.Equal(model.Configuration) {
				t.Fatal("Create replaced a credential with its display mask")
			}
			refreshed := resource.ReadResponse{State: created.State}
			subject.Read(ctx, resource.ReadRequest{State: created.State}, &refreshed)
			requireModelProviderDiagnostics(t, refreshed.Diagnostics)
			current := testModelProviderState(t, refreshed.State)
			if !current.Configuration.Equal(model.Configuration) {
				t.Fatal("Refresh replaced a credential with its display mask")
			}

			configuration[tc.secret] = "test-rotated-secret"
			current.Configuration = testModelProviderMap(t, configuration)
			requireModelProviderDiagnostics(t, plan.Set(ctx, &current))
			updated := resource.UpdateResponse{State: refreshed.State}
			subject.Update(ctx, resource.UpdateRequest{Plan: plan, State: refreshed.State}, &updated)
			requireModelProviderDiagnostics(t, updated.Diagnostics)
			updateBody := <-requests
			if updateBody.Configuration[tc.secret] != configuration[tc.secret] {
				t.Fatal("Rotation did not send the new credential")
			}
			reread := resource.ReadResponse{State: updated.State}
			subject.Read(ctx, resource.ReadRequest{State: updated.State}, &reread)
			requireModelProviderDiagnostics(t, reread.Diagnostics)
			if !testModelProviderState(t, reread.State).Configuration.Equal(current.Configuration) {
				t.Fatal("Refresh lost the rotated credential")
			}

			if tc.secret != certificatePasswordConfigurationKey {
				return
			}

			switched := current
			switched.ProviderType = types.StringValue("azure_ai_service_principal")
			switched.Configuration = testModelProviderMap(t, map[string]string{"api_endpoint": "https://example.services.ai.azure.com", "tenant_id": "test-tenant", "client_id": "test-client", "client_secret": "test-new-secret"})
			requireModelProviderDiagnostics(t, plan.Set(ctx, &switched))
			switchedResponse := resource.UpdateResponse{State: reread.State}
			subject.Update(ctx, resource.UpdateRequest{Plan: plan, State: reread.State}, &switchedResponse)
			requireModelProviderDiagnostics(t, switchedResponse.Diagnostics)
			switchedBody := <-requests
			if _, copied := switchedBody.Configuration[certificatePasswordConfigurationKey]; copied {
				t.Fatal("A provider type change must not add the old certificate password field")
			}
			delete(configuration, tc.secret)
			current.Configuration = testModelProviderMap(t, configuration)
			requireModelProviderDiagnostics(t, plan.Set(ctx, &current))
			cleared := resource.UpdateResponse{State: reread.State}
			subject.Update(ctx, resource.UpdateRequest{Plan: plan, State: reread.State}, &cleared)
			requireModelProviderDiagnostics(t, cleared.Diagnostics)
			clearBody := <-requests
			password, explicit := clearBody.Configuration[tc.secret]
			if !explicit || password != "" {
				t.Fatal("Removing an optional password must explicitly clear it in Corax")
			}
			readCleared := resource.ReadResponse{State: cleared.State}
			subject.Read(ctx, resource.ReadRequest{State: cleared.State}, &readCleared)
			requireModelProviderDiagnostics(t, readCleared.Diagnostics)
			if !testModelProviderState(t, readCleared.State).Configuration.Equal(current.Configuration) {
				t.Fatal("Refresh reintroduced an omitted, cleared password")
			}
		})
	}
}

func TestModelProviderSecretResponseReconciliation(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		priorType, remoteType   string
		prior, remote, expected map[string]string
	}{
		{"legacy API-key mask", "azure_ai", "azure_ai", map[string]string{"api_key": "sk-legacy-secret"}, map[string]string{"api_key": "sk-...cret"}, map[string]string{"api_key": "sk-legacy-secret"}},
		{"legacy API-key omitted", "azure_ai", "azure_ai", map[string]string{"api_key": "sk-legacy-secret"}, map[string]string{}, map[string]string{"api_key": "sk-legacy-secret"}},
		{"short secret mask", "azure_ai_service_principal", "azure_ai_service_principal", map[string]string{"client_secret": "short"}, map[string]string{"client_secret": "********"}, map[string]string{"client_secret": "short"}},
		{"unicode secret mask", "azure_ai_service_principal_certificate", "azure_ai_service_principal_certificate", map[string]string{"certificate_password": "æø-long-密碼abcd"}, map[string]string{"certificate_password": "æø...abcd"}, map[string]string{"certificate_password": "æø-long-密碼abcd"}},
		{"nonsecret drift", "azure_ai_service_principal", "azure_ai_service_principal", map[string]string{"api_endpoint": "https://before.invalid"}, map[string]string{"api_endpoint": "https://after.invalid"}, map[string]string{"api_endpoint": "https://after.invalid"}},
		{"explicit empty password", "azure_ai_service_principal_certificate", "azure_ai_service_principal_certificate", map[string]string{"certificate_password": ""}, map[string]string{"certificate_password": ""}, map[string]string{"certificate_password": ""}},
		{"unencrypted certificate without password", "azure_ai_service_principal_certificate", "azure_ai_service_principal_certificate", map[string]string{}, map[string]string{}, map[string]string{}},
		{"import cannot recover secrets", "", "azure_ai_service_principal", nil, map[string]string{"client_secret": "ab...cdef"}, map[string]string{"client_secret": "ab...cdef"}},
		{"provider type changed", "azure_ai_service_principal", "azure_service_principal", map[string]string{"client_secret": "prior-secret"}, map[string]string{"client_secret": "ab...cdef"}, map[string]string{"client_secret": "ab...cdef"}},
		{"remote clearing is not masked", "azure_ai_service_principal_certificate", "azure_ai_service_principal_certificate", map[string]string{"certificate_password": "prior-password"}, map[string]string{"certificate_password": ""}, map[string]string{"certificate_password": ""}},
		{"different actual secret", "azure_ai_service_principal", "azure_ai_service_principal", map[string]string{"client_secret": "prior-secret"}, map[string]string{"client_secret": "remote-new-secret"}, map[string]string{"client_secret": "remote-new-secret"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := ModelProviderResourceModel{ProviderType: types.StringValue(tc.priorType), Configuration: testModelProviderMap(t, tc.prior)}
			remote := coraxclient.ModelProvider{ProviderType: tc.remoteType, Configuration: tc.remote}
			var diagnostics diag.Diagnostics
			mapAPIModelProviderToResourceModel(context.Background(), &remote, &model, &diagnostics)
			requireModelProviderDiagnostics(t, diagnostics)
			expected := testModelProviderMap(t, tc.expected)
			if !model.Configuration.Equal(expected) {
				t.Fatal("Unexpected reconciled configuration")
			}
		})
	}
}
