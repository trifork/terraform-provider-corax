// Copyright (c) Trifork

package coraxclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	api "terraform-provider-corax/internal/generated"
)

const (
	defaultTimeout = 30 * time.Second
	apiKeyHeader   = "X-API-Key"
)

// Client manages communication with the Corax API.
// This is a wrapper around the generated OpenAPI client that provides
// a simpler interface with string-based date/time fields.
type Client struct {
	// HTTP client used to communicate with the API.
	httpClient *http.Client

	// Base URL for API requests. Must include scheme and host.
	BaseURL *url.URL

	// API key for authentication.
	APIKey string

	// UserAgent for client
	UserAgent string

	// generated is the underlying OpenAPI-generated client
	generated *api.APIClient
}

// NewClient returns a new Corax API client.
func NewClient(baseURLStr string, apiKey string) (*Client, error) {
	if strings.TrimSpace(baseURLStr) == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("apiKey cannot be empty")
	}

	parsedBaseURL, err := url.ParseRequestURI(baseURLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid baseURL: %w", err)
	}
	if parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" {
		return nil, fmt.Errorf("baseURL must include scheme and host")
	}

	// Configure the generated client
	cfg := api.NewConfiguration()
	cfg.Servers = api.ServerConfigurations{
		{URL: baseURLStr},
	}
	cfg.UserAgent = "terraform-provider-corax/0.0.1"
	cfg.HTTPClient = &http.Client{
		Timeout: defaultTimeout,
	}

	return &Client{
		httpClient: cfg.HTTPClient,
		BaseURL:    parsedBaseURL,
		APIKey:     apiKey,
		UserAgent:  cfg.UserAgent,
		generated:  api.NewAPIClient(cfg),
	}, nil
}

// withAuth returns a context with the API key configured for authentication.
func (c *Client) withAuth(ctx context.Context) context.Context {
	return context.WithValue(ctx, api.ContextAPIKeys, map[string]api.APIKey{
		"APIKeyHeader": {Key: c.APIKey},
	})
}

// APIError represents an error response from the Corax API.
type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
	// TODO: Could include a more structured error, e.g. from HTTPValidationError schema
}

func (e *APIError) Error() string {
	if len(e.Body) > 0 {
		return fmt.Sprintf("API Error: status %d, body: %s", e.StatusCode, string(e.Body))
	}
	return fmt.Sprintf("API Error: status %d, message: %s", e.StatusCode, e.Message)
}

// Is implements errors.Is interface to allow checking error types.
// This allows errors.Is(err, ErrNotFound) to work with any 404 APIError.
func (e *APIError) Is(target error) bool {
	if t, ok := target.(*APIError); ok {
		return e.StatusCode == t.StatusCode
	}
	return false
}

// ErrNotFound is returned when a resource is not found (HTTP 404).
var ErrNotFound = &APIError{StatusCode: http.StatusNotFound, Message: "resource not found"}

// convertError converts an error from the generated client to an APIError.
// It handles GenericOpenAPIError and extracts status code from HTTP response if available.
func convertError(err error, resp *http.Response) error {
	if err == nil {
		return nil
	}

	apiErr := &APIError{
		Message: err.Error(),
	}

	// Extract status code from response if available
	if resp != nil {
		apiErr.StatusCode = resp.StatusCode
	}

	// Extract body from GenericOpenAPIError if available
	if genErr, ok := err.(*api.GenericOpenAPIError); ok {
		apiErr.Body = genErr.Body()
		if len(apiErr.Body) > 0 && len(apiErr.Body) < 512 {
			apiErr.Message = string(apiErr.Body)
		}
	}

	// Set proper message for not found errors
	if apiErr.StatusCode == http.StatusNotFound {
		apiErr.Message = "resource not found"
	}

	return apiErr
}

// parseTime parses a datetime string into time.Time.
func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, s)
}

// formatTime formats a time.Time as an RFC3339 string.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// formatTimePtr formats a *time.Time as a *string (RFC3339 format).
func formatTimePtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	relURL, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %w", err)
	}

	fullURL := c.BaseURL.ResolveReference(relURL)

	var reqBody io.ReadWriter
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(apiKeyHeader, c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

func (c *Client) doRequest(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Body:       respBodyBytes,
		}
		// Try to unmarshal into a standard error structure if available
		// For now, just use a generic message or the body itself if it's short.
		if len(respBodyBytes) > 0 && len(respBodyBytes) < 512 { // Arbitrary limit for error message
			apiErr.Message = string(respBodyBytes)
		} else {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}
		if resp.StatusCode == http.StatusNotFound {
			apiErr.Message = "resource not found"
			return apiErr
		}
		return apiErr
	}

	if v != nil {
		if err := json.Unmarshal(respBodyBytes, v); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w, body: %s", err, string(respBodyBytes))
		}
	}

	return nil
}

// CreateAPIKey creates a new API key.
// Corresponds to POST /v1/api-keys.
func (c *Client) CreateAPIKey(ctx context.Context, apiKeyData ApiKeyCreate) (*ApiKey, error) {
	// Parse the expires_at string to time.Time for the generated client
	expiresAt, err := parseTime(apiKeyData.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid expires_at format: %w", err)
	}

	// Call the generated client
	genApiKey := api.ApiKeyCreate{
		Name:      apiKeyData.Name,
		ExpiresAt: expiresAt,
	}

	result, resp, err := c.generated.APIKeysAPI.CreateApiKeyV1ApiKeysPost(c.withAuth(ctx)).
		ApiKeyCreate(genApiKey).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	// Convert the generated type back to our custom type
	return convertApiKey(result), nil
}

// convertApiKey converts a generated ApiKey to our custom ApiKey type.
func convertApiKey(gen *api.ApiKey) *ApiKey {
	if gen == nil {
		return nil
	}

	result := &ApiKey{
		ID:         gen.Id,
		Name:       gen.Name,
		Key:        gen.Key,
		CreatedBy:  gen.CreatedBy,
		CreatedAt:  formatTime(gen.CreatedAt),
		IsActive:   gen.GetIsActive(),
		UsageCount: int(gen.GetUsageCount()),
	}

	if gen.Prefix != nil {
		result.Prefix = *gen.Prefix
	}

	if gen.ExpiresAt.IsSet() {
		result.ExpiresAt = formatTimePtr(gen.ExpiresAt.Get())
	}

	if gen.UpdatedAt.IsSet() {
		result.UpdatedAt = formatTimePtr(gen.UpdatedAt.Get())
	}

	if gen.LastUsedAt.IsSet() {
		result.LastUsedAt = formatTimePtr(gen.LastUsedAt.Get())
	}

	return result
}

// GetAPIKey retrieves a specific API key by its ID.
// Corresponds to GET /v1/api-keys/{key_id}.
func (c *Client) GetAPIKey(ctx context.Context, keyID string) (*ApiKey, error) {
	if strings.TrimSpace(keyID) == "" {
		return nil, fmt.Errorf("keyID cannot be empty")
	}
	path := fmt.Sprintf("/v1/api-keys/%s", keyID)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var apiKey ApiKey
	if err := c.doRequest(req, &apiKey); err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// DeleteAPIKey deletes a specific API key by its ID.
// Corresponds to DELETE /v1/api-keys/{key_id}.
// The OpenAPI spec indicates a 200 response with an empty JSON object {} on success.
func (c *Client) DeleteAPIKey(ctx context.Context, keyID string) error {
	if strings.TrimSpace(keyID) == "" {
		return fmt.Errorf("keyID cannot be empty")
	}

	resp, err := c.generated.APIKeysAPI.DeleteApiKeyV1ApiKeysKeyIdDelete(c.withAuth(ctx), keyID).Execute()
	if err != nil {
		return convertError(err, resp)
	}
	return nil
}

// --- Project Methods ---

// CreateProject creates a new project.
// Corresponds to POST /v1/projects.
func (c *Client) CreateProject(ctx context.Context, projectData ProjectCreate) (*Project, error) {
	genProjectCreate := api.NewProjectCreate(projectData.Name)
	if projectData.Description != nil {
		genProjectCreate.SetDescription(*projectData.Description)
	}
	if projectData.IsPublic != nil {
		genProjectCreate.SetIsPublic(*projectData.IsPublic)
	}

	result, resp, err := c.generated.ProjectsAPI.CreateProjectV1ProjectsPost(c.withAuth(ctx)).
		ProjectCreate(*genProjectCreate).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertProject(result), nil
}

// GetProject retrieves a specific project by its ID.
// Corresponds to GET /v1/projects/{project_id}.
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fmt.Errorf("projectID cannot be empty")
	}

	result, resp, err := c.generated.ProjectsAPI.GetProjectV1ProjectsProjectIdGet(c.withAuth(ctx), projectID).Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertProject(result), nil
}

// UpdateProject updates a specific project by its ID.
// Corresponds to PUT /v1/projects/{project_id}.
func (c *Client) UpdateProject(ctx context.Context, projectID string, projectData ProjectUpdate) (*Project, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fmt.Errorf("projectID cannot be empty")
	}

	genProjectUpdate := api.ProjectUpdate{
		Name:     projectData.Name,
		IsPublic: *api.NewNullableBool(&projectData.IsPublic),
	}
	if projectData.Description != nil {
		genProjectUpdate.SetDescription(*projectData.Description)
	}

	result, resp, err := c.generated.ProjectsAPI.UpdateProjectV1ProjectsProjectIdPut(c.withAuth(ctx), projectID).
		ProjectUpdate(genProjectUpdate).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertProject(result), nil
}

// DeleteProject deletes a specific project by its ID.
// Corresponds to DELETE /v1/projects/{project_id}.
// Expects a 204 No Content on success.
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	if strings.TrimSpace(projectID) == "" {
		return fmt.Errorf("projectID cannot be empty")
	}

	resp, err := c.generated.ProjectsAPI.DeleteProjectV1ProjectsProjectIdDelete(c.withAuth(ctx), projectID).Execute()
	if err != nil {
		return convertError(err, resp)
	}
	return nil
}

// convertProject converts a generated Project to our custom Project type.
func convertProject(gen *api.Project) *Project {
	if gen == nil {
		return nil
	}

	result := &Project{
		ID:        gen.Id,
		Name:      gen.Name,
		IsPublic:  gen.GetIsPublic(),
		CreatedBy: gen.CreatedBy,
		CreatedAt: formatTime(gen.CreatedAt),
		Owner:     gen.Owner,
	}

	if gen.Description.IsSet() {
		result.Description = gen.Description.Get()
	}

	if gen.UpdatedBy != "" {
		result.UpdatedBy = &gen.UpdatedBy
	}

	if !gen.UpdatedAt.IsZero() {
		s := formatTime(gen.UpdatedAt)
		result.UpdatedAt = &s
	}

	if gen.CollectionCount != nil {
		result.CollectionCount = int(*gen.CollectionCount)
	}

	if gen.CapabilityCount != nil {
		result.CapabilityCount = int(*gen.CapabilityCount)
	}

	return result
}

// --- Collection Methods --- (REMOVED)
// --- Document Methods --- (REMOVED)
// --- Embeddings Model Methods --- (REMOVED)

// --- Capability Methods ---

// convertCapabilityConfig converts a generated CapabilityConfig to our custom type.
func convertCapabilityConfig(gen *api.CapabilityConfig) *CapabilityConfig {
	if gen == nil {
		return nil
	}

	result := &CapabilityConfig{
		ContentTracing:   gen.ContentTracing,
		CustomParameters: gen.CustomParameters,
	}

	if gen.Temperature.IsSet() {
		temp := gen.Temperature.Get()
		if temp != nil {
			f64 := float64(*temp)
			result.Temperature = &f64
		}
	}

	if gen.BlobConfig.IsSet() {
		bc := gen.BlobConfig.Get()
		if bc != nil {
			result.BlobConfig = &BlobConfig{
				AllowedMimeTypes: bc.AllowedMimeTypes,
			}
			if bc.MaxFileSizeMb != nil {
				val := int(*bc.MaxFileSizeMb)
				result.BlobConfig.MaxFileSizeMB = &val
			}
			if bc.MaxBlobs != nil {
				val := int(*bc.MaxBlobs)
				result.BlobConfig.MaxBlobs = &val
			}
		}
	}

	if gen.DataRetention != nil {
		if gen.DataRetention.InfiniteDataRetention != nil {
			result.DataRetention = &DataRetention{
				Type: "infinite",
			}
		} else if gen.DataRetention.TimedDataRetention != nil {
			hours := int(gen.DataRetention.TimedDataRetention.Hours)
			result.DataRetention = &DataRetention{
				Type:  "timed",
				Hours: &hours,
			}
		}
	}

	return result
}

// convertCapabilityRepresentation converts a generated CapabilityRepresentation to our custom type.
func convertCapabilityRepresentation(gen *api.CapabilityRepresentation) *CapabilityRepresentation {
	if gen == nil {
		return nil
	}

	result := &CapabilityRepresentation{
		ID:            gen.Id,
		Name:          gen.Name,
		Type:          gen.Type,
		SemanticID:    gen.SemanticId,
		CreatedBy:     gen.CreatedBy,
		UpdatedBy:     gen.UpdatedBy,
		CreatedAt:     formatTime(gen.CreatedAt),
		UpdatedAt:     formatTime(gen.UpdatedAt),
		Owner:         gen.Owner,
		Input:         gen.Input,
		Output:        gen.Output,
		Configuration: gen.Configuration,
	}

	if gen.IsPublic.IsSet() {
		result.IsPublic = gen.IsPublic.Get()
	}

	if gen.ModelId.IsSet() {
		result.ModelID = gen.ModelId.Get()
	}

	if gen.Config.IsSet() {
		cfg := gen.Config.Get()
		result.Config = convertCapabilityConfig(cfg)
	}

	if gen.ProjectId.IsSet() {
		result.ProjectID = gen.ProjectId.Get()
	}

	if gen.ArchivedAt.IsSet() {
		result.ArchivedAt = formatTimePtr(gen.ArchivedAt.Get())
	}

	return result
}

// convertCapabilityConfigToGen converts our CapabilityConfig to the generated type.
func convertCapabilityConfigToGen(cfg *CapabilityConfig) *api.CapabilityConfig {
	if cfg == nil {
		return nil
	}

	result := api.NewCapabilityConfig()

	if cfg.Temperature != nil {
		result.SetTemperature(float32(*cfg.Temperature))
	}

	if cfg.CustomParameters != nil {
		result.CustomParameters = cfg.CustomParameters
	}

	if cfg.ContentTracing != nil {
		result.ContentTracing = cfg.ContentTracing
	}

	if cfg.BlobConfig != nil {
		bc := api.BlobConfig{
			AllowedMimeTypes: cfg.BlobConfig.AllowedMimeTypes,
		}
		if cfg.BlobConfig.MaxFileSizeMB != nil {
			val := int32(*cfg.BlobConfig.MaxFileSizeMB)
			bc.MaxFileSizeMb = &val
		}
		if cfg.BlobConfig.MaxBlobs != nil {
			val := int32(*cfg.BlobConfig.MaxBlobs)
			bc.MaxBlobs = &val
		}
		result.BlobConfig = *api.NewNullableBlobConfig(&bc)
	}

	if cfg.DataRetention != nil {
		switch cfg.DataRetention.Type {
		case "infinite":
			result.DataRetention = &api.DataRetention{
				InfiniteDataRetention: api.NewInfiniteDataRetention(),
			}
		case "timed":
			hours := int32(0)
			if cfg.DataRetention.Hours != nil {
				hours = int32(*cfg.DataRetention.Hours)
			}
			result.DataRetention = &api.DataRetention{
				TimedDataRetention: api.NewTimedDataRetention(hours),
			}
		}
	}

	return result
}

// convertChatCapabilityUpdateToGen converts our ChatCapabilityUpdate to the generated type.
func convertChatCapabilityUpdateToGen(u *ChatCapabilityUpdate) *api.ChatCapabilityUpdate {
	if u == nil {
		return nil
	}

	// Name and Type are required in the generated type
	name := ""
	if u.Name != nil {
		name = *u.Name
	}
	capType := "chat"
	if u.Type != nil {
		capType = *u.Type
	}

	result := api.NewChatCapabilityUpdate(name, capType)

	if u.IsPublic != nil {
		result.SetIsPublic(*u.IsPublic)
	}

	if u.ModelID != nil {
		result.SetModelId(*u.ModelID)
	}

	if u.Config != nil {
		genCfg := convertCapabilityConfigToGen(u.Config)
		if genCfg != nil {
			result.SetConfig(*genCfg)
		}
	}

	if u.ProjectID != nil {
		result.SetProjectId(*u.ProjectID)
	}

	if u.SystemPrompt != nil {
		result.SetSystemPrompt(*u.SystemPrompt)
	}

	return result
}

// convertCompletionCapabilityUpdateToGen converts our CompletionCapabilityUpdate to the generated type.
func convertCompletionCapabilityUpdateToGen(u *CompletionCapabilityUpdate) *api.CompletionCapabilityUpdate {
	if u == nil {
		return nil
	}

	// Name and Type are required in the generated type
	name := ""
	if u.Name != nil {
		name = *u.Name
	}
	capType := "completion"
	if u.Type != nil {
		capType = *u.Type
	}

	result := api.NewCompletionCapabilityUpdate(name, capType)

	if u.IsPublic != nil {
		result.SetIsPublic(*u.IsPublic)
	}

	if u.SemanticID != nil {
		result.SetSemanticId(*u.SemanticID)
	}

	if u.ModelID != nil {
		result.SetModelId(*u.ModelID)
	}

	if u.Config != nil {
		genCfg := convertCapabilityConfigToGen(u.Config)
		if genCfg != nil {
			result.SetConfig(*genCfg)
		}
	}

	if u.ProjectID != nil {
		result.SetProjectId(*u.ProjectID)
	}

	if u.SystemPrompt != nil {
		result.SetSystemPrompt(*u.SystemPrompt)
	}

	if u.CompletionPrompt != nil {
		result.SetCompletionPrompt(*u.CompletionPrompt)
	}

	if u.Variables != nil {
		result.Variables = u.Variables
	}

	if u.OutputType != nil {
		result.SetOutputType(*u.OutputType)
	}

	// Note: SchemaDef conversion is not implemented as the generated type uses a complex
	// polymorphic structure (BasicProperty, EnumProperty, etc.) while our hand-written
	// type uses map[string]interface{}. This can be added if schema capabilities are needed.

	return result
}

// convertChatCapabilityCreateToGen converts our ChatCapabilityCreate to the generated type.
func convertChatCapabilityCreateToGen(c *ChatCapabilityCreate) *api.ChatCapabilityCreate {
	if c == nil {
		return nil
	}

	result := api.NewChatCapabilityCreate(c.Name, c.Type, c.SystemPrompt)

	if c.IsPublic != nil {
		result.SetIsPublic(*c.IsPublic)
	}

	if c.ModelID != nil {
		result.SetModelId(*c.ModelID)
	}

	if c.Config != nil {
		genCfg := convertCapabilityConfigToGen(c.Config)
		if genCfg != nil {
			result.SetConfig(*genCfg)
		}
	}

	if c.ProjectID != nil {
		result.SetProjectId(*c.ProjectID)
	}

	return result
}

// convertCompletionCapabilityCreateToGen converts our CompletionCapabilityCreate to the generated type.
func convertCompletionCapabilityCreateToGen(c *CompletionCapabilityCreate) *api.CompletionCapabilityCreate {
	if c == nil {
		return nil
	}

	result := api.NewCompletionCapabilityCreate(c.Name, c.Type, c.SystemPrompt, c.CompletionPrompt, c.OutputType)

	if c.IsPublic != nil {
		result.SetIsPublic(*c.IsPublic)
	}

	if c.SemanticID != nil {
		result.SetSemanticId(*c.SemanticID)
	}

	if c.ModelID != nil {
		result.SetModelId(*c.ModelID)
	}

	if c.Config != nil {
		genCfg := convertCapabilityConfigToGen(c.Config)
		if genCfg != nil {
			result.SetConfig(*genCfg)
		}
	}

	if c.ProjectID != nil {
		result.SetProjectId(*c.ProjectID)
	}

	if c.Variables != nil {
		result.Variables = c.Variables
	}

	// Note: SchemaDef conversion is not implemented (see CompletionCapabilityUpdate comment)

	return result
}

// convertChatCapabilityToRepresentation converts a generated ChatCapability to CapabilityRepresentation.
func convertChatCapabilityToRepresentation(gen *api.ChatCapability) *CapabilityRepresentation {
	if gen == nil {
		return nil
	}

	capType := "chat"
	if gen.Type != nil {
		capType = *gen.Type
	}

	result := &CapabilityRepresentation{
		ID:            gen.Id,
		Name:          gen.Name,
		Type:          capType,
		CreatedBy:     gen.CreatedBy,
		UpdatedBy:     gen.UpdatedBy,
		CreatedAt:     formatTime(gen.CreatedAt),
		UpdatedAt:     formatTime(gen.UpdatedAt),
		Owner:         gen.Owner,
		Input:         make(map[string]interface{}),
		Output:        make(map[string]interface{}),
		Configuration: make(map[string]interface{}),
	}

	if gen.IsPublic.IsSet() {
		result.IsPublic = gen.IsPublic.Get()
	}

	if gen.ModelId.IsSet() {
		result.ModelID = gen.ModelId.Get()
	}

	if gen.Config.IsSet() {
		cfg := gen.Config.Get()
		result.Config = convertCapabilityConfig(cfg)
	}

	if gen.ProjectId.IsSet() {
		result.ProjectID = gen.ProjectId.Get()
	}

	if gen.SemanticId.IsSet() && gen.SemanticId.Get() != nil {
		result.SemanticID = *gen.SemanticId.Get()
	}

	if gen.ArchivedAt.IsSet() {
		result.ArchivedAt = formatTimePtr(gen.ArchivedAt.Get())
	}

	// Store chat-specific fields in Configuration
	result.Configuration["system_prompt"] = gen.SystemPrompt

	return result
}

// convertCompletionCapabilityToRepresentation converts a generated CompletionCapability to CapabilityRepresentation.
func convertCompletionCapabilityToRepresentation(gen *api.CompletionCapability) *CapabilityRepresentation {
	if gen == nil {
		return nil
	}

	capType := "completion"
	if gen.Type != nil {
		capType = *gen.Type
	}

	result := &CapabilityRepresentation{
		ID:            gen.Id,
		Name:          gen.Name,
		Type:          capType,
		CreatedBy:     gen.CreatedBy,
		UpdatedBy:     gen.UpdatedBy,
		CreatedAt:     formatTime(gen.CreatedAt),
		UpdatedAt:     formatTime(gen.UpdatedAt),
		Owner:         gen.Owner,
		Input:         make(map[string]interface{}),
		Output:        make(map[string]interface{}),
		Configuration: make(map[string]interface{}),
	}

	if gen.IsPublic.IsSet() {
		result.IsPublic = gen.IsPublic.Get()
	}

	if gen.ModelId.IsSet() {
		result.ModelID = gen.ModelId.Get()
	}

	if gen.Config.IsSet() {
		cfg := gen.Config.Get()
		result.Config = convertCapabilityConfig(cfg)
	}

	if gen.ProjectId.IsSet() {
		result.ProjectID = gen.ProjectId.Get()
	}

	if gen.SemanticId.IsSet() && gen.SemanticId.Get() != nil {
		result.SemanticID = *gen.SemanticId.Get()
	}

	if gen.ArchivedAt.IsSet() {
		result.ArchivedAt = formatTimePtr(gen.ArchivedAt.Get())
	}

	// Store completion-specific fields
	result.Configuration["system_prompt"] = gen.SystemPrompt
	result.Configuration["completion_prompt"] = gen.CompletionPrompt
	result.Output["type"] = gen.OutputType
	if gen.Variables != nil {
		result.Input["variables"] = gen.Variables
	}
	// Note: schema_def would need conversion from the generated type

	return result
}

// convertCreateCapabilityResponse converts the generated create response to CapabilityRepresentation.
func convertCreateCapabilityResponse(resp *api.ResponseCreateCapabilityV1CapabilitiesPost) (*CapabilityRepresentation, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}

	if resp.ChatCapability != nil {
		return convertChatCapabilityToRepresentation(resp.ChatCapability), nil
	}

	if resp.CompletionCapability != nil {
		return convertCompletionCapabilityToRepresentation(resp.CompletionCapability), nil
	}

	// Note: ExtractionCapability and SpeechToTextCapability not supported yet
	if resp.ExtractionCapability != nil {
		return nil, fmt.Errorf("ExtractionCapability not supported")
	}

	if resp.SpeechToTextCapability != nil {
		return nil, fmt.Errorf("SpeechToTextCapability not supported")
	}

	return nil, fmt.Errorf("unknown capability type in response")
}

// CreateCapability creates a new capability.
// The payload should be either ChatCapabilityCreate or CompletionCapabilityCreate.
// Corresponds to POST /v1/capabilities.
func (c *Client) CreateCapability(ctx context.Context, capabilityData interface{}) (*CapabilityRepresentation, error) {
	// Convert capabilityData to the generated Capability1 union type
	var cap1 api.Capability1
	switch v := capabilityData.(type) {
	case ChatCapabilityCreate:
		cap1.ChatCapabilityCreate = convertChatCapabilityCreateToGen(&v)
	case *ChatCapabilityCreate:
		cap1.ChatCapabilityCreate = convertChatCapabilityCreateToGen(v)
	case CompletionCapabilityCreate:
		cap1.CompletionCapabilityCreate = convertCompletionCapabilityCreateToGen(&v)
	case *CompletionCapabilityCreate:
		cap1.CompletionCapabilityCreate = convertCompletionCapabilityCreateToGen(v)
	default:
		return nil, fmt.Errorf("CreateCapability: unsupported capability type %T", capabilityData)
	}

	result, resp, err := c.generated.CapabilitiesAPI.CreateCapabilityV1CapabilitiesPost(c.withAuth(ctx)).
		Capability1(cap1).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertCreateCapabilityResponse(result)
}

// GetCapability retrieves a specific capability by its ID.
// Corresponds to GET /v1/capabilities/{capability_id}.
func (c *Client) GetCapability(ctx context.Context, capabilityID string) (*CapabilityRepresentation, error) {
	if strings.TrimSpace(capabilityID) == "" {
		return nil, fmt.Errorf("capabilityID cannot be empty")
	}

	capId := api.CapabilityId{String: &capabilityID}

	result, resp, err := c.generated.CapabilitiesAPI.ReadCapabilityV1CapabilitiesCapabilityIdGet(c.withAuth(ctx), capId).Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertCapabilityRepresentation(result), nil
}

// UpdateCapability updates a specific capability by its ID.
// The payload should be either ChatCapabilityUpdate or CompletionCapabilityUpdate.
// Corresponds to PUT /v1/capabilities/{capability_id}.
func (c *Client) UpdateCapability(ctx context.Context, capabilityID string, capabilityData interface{}) (*CapabilityRepresentation, error) {
	if strings.TrimSpace(capabilityID) == "" {
		return nil, fmt.Errorf("capabilityID cannot be empty")
	}

	capId := api.CapabilityId1{String: &capabilityID}

	// Convert capabilityData to the generated Capability2 union type
	var cap2 api.Capability2
	switch v := capabilityData.(type) {
	case ChatCapabilityUpdate:
		cap2.ChatCapabilityUpdate = convertChatCapabilityUpdateToGen(&v)
	case *ChatCapabilityUpdate:
		cap2.ChatCapabilityUpdate = convertChatCapabilityUpdateToGen(v)
	case CompletionCapabilityUpdate:
		cap2.CompletionCapabilityUpdate = convertCompletionCapabilityUpdateToGen(&v)
	case *CompletionCapabilityUpdate:
		cap2.CompletionCapabilityUpdate = convertCompletionCapabilityUpdateToGen(v)
	default:
		return nil, fmt.Errorf("UpdateCapability: unsupported capability type %T", capabilityData)
	}

	result, resp, err := c.generated.CapabilitiesAPI.UpdateCapabilityV1CapabilitiesCapabilityIdPut(c.withAuth(ctx), capId).
		Capability2(cap2).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertCapabilityRepresentation(result), nil
}

// DeleteCapability deletes a specific capability by its ID.
// Corresponds to DELETE /v1/capabilities/{capability_id}.
// Expects a 204 No Content on success.
func (c *Client) DeleteCapability(ctx context.Context, capabilityID string) error {
	if strings.TrimSpace(capabilityID) == "" {
		return fmt.Errorf("capabilityID cannot be empty")
	}

	capId := api.CapabilityId1{String: &capabilityID}

	resp, err := c.generated.CapabilitiesAPI.DeleteCapabilityV1CapabilitiesCapabilityIdDelete(c.withAuth(ctx), capId).Execute()

	if err != nil {
		return convertError(err, resp)
	}

	return nil
}

// --- ModelDeployment Methods ---

// convertSupportedTasksToGen converts []string to []api.CapabilityType.
func convertSupportedTasksToGen(tasks []string) []api.CapabilityType {
	result := make([]api.CapabilityType, len(tasks))
	for i, t := range tasks {
		result[i] = api.CapabilityType(t)
	}
	return result
}

// convertSupportedTasksFromGen converts []api.CapabilityType to []string.
func convertSupportedTasksFromGen(tasks []api.CapabilityType) []string {
	result := make([]string, len(tasks))
	for i, t := range tasks {
		result[i] = string(t)
	}
	return result
}

// convertConfigurationToGen converts map[string]string to map[string]interface{}.
func convertConfigurationToGen(config map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(config))
	for k, v := range config {
		result[k] = v
	}
	return result
}

// convertConfigurationFromGen converts map[string]interface{} to map[string]string.
func convertConfigurationFromGen(config map[string]interface{}) map[string]string {
	result := make(map[string]string, len(config))
	for k, v := range config {
		if s, ok := v.(string); ok {
			result[k] = s
		} else {
			// Convert non-string values to string representation
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

// convertModelDeployment converts a generated ModelDeployment to our custom type.
func convertModelDeployment(gen *api.ModelDeployment) *ModelDeployment {
	if gen == nil {
		return nil
	}

	result := &ModelDeployment{
		ID:             gen.Id,
		Name:           gen.Name,
		SupportedTasks: convertSupportedTasksFromGen(gen.SupportedTasks),
		Configuration:  convertConfigurationFromGen(gen.Configuration),
		ProviderID:     gen.ProviderId,
		CreatedAt:      formatTime(gen.CreatedAt),
		CreatedBy:      gen.CreatedBy,
	}

	if gen.Description.IsSet() {
		result.Description = gen.Description.Get()
	}

	if gen.IsActive != nil {
		result.IsActive = gen.IsActive
	}

	if gen.UpdatedAt.IsSet() {
		result.UpdatedAt = formatTimePtr(gen.UpdatedAt.Get())
	}

	if gen.UpdatedBy.IsSet() {
		result.UpdatedBy = gen.UpdatedBy.Get()
	}

	return result
}

// CreateModelDeployment creates a new model deployment.
// Corresponds to POST /v1/model-deployments.
func (c *Client) CreateModelDeployment(ctx context.Context, deploymentData ModelDeploymentCreate) (*ModelDeployment, error) {
	genCreate := api.NewModelDeploymentCreate(
		deploymentData.Name,
		convertSupportedTasksToGen(deploymentData.SupportedTasks),
		convertConfigurationToGen(deploymentData.Configuration),
		deploymentData.ProviderID,
	)

	if deploymentData.Description != nil {
		genCreate.SetDescription(*deploymentData.Description)
	}
	if deploymentData.IsActive != nil {
		genCreate.SetIsActive(*deploymentData.IsActive)
	}

	result, resp, err := c.generated.ModelDeploymentsAPI.CreateModelDeploymentV1ModelDeploymentsPost(c.withAuth(ctx)).
		ModelDeploymentCreate(*genCreate).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertModelDeployment(result), nil
}

// GetModelDeployment retrieves a specific model deployment by its ID.
// Corresponds to GET /v1/model-deployments/{deployment_id}.
func (c *Client) GetModelDeployment(ctx context.Context, deploymentID string) (*ModelDeployment, error) {
	if strings.TrimSpace(deploymentID) == "" {
		return nil, fmt.Errorf("deploymentID cannot be empty")
	}

	result, resp, err := c.generated.ModelDeploymentsAPI.GetModelDeploymentV1ModelDeploymentsDeploymentIdGet(c.withAuth(ctx), deploymentID).Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertModelDeployment(result), nil
}

// UpdateModelDeployment updates a specific model deployment by its ID.
// Corresponds to PUT /v1/model-deployments/{deployment_id}.
func (c *Client) UpdateModelDeployment(ctx context.Context, deploymentID string, deploymentData ModelDeploymentUpdate) (*ModelDeployment, error) {
	if strings.TrimSpace(deploymentID) == "" {
		return nil, fmt.Errorf("deploymentID cannot be empty")
	}

	genUpdate := api.NewModelDeploymentUpdate(
		deploymentData.Name,
		convertSupportedTasksToGen(deploymentData.SupportedTasks),
		convertConfigurationToGen(deploymentData.Configuration),
		deploymentData.ProviderID,
	)

	if deploymentData.Description != nil {
		genUpdate.SetDescription(*deploymentData.Description)
	}
	if deploymentData.IsActive != nil {
		genUpdate.SetIsActive(*deploymentData.IsActive)
	}

	result, resp, err := c.generated.ModelDeploymentsAPI.UpdateModelDeploymentV1ModelDeploymentsDeploymentIdPut(c.withAuth(ctx), deploymentID).
		ModelDeploymentUpdate(*genUpdate).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertModelDeployment(result), nil
}

// DeleteModelDeployment deletes a specific model deployment by its ID.
// Corresponds to DELETE /v1/model-deployments/{deployment_id}.
// Expects a 204 No Content on success.
func (c *Client) DeleteModelDeployment(ctx context.Context, deploymentID string) error {
	if strings.TrimSpace(deploymentID) == "" {
		return fmt.Errorf("deploymentID cannot be empty")
	}

	resp, err := c.generated.ModelDeploymentsAPI.DeleteModelDeploymentV1ModelDeploymentsDeploymentIdDelete(c.withAuth(ctx), deploymentID).Execute()
	if err != nil {
		return convertError(err, resp)
	}
	return nil
}

// --- ModelProvider Methods ---

// convertModelProvider converts a generated ModelProvider to our custom type.
func convertModelProvider(gen *api.ModelProvider) *ModelProvider {
	if gen == nil {
		return nil
	}

	result := &ModelProvider{
		ID:            gen.Id,
		Name:          gen.Name,
		ProviderType:  gen.ProviderType,
		Configuration: convertConfigurationFromGen(gen.Configuration),
		CreatedAt:     formatTime(gen.CreatedAt),
		CreatedBy:     gen.CreatedBy,
	}

	if gen.UpdatedAt.IsSet() {
		result.UpdatedAt = formatTimePtr(gen.UpdatedAt.Get())
	}

	if gen.UpdatedBy.IsSet() {
		result.UpdatedBy = gen.UpdatedBy.Get()
	}

	return result
}

// CreateModelProvider creates a new model provider.
// Corresponds to POST /v1/model-providers.
func (c *Client) CreateModelProvider(ctx context.Context, providerData ModelProviderCreate) (*ModelProvider, error) {
	genCreate := api.NewModelProviderCreate(
		providerData.Name,
		providerData.ProviderType,
		convertConfigurationToGen(providerData.Configuration),
	)

	result, resp, err := c.generated.ModelProvidersAPI.CreateModelProviderV1ModelProvidersPost(c.withAuth(ctx)).
		ModelProviderCreate(*genCreate).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertModelProvider(result), nil
}

// GetModelProvider retrieves a specific model provider by its ID.
// Corresponds to GET /v1/model-providers/{provider_id}.
func (c *Client) GetModelProvider(ctx context.Context, providerID string) (*ModelProvider, error) {
	if strings.TrimSpace(providerID) == "" {
		return nil, fmt.Errorf("providerID cannot be empty")
	}

	result, resp, err := c.generated.ModelProvidersAPI.GetModelProviderV1ModelProvidersProviderIdGet(c.withAuth(ctx), providerID).Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertModelProvider(result), nil
}

// UpdateModelProvider updates a specific model provider by its ID.
// Corresponds to PUT /v1/model-providers/{provider_id}.
func (c *Client) UpdateModelProvider(ctx context.Context, providerID string, providerData ModelProviderUpdate) (*ModelProvider, error) {
	if strings.TrimSpace(providerID) == "" {
		return nil, fmt.Errorf("providerID cannot be empty")
	}

	genUpdate := api.NewModelProviderUpdate(
		providerData.Name,
		providerData.ProviderType,
		convertConfigurationToGen(providerData.Configuration),
		providerData.ID,
	)

	result, resp, err := c.generated.ModelProvidersAPI.UpdateModelProviderV1ModelProvidersProviderIdPut(c.withAuth(ctx), providerID).
		ModelProviderUpdate(*genUpdate).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertModelProvider(result), nil
}

// DeleteModelProvider deletes a specific model provider by its ID.
// Corresponds to DELETE /v1/model-providers/{provider_id}.
// Expects a 204 No Content on success.
func (c *Client) DeleteModelProvider(ctx context.Context, providerID string) error {
	if strings.TrimSpace(providerID) == "" {
		return fmt.Errorf("providerID cannot be empty")
	}

	resp, err := c.generated.ModelProvidersAPI.DeleteModelProviderV1ModelProvidersProviderIdDelete(c.withAuth(ctx), providerID).Execute()
	if err != nil {
		return convertError(err, resp)
	}
	return nil
}

// --- CapabilityType Methods ---

// convertCapabilityTypeRepresentation converts a generated CapabilityTypeRepresentation to our custom type.
func convertCapabilityTypeRepresentation(gen *api.CapabilityTypeRepresentation) *CapabilityTypeRepresentation {
	if gen == nil {
		return nil
	}

	result := &CapabilityTypeRepresentation{
		ID:   gen.Id,
		Name: gen.Name,
	}

	if gen.DefaultModelDeploymentId.IsSet() {
		result.DefaultModelDeploymentID = gen.DefaultModelDeploymentId.Get()
	}

	return result
}

// convertCapabilityTypesRepresentation converts a generated CapabilityTypesRepresentation to our custom type.
func convertCapabilityTypesRepresentation(gen *api.CapabilityTypesRepresentation) *CapabilityTypesRepresentation {
	if gen == nil {
		return nil
	}

	embedded := make([]CapabilityTypeRepresentation, len(gen.Embedded))
	for i, ct := range gen.Embedded {
		converted := convertCapabilityTypeRepresentation(&ct)
		if converted != nil {
			embedded[i] = *converted
		}
	}

	return &CapabilityTypesRepresentation{
		Embedded: embedded,
	}
}

// GetCapabilityType retrieves a specific capability type definition.
// Corresponds to GET /v1/capability-types/{capability_type}.
func (c *Client) GetCapabilityType(ctx context.Context, capabilityType string) (*CapabilityTypeRepresentation, error) {
	if strings.TrimSpace(capabilityType) == "" {
		return nil, fmt.Errorf("capabilityType cannot be empty")
	}

	result, resp, err := c.generated.CapabilityTypesAPI.GetCapabilityTypeV1CapabilityTypesCapabilityTypeGet(c.withAuth(ctx), capabilityType).Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertCapabilityTypeRepresentation(result), nil
}

// SetCapabilityTypeDefaultModel sets the default model deployment for a capability type.
// Corresponds to PUT /v1/capability-types/{capability_type}.
func (c *Client) SetCapabilityTypeDefaultModel(ctx context.Context, capabilityType string, data DefaultModelDeploymentUpdate) (*CapabilityTypeRepresentation, error) {
	if strings.TrimSpace(capabilityType) == "" {
		return nil, fmt.Errorf("capabilityType cannot be empty")
	}

	genUpdate := api.NewDefaultModelDeploymentUpdate(data.DefaultModelDeploymentID)

	result, resp, err := c.generated.CapabilityTypesAPI.UpdateCapabilityTypeV1CapabilityTypesCapabilityTypePut(c.withAuth(ctx), capabilityType).
		DefaultModelDeploymentUpdate(*genUpdate).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertCapabilityTypeRepresentation(result), nil
}

// ListCapabilityTypes retrieves all capability type definitions.
// Corresponds to GET /v1/capability-types.
func (c *Client) ListCapabilityTypes(ctx context.Context) (*CapabilityTypesRepresentation, error) {
	result, resp, err := c.generated.CapabilityTypesAPI.ListCapabilityTypesV1CapabilityTypesGet(c.withAuth(ctx)).Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return convertCapabilityTypesRepresentation(result), nil
}
