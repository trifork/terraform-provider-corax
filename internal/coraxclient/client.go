// Copyright (c) Trifork

package coraxclient

import (
	"context"
	"fmt"
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
// Since there's no GET /v1/api-keys/{key_id} endpoint, we use the list endpoint
// with a filter to find the specific API key.
func (c *Client) GetAPIKey(ctx context.Context, keyID string) (*ApiKey, error) {
	if strings.TrimSpace(keyID) == "" {
		return nil, fmt.Errorf("keyID cannot be empty")
	}

	// Use the list endpoint with a filter to find the specific key
	filter := fmt.Sprintf("id::%s", keyID)
	result, resp, err := c.generated.APIKeysAPI.GetApiKeysV1ApiKeysGet(c.withAuth(ctx)).
		Filter(filter).
		Size(100). // Get more results in case filter doesn't work perfectly
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	embedded := result.GetEmbedded()
	if len(embedded) == 0 {
		return nil, ErrNotFound
	}

	// Find the key with the matching ID (in case filter returns multiple results)
	var genKey *api.ApiKey
	for i := range embedded {
		if embedded[i].GetId() == keyID {
			genKey = &embedded[i]
			break
		}
	}
	if genKey == nil {
		return nil, ErrNotFound
	}

	// Convert from generated ApiKey to our wrapper ApiKey
	apiKey := &ApiKey{
		ID:         genKey.GetId(),
		Name:       genKey.GetName(),
		Prefix:     genKey.GetPrefix(),
		IsActive:   genKey.GetIsActive(),
		UsageCount: int(genKey.GetUsageCount()),
		CreatedBy:  genKey.GetCreatedBy(),
		CreatedAt:  genKey.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	if expiresAt, ok := genKey.GetExpiresAtOk(); ok && expiresAt != nil {
		expStr := expiresAt.Format("2006-01-02T15:04:05Z07:00")
		apiKey.ExpiresAt = &expStr
	}
	if lastUsedAt, ok := genKey.GetLastUsedAtOk(); ok && lastUsedAt != nil {
		lastUsedStr := lastUsedAt.Format("2006-01-02T15:04:05Z07:00")
		apiKey.LastUsedAt = &lastUsedStr
	}
	if updatedAt, ok := genKey.GetUpdatedAtOk(); ok && updatedAt != nil {
		updatedStr := updatedAt.Format("2006-01-02T15:04:05Z07:00")
		apiKey.UpdatedAt = &updatedStr
	}

	return apiKey, nil
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

// CreateChatCapability creates a new chat capability.
// Corresponds to POST /v1/capabilities.
func (c *Client) CreateChatCapability(ctx context.Context, create api.ChatCapabilityCreate) (*api.ChatCapability, error) {
	cap1 := api.Capability1{ChatCapabilityCreate: &create}

	result, resp, err := c.generated.CapabilitiesAPI.CreateCapabilityV1CapabilitiesPost(c.withAuth(ctx)).
		Capability1(cap1).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	if result == nil {
		return nil, fmt.Errorf("nil response from create capability")
	}
	if result.ChatCapability != nil {
		return result.ChatCapability, nil
	}

	return nil, fmt.Errorf("expected ChatCapability in response but got a different type")
}

// CreateCompletionCapability creates a new completion capability.
// Corresponds to POST /v1/capabilities.
func (c *Client) CreateCompletionCapability(ctx context.Context, create api.CompletionCapabilityCreate) (*api.CompletionCapability, error) {
	cap1 := api.Capability1{CompletionCapabilityCreate: &create}

	result, resp, err := c.generated.CapabilitiesAPI.CreateCapabilityV1CapabilitiesPost(c.withAuth(ctx)).
		Capability1(cap1).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	if result == nil {
		return nil, fmt.Errorf("nil response from create capability")
	}
	if result.CompletionCapability != nil {
		return result.CompletionCapability, nil
	}

	return nil, fmt.Errorf("expected CompletionCapability in response but got a different type")
}

// GetCapability retrieves a specific capability by its ID.
// Returns the generic CapabilityRepresentation from the API.
// Corresponds to GET /v1/capabilities/{capability_id}.
func (c *Client) GetCapability(ctx context.Context, capabilityID string) (*api.CapabilityRepresentation, error) {
	if strings.TrimSpace(capabilityID) == "" {
		return nil, fmt.Errorf("capabilityID cannot be empty")
	}

	capId := api.CapabilityId{String: &capabilityID}

	result, resp, err := c.generated.CapabilitiesAPI.ReadCapabilityV1CapabilitiesCapabilityIdGet(c.withAuth(ctx), capId).Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return result, nil
}

// UpdateChatCapability updates a chat capability by its ID.
// Corresponds to PUT /v1/capabilities/{capability_id}.
func (c *Client) UpdateChatCapability(ctx context.Context, capabilityID string, update api.ChatCapabilityUpdate) (*api.CapabilityRepresentation, error) {
	if strings.TrimSpace(capabilityID) == "" {
		return nil, fmt.Errorf("capabilityID cannot be empty")
	}

	capId := api.CapabilityId1{String: &capabilityID}
	cap2 := api.Capability2{ChatCapabilityUpdate: &update}

	result, resp, err := c.generated.CapabilitiesAPI.UpdateCapabilityV1CapabilitiesCapabilityIdPut(c.withAuth(ctx), capId).
		Capability2(cap2).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return result, nil
}

// UpdateCompletionCapability updates a completion capability by its ID.
// Corresponds to PUT /v1/capabilities/{capability_id}.
func (c *Client) UpdateCompletionCapability(ctx context.Context, capabilityID string, update api.CompletionCapabilityUpdate) (*api.CapabilityRepresentation, error) {
	if strings.TrimSpace(capabilityID) == "" {
		return nil, fmt.Errorf("capabilityID cannot be empty")
	}

	capId := api.CapabilityId1{String: &capabilityID}
	cap2 := api.Capability2{CompletionCapabilityUpdate: &update}

	result, resp, err := c.generated.CapabilitiesAPI.UpdateCapabilityV1CapabilitiesCapabilityIdPut(c.withAuth(ctx), capId).
		Capability2(cap2).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return result, nil
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
		deploymentData.ProviderID,
	)
	genCreate.Configuration = convertConfigurationToGen(deploymentData.Configuration)

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
		deploymentData.ProviderID,
	)
	genUpdate.Configuration = convertConfigurationToGen(deploymentData.Configuration)

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

// GetCapabilityType retrieves a specific capability type definition.
// Corresponds to GET /v1/capability-types/{capability_type}.
func (c *Client) GetCapabilityType(ctx context.Context, capabilityType string) (*api.CapabilityTypeRepresentation, error) {
	if strings.TrimSpace(capabilityType) == "" {
		return nil, fmt.Errorf("capabilityType cannot be empty")
	}

	result, resp, err := c.generated.CapabilityTypesAPI.GetCapabilityTypeV1CapabilityTypesCapabilityTypeGet(c.withAuth(ctx), capabilityType).Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return result, nil
}

// SetCapabilityTypeDefaultModel sets the default model deployment for a capability type.
// Corresponds to PUT /v1/capability-types/{capability_type}.
func (c *Client) SetCapabilityTypeDefaultModel(ctx context.Context, capabilityType string, data api.DefaultModelDeploymentUpdate) (*api.CapabilityTypeRepresentation, error) {
	if strings.TrimSpace(capabilityType) == "" {
		return nil, fmt.Errorf("capabilityType cannot be empty")
	}

	result, resp, err := c.generated.CapabilityTypesAPI.UpdateCapabilityTypeV1CapabilityTypesCapabilityTypePut(c.withAuth(ctx), capabilityType).
		DefaultModelDeploymentUpdate(data).
		Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return result, nil
}

// ListCapabilityTypes retrieves all capability type definitions.
// Corresponds to GET /v1/capability-types.
func (c *Client) ListCapabilityTypes(ctx context.Context) (*api.CapabilityTypesRepresentation, error) {
	result, resp, err := c.generated.CapabilityTypesAPI.ListCapabilityTypesV1CapabilityTypesGet(c.withAuth(ctx)).Execute()

	if err != nil {
		return nil, convertError(err, resp)
	}

	return result, nil
}
