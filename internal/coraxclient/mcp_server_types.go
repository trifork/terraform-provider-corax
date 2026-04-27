// Copyright (c) Trifork

package coraxclient

// MCPServer represents the MCP (Model Context Protocol) server response.
// Based on openapi.json components.schemas.MCPServerResponse.
type MCPServer struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	URL    string                 `json:"url"`
	Type   string                 `json:"type,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
	Owner  string                 `json:"owner"`
	Slug   string                 `json:"slug"`
}

// MCPServerCreate represents the request body for creating an MCP server.
// Based on openapi.json components.schemas.MCPServerBase.
type MCPServerCreate struct {
	Name   string                 `json:"name"`
	URL    string                 `json:"url"`
	Type   string                 `json:"type,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// MCPServerUpdate represents the request body for updating an MCP server.
// Same shape as MCPServerCreate (the API reuses MCPServerBase for PUT).
type MCPServerUpdate struct {
	Name   string                 `json:"name"`
	URL    string                 `json:"url"`
	Type   string                 `json:"type,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}
