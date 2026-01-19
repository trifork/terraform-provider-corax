// Copyright (c) Trifork

package coraxclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Helper to create a test server and client.
func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(handler)
	client, err := NewClient(server.URL, "test-api-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return server, client
}

// TestNewClient tests client creation.
func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		apiKey    string
		wantError bool
	}{
		{
			name:      "valid inputs",
			baseURL:   "https://api.example.com",
			apiKey:    "test-key",
			wantError: false,
		},
		{
			name:      "empty base URL",
			baseURL:   "",
			apiKey:    "test-key",
			wantError: true,
		},
		{
			name:      "empty API key",
			baseURL:   "https://api.example.com",
			apiKey:    "",
			wantError: true,
		},
		{
			name:      "whitespace base URL",
			baseURL:   "   ",
			apiKey:    "test-key",
			wantError: true,
		},
		{
			name:      "invalid URL",
			baseURL:   "not-a-valid-url",
			apiKey:    "test-key",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, tt.apiKey)
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Error("Expected client but got nil")
				}
			}
		})
	}
}

// TestCreateAPIKey tests the CreateAPIKey method.
func TestCreateAPIKey(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/v1/api-keys" {
				t.Errorf("Expected /v1/api-keys, got %s", r.URL.Path)
			}
			if r.Header.Get("X-API-Key") != "test-api-key" {
				t.Errorf("Expected X-API-Key header, got %s", r.Header.Get("X-API-Key"))
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			// Verify request body
			var reqBody ApiKeyCreate
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}
			if reqBody.Name != "test-key" {
				t.Errorf("Expected name 'test-key', got %s", reqBody.Name)
			}

			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ApiKey{
				ID:        "key-123",
				Name:      "test-key",
				Key:       "secret-key-value",
				Prefix:    "corax-sk-",
				IsActive:  true,
				CreatedBy: "user-1",
				CreatedAt: "2024-01-01T00:00:00Z",
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.CreateAPIKey(context.Background(), ApiKeyCreate{
			Name:      "test-key",
			ExpiresAt: "2025-01-01T00:00:00Z",
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "key-123" {
			t.Errorf("Expected ID 'key-123', got %s", result.ID)
		}
		if result.Key != "secret-key-value" {
			t.Errorf("Expected Key 'secret-key-value', got %s", result.Key)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"detail": "Validation error"}`))
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		_, err := client.CreateAPIKey(context.Background(), ApiKeyCreate{})

		if err == nil {
			t.Fatal("Expected error but got nil")
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("Expected APIError, got %T", err)
		}
		if apiErr.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("Expected status 422, got %d", apiErr.StatusCode)
		}
	})
}

// TestGetAPIKey tests the GetAPIKey method.
func TestGetAPIKey(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/v1/api-keys/key-123" {
				t.Errorf("Expected /v1/api-keys/key-123, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ApiKey{
				ID:        "key-123",
				Name:      "test-key",
				Key:       "",
				Prefix:    "corax-sk-",
				IsActive:  true,
				CreatedBy: "user-1",
				CreatedAt: "2024-01-01T00:00:00Z",
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.GetAPIKey(context.Background(), "key-123")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "key-123" {
			t.Errorf("Expected ID 'key-123', got %s", result.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail": "Not found"}`))
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		_, err := client.GetAPIKey(context.Background(), "nonexistent")

		if err == nil {
			t.Fatal("Expected error but got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("empty key ID", func(t *testing.T) {
		_, client := setupTestServer(t, nil)

		_, err := client.GetAPIKey(context.Background(), "")

		if err == nil {
			t.Fatal("Expected error but got nil")
		}
	})
}

// TestDeleteAPIKey tests the DeleteAPIKey method.
func TestDeleteAPIKey(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("Expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/v1/api-keys/key-123" {
				t.Errorf("Expected /v1/api-keys/key-123, got %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusNoContent)
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		err := client.DeleteAPIKey(context.Background(), "key-123")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		err := client.DeleteAPIKey(context.Background(), "nonexistent")

		if err == nil {
			t.Fatal("Expected error but got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

// TestCreateProject tests the CreateProject method.
func TestCreateProject(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/v1/projects" {
				t.Errorf("Expected /v1/projects, got %s", r.URL.Path)
			}

			var reqBody ProjectCreate
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}
			if reqBody.Name != "test-project" {
				t.Errorf("Expected name 'test-project', got %s", reqBody.Name)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Include all required fields for the generated client
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":               "proj-123",
				"name":             "test-project",
				"is_public":        false,
				"created_by":       "user-1",
				"created_at":       "2024-01-01T00:00:00Z",
				"updated_by":       "user-1",
				"updated_at":       "2024-01-01T00:00:00Z",
				"owner":            "user-1",
				"collection_count": 0,
				"capability_count": 0,
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.CreateProject(context.Background(), ProjectCreate{
			Name: "test-project",
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "proj-123" {
			t.Errorf("Expected ID 'proj-123', got %s", result.ID)
		}
	})
}

// TestGetProject tests the GetProject method.
func TestGetProject(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/v1/projects/proj-123" {
				t.Errorf("Expected /v1/projects/proj-123, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Include all required fields for the generated client
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":               "proj-123",
				"name":             "test-project",
				"is_public":        false,
				"created_by":       "user-1",
				"created_at":       "2024-01-01T00:00:00Z",
				"updated_by":       "user-1",
				"updated_at":       "2024-01-01T00:00:00Z",
				"owner":            "user-1",
				"collection_count": 0,
				"capability_count": 0,
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.GetProject(context.Background(), "proj-123")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "proj-123" {
			t.Errorf("Expected ID 'proj-123', got %s", result.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		_, err := client.GetProject(context.Background(), "nonexistent")

		if err == nil {
			t.Fatal("Expected error but got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

// TestUpdateProject tests the UpdateProject method.
func TestUpdateProject(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("Expected PUT, got %s", r.Method)
			}
			if r.URL.Path != "/v1/projects/proj-123" {
				t.Errorf("Expected /v1/projects/proj-123, got %s", r.URL.Path)
			}

			var reqBody ProjectUpdate
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Include all required fields for the generated client
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":               "proj-123",
				"name":             reqBody.Name,
				"is_public":        reqBody.IsPublic,
				"created_by":       "user-1",
				"created_at":       "2024-01-01T00:00:00Z",
				"updated_by":       "user-1",
				"updated_at":       "2024-01-01T00:00:00Z",
				"owner":            "user-1",
				"collection_count": 0,
				"capability_count": 0,
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.UpdateProject(context.Background(), "proj-123", ProjectUpdate{
			Name:     "updated-project",
			IsPublic: true,
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Name != "updated-project" {
			t.Errorf("Expected name 'updated-project', got %s", result.Name)
		}
	})
}

// TestDeleteProject tests the DeleteProject method.
func TestDeleteProject(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("Expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/v1/projects/proj-123" {
				t.Errorf("Expected /v1/projects/proj-123, got %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusNoContent)
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		err := client.DeleteProject(context.Background(), "proj-123")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

// TestCreateModelDeployment tests the CreateModelDeployment method.
func TestCreateModelDeployment(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/v1/model-deployments" {
				t.Errorf("Expected /v1/model-deployments, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ModelDeployment{
				ID:             "deploy-123",
				Name:           "test-deployment",
				ProviderID:     "prov-123",
				SupportedTasks: []string{"chat"},
				Configuration:  map[string]string{"model": "gpt-4"},
				CreatedBy:      "user-1",
				CreatedAt:      "2024-01-01T00:00:00Z",
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.CreateModelDeployment(context.Background(), ModelDeploymentCreate{
			Name:           "test-deployment",
			ProviderID:     "prov-123",
			SupportedTasks: []string{"chat"},
			Configuration:  map[string]string{"model": "gpt-4"},
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "deploy-123" {
			t.Errorf("Expected ID 'deploy-123', got %s", result.ID)
		}
	})
}

// TestGetModelDeployment tests the GetModelDeployment method.
func TestGetModelDeployment(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/v1/model-deployments/deploy-123" {
				t.Errorf("Expected /v1/model-deployments/deploy-123, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ModelDeployment{
				ID:             "deploy-123",
				Name:           "test-deployment",
				ProviderID:     "prov-123",
				SupportedTasks: []string{"chat"},
				Configuration:  map[string]string{"model": "gpt-4"},
				CreatedBy:      "user-1",
				CreatedAt:      "2024-01-01T00:00:00Z",
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.GetModelDeployment(context.Background(), "deploy-123")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "deploy-123" {
			t.Errorf("Expected ID 'deploy-123', got %s", result.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		_, err := client.GetModelDeployment(context.Background(), "nonexistent")

		if err == nil {
			t.Fatal("Expected error but got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

// TestCreateModelProvider tests the CreateModelProvider method.
func TestCreateModelProvider(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/v1/model-providers" {
				t.Errorf("Expected /v1/model-providers, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ModelProvider{
				ID:            "prov-123",
				Name:          "test-provider",
				ProviderType:  "openai",
				Configuration: map[string]string{"api_key": "secret"},
				CreatedBy:     "user-1",
				CreatedAt:     "2024-01-01T00:00:00Z",
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.CreateModelProvider(context.Background(), ModelProviderCreate{
			Name:          "test-provider",
			ProviderType:  "openai",
			Configuration: map[string]string{"api_key": "secret"},
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "prov-123" {
			t.Errorf("Expected ID 'prov-123', got %s", result.ID)
		}
	})
}

// TestGetModelProvider tests the GetModelProvider method.
func TestGetModelProvider(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/v1/model-providers/prov-123" {
				t.Errorf("Expected /v1/model-providers/prov-123, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ModelProvider{
				ID:            "prov-123",
				Name:          "test-provider",
				ProviderType:  "openai",
				Configuration: map[string]string{"api_key": "secret"},
				CreatedBy:     "user-1",
				CreatedAt:     "2024-01-01T00:00:00Z",
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.GetModelProvider(context.Background(), "prov-123")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "prov-123" {
			t.Errorf("Expected ID 'prov-123', got %s", result.ID)
		}
	})
}

// TestAPIErrorIs tests the errors.Is functionality for APIError.
func TestAPIErrorIs(t *testing.T) {
	t.Run("is ErrNotFound", func(t *testing.T) {
		err := &APIError{StatusCode: http.StatusNotFound, Message: "not found"}
		if !errors.Is(err, ErrNotFound) {
			t.Error("Expected error to match ErrNotFound")
		}
	})

	t.Run("is not ErrNotFound", func(t *testing.T) {
		err := &APIError{StatusCode: http.StatusBadRequest, Message: "bad request"}
		if errors.Is(err, ErrNotFound) {
			t.Error("Expected error to not match ErrNotFound")
		}
	})
}

// TestGetCapability tests the GetCapability method.
func TestGetCapability(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/v1/capabilities/cap-123" {
				t.Errorf("Expected /v1/capabilities/cap-123, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":                 "cap-123",
				"name":               "test-capability",
				"type":               "chat",
				"owner":              "user-1",
				"is_public":          false,
				"semantic_id":        "my-chat-capability",
				"created_by":         "user-1",
				"updated_by":         "user-1",
				"created_at":         "2024-01-15T10:30:00Z",
				"updated_at":         "2024-01-15T10:30:00Z",
				"input":              map[string]interface{}{},
				"output":             map[string]interface{}{},
				"configuration":      map[string]interface{}{},
				"is_default_version": true,
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.GetCapability(context.Background(), "cap-123")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "cap-123" {
			t.Errorf("Expected ID 'cap-123', got %s", result.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		_, err := client.GetCapability(context.Background(), "nonexistent")

		if err == nil {
			t.Fatal("Expected error but got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

// TestDeleteCapability tests the DeleteCapability method.
func TestDeleteCapability(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("Expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/v1/capabilities/cap-123" {
				t.Errorf("Expected /v1/capabilities/cap-123, got %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusNoContent)
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		err := client.DeleteCapability(context.Background(), "cap-123")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

// TestListCapabilityTypes tests the ListCapabilityTypes method.
func TestListCapabilityTypes(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/v1/capability-types" {
				t.Errorf("Expected /v1/capability-types, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(CapabilityTypesRepresentation{
				Embedded: []CapabilityTypeRepresentation{
					{ID: "chat", Name: "Chat"},
					{ID: "completion", Name: "Completion"},
				},
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.ListCapabilityTypes(context.Background())

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(result.Embedded) != 2 {
			t.Errorf("Expected 2 capability types, got %d", len(result.Embedded))
		}
	})
}

// TestGetCapabilityType tests the GetCapabilityType method.
func TestGetCapabilityType(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/v1/capability-types/chat" {
				t.Errorf("Expected /v1/capability-types/chat, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(CapabilityTypeRepresentation{
				ID:   "chat",
				Name: "Chat",
			})
		}

		server, client := setupTestServer(t, handler)
		defer server.Close()

		result, err := client.GetCapabilityType(context.Background(), "chat")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.ID != "chat" {
			t.Errorf("Expected ID 'chat', got %s", result.ID)
		}
	})
}
