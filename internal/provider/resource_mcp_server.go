// Copyright (c) Trifork

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-corax/internal/coraxclient"
)

var _ resource.Resource = &MCPServerResource{}
var _ resource.ResourceWithImportState = &MCPServerResource{}

func NewMCPServerResource() resource.Resource {
	return &MCPServerResource{}
}

type MCPServerResource struct {
	client *coraxclient.Client
}

// MCPServerResourceModel describes the resource data model.
type MCPServerResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	URL    types.String `tfsdk:"url"`
	Type   types.String `tfsdk:"type"`
	Config types.String `tfsdk:"config"` // JSON-encoded map
	Owner  types.String `tfsdk:"owner"`
	Slug   types.String `tfsdk:"slug"`
}

func (r *MCPServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_server"
}

func (r *MCPServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Corax MCP (Model Context Protocol) server. MCP servers expose tools, resources, and prompts that can be wired into capabilities.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for the MCP server (UUID).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for the MCP server. Up to 64 characters.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "URL the Corax backend uses to reach the MCP server.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Transport protocol. One of `streamablehttp` (default) or `sse`.",
				Validators: []validator.String{
					stringvalidator.OneOf("streamablehttp", "sse"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"config": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Server-specific configuration as a JSON-encoded object.",
			},
			"owner": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the user that owns the MCP server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"slug": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URL-safe slug derived from the name.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *MCPServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*coraxclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *coraxclient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	r.client = client
}

// parseConfigJSON decodes the config JSON string from a TF model into a map.
// Returns nil for null/unknown/empty config.
func parseConfigJSON(configField types.String) (map[string]interface{}, error) {
	if configField.IsNull() || configField.IsUnknown() {
		return nil, nil
	}
	raw := configField.ValueString()
	if raw == "" {
		return nil, nil
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("config is not a valid JSON object: %w", err)
	}
	return parsed, nil
}

// mapMCPServerToModel populates a TF model from the API response.
// Config is re-marshalled through encoding/json so map keys are sorted
// deterministically.
func mapMCPServerToModel(server *coraxclient.MCPServer, model *MCPServerResourceModel) error {
	model.ID = types.StringValue(server.ID)
	model.Name = types.StringValue(server.Name)
	model.URL = types.StringValue(server.URL)
	model.Type = types.StringValue(server.Type)
	model.Owner = types.StringValue(server.Owner)
	model.Slug = types.StringValue(server.Slug)

	if server.Config == nil {
		model.Config = types.StringNull()
		return nil
	}
	encoded, err := json.Marshal(server.Config)
	if err != nil {
		return fmt.Errorf("failed to encode config returned by API: %w", err)
	}
	model.Config = types.StringValue(string(encoded))
	return nil
}

func (r *MCPServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MCPServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configMap, err := parseConfigJSON(plan.Config)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid config", err.Error())
		return
	}

	payload := coraxclient.MCPServerCreate{
		Name:   plan.Name.ValueString(),
		URL:    plan.URL.ValueString(),
		Type:   plan.Type.ValueString(),
		Config: configMap,
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating MCP server: %s", payload.Name))
	created, err := r.client.CreateMCPServer(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create MCP server '%s': %s", payload.Name, err))
		return
	}

	if err := mapMCPServerToModel(created, &plan); err != nil {
		resp.Diagnostics.AddError("Mapping Error", err.Error())
		return
	}

	tflog.Info(ctx, fmt.Sprintf("MCP server %s created successfully with ID %s", plan.Name.ValueString(), plan.ID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MCPServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MCPServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Reading MCP server with ID: %s", serverID))

	server, err := r.client.GetMCPServer(ctx, serverID)
	if err != nil {
		if errors.Is(err, coraxclient.ErrNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("MCP server %s not found, removing from state", serverID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read MCP server '%s': %s", serverID, err))
		return
	}

	if err := mapMCPServerToModel(server, &state); err != nil {
		resp.Diagnostics.AddError("Mapping Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MCPServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MCPServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MCPServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := state.ID.ValueString()

	configMap, err := parseConfigJSON(plan.Config)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Invalid config", err.Error())
		return
	}

	payload := coraxclient.MCPServerUpdate{
		Name:   plan.Name.ValueString(),
		URL:    plan.URL.ValueString(),
		Type:   plan.Type.ValueString(),
		Config: configMap,
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating MCP server with ID: %s", serverID))
	updated, err := r.client.UpdateMCPServer(ctx, serverID, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update MCP server '%s': %s", serverID, err))
		return
	}

	if err := mapMCPServerToModel(updated, &plan); err != nil {
		resp.Diagnostics.AddError("Mapping Error", err.Error())
		return
	}

	tflog.Info(ctx, fmt.Sprintf("MCP server %s updated successfully", serverID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MCPServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MCPServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Deleting MCP server with ID: %s", serverID))

	err := r.client.DeleteMCPServer(ctx, serverID)
	if err != nil {
		if errors.Is(err, coraxclient.ErrNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("MCP server %s already deleted", serverID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete MCP server '%s': %s", serverID, err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("MCP server %s deleted successfully", serverID))
}

func (r *MCPServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
