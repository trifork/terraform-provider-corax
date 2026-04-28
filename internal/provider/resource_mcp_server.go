// Copyright (c) Trifork

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
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

// mcpServerConfigEntry is a single config entry — keyed by the user-defined
// config name (e.g. "token", "filters") in the parent map.
type mcpServerConfigEntry struct {
	Type     types.String `tfsdk:"type"`
	Label    types.String `tfsdk:"label"`
	Default  types.String `tfsdk:"default"`
	Required types.Bool   `tfsdk:"required"`
}

// configEntryAttrTypes mirrors the schema attribute types for one entry.
// Used when constructing types.Map values from API responses.
var configEntryAttrTypes = map[string]attr.Type{
	"type":     types.StringType,
	"label":    types.StringType,
	"default":  types.StringType,
	"required": types.BoolType,
}

// MCPServerResourceModel describes the resource data model.
type MCPServerResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	URL      types.String `tfsdk:"url"`
	Type     types.String `tfsdk:"type"`
	Config   types.Map    `tfsdk:"config"`
	IsPublic types.Bool   `tfsdk:"is_public"`
	Owner    types.String `tfsdk:"owner"`
	Slug     types.String `tfsdk:"slug"`
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
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1), stringvalidator.LengthAtMost(64)},
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
				Validators:          []validator.String{stringvalidator.OneOf("streamablehttp", "sse")},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"is_public": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the server is publicly accessible. Defaults to false.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"config": schema.MapNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Per-binding configuration, keyed by config name (e.g. `token`, `filters`). Each entry describes how a header or parameter is supplied to the MCP server.",
				PlanModifiers:       []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Binding type, e.g. `header`.",
						},
						"label": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "User-facing label or, for headers, the header name (e.g. `Authorization`, `X-Filters`).",
						},
						"default": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Default value when the caller does not supply one. Pass `null` for no default.",
						},
						"required": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Whether the caller must supply this value. Server default is true.",
						},
					},
				},
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

// configMapToAPI converts the TF nested-map config to the API's
// map[string]interface{} payload. Returns nil when the map is null/unknown.
func configMapToAPI(ctx context.Context, configMap types.Map) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	if configMap.IsNull() || configMap.IsUnknown() {
		return nil, diags
	}

	entries := map[string]mcpServerConfigEntry{}
	diags.Append(configMap.ElementsAs(ctx, &entries, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make(map[string]interface{}, len(entries))
	for key, entry := range entries {
		obj := map[string]interface{}{
			"type":  entry.Type.ValueString(),
			"label": entry.Label.ValueString(),
		}
		if entry.Default.IsNull() || entry.Default.IsUnknown() {
			obj["default"] = nil
		} else {
			obj["default"] = entry.Default.ValueString()
		}
		if !entry.Required.IsNull() && !entry.Required.IsUnknown() {
			obj["required"] = entry.Required.ValueBool()
		}
		out[key] = obj
	}
	return out, diags
}

// configMapFromAPI builds a types.Map from the API response. Best-effort
// type coercion: unknown shapes fall back to null fields.
func configMapFromAPI(apiConfig map[string]interface{}) (types.Map, diag.Diagnostics) {
	objectType := types.ObjectType{AttrTypes: configEntryAttrTypes}

	if apiConfig == nil {
		return types.MapNull(objectType), nil
	}

	values := make(map[string]attr.Value, len(apiConfig))
	var diags diag.Diagnostics

	for key, raw := range apiConfig {
		entryMap, ok := raw.(map[string]interface{})
		if !ok {
			diags.AddError("Unexpected config shape", fmt.Sprintf("config entry %q is not a JSON object", key))
			continue
		}

		attrs := map[string]attr.Value{
			"type":     stringFromAny(entryMap["type"]),
			"label":    stringFromAny(entryMap["label"]),
			"default":  stringFromAnyNullable(entryMap["default"]),
			"required": boolFromAny(entryMap["required"]),
		}
		obj, objDiags := types.ObjectValue(configEntryAttrTypes, attrs)
		diags.Append(objDiags...)
		if objDiags.HasError() {
			continue
		}
		values[key] = obj
	}

	if diags.HasError() {
		return types.MapNull(objectType), diags
	}

	mapVal, mapDiags := types.MapValue(objectType, values)
	diags.Append(mapDiags...)
	return mapVal, diags
}

func stringFromAny(v interface{}) types.String {
	if v == nil {
		return types.StringNull()
	}
	if s, ok := v.(string); ok {
		return types.StringValue(s)
	}
	return types.StringValue(fmt.Sprintf("%v", v))
}

func stringFromAnyNullable(v interface{}) types.String {
	if v == nil {
		return types.StringNull()
	}
	if s, ok := v.(string); ok {
		return types.StringValue(s)
	}
	return types.StringValue(fmt.Sprintf("%v", v))
}

func boolFromAny(v interface{}) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}

// mapMCPServerToModel populates a TF model from the API response.
func mapMCPServerToModel(server *coraxclient.MCPServer, model *MCPServerResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(server.ID)
	model.Name = types.StringValue(server.Name)
	model.URL = types.StringValue(server.URL)
	model.Type = types.StringValue(server.Type)
	model.IsPublic = types.BoolValue(server.IsPublic)
	model.Owner = types.StringValue(server.Owner)
	model.Slug = types.StringValue(server.Slug)

	cfgVal, cfgDiags := configMapFromAPI(server.Config)
	diags.Append(cfgDiags...)
	model.Config = cfgVal

	return diags
}

func (r *MCPServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MCPServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configMap, cfgDiags := configMapToAPI(ctx, plan.Config)
	resp.Diagnostics.Append(cfgDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := coraxclient.MCPServerCreate{
		Name:   plan.Name.ValueString(),
		URL:    plan.URL.ValueString(),
		Type:   plan.Type.ValueString(),
		Config: configMap,
	}
	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		isPublic := plan.IsPublic.ValueBool()
		payload.IsPublic = &isPublic
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating MCP server: %s", payload.Name))
	created, err := r.client.CreateMCPServer(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create MCP server '%s': %s", payload.Name, err))
		return
	}

	resp.Diagnostics.Append(mapMCPServerToModel(created, &plan)...)
	if resp.Diagnostics.HasError() {
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

	resp.Diagnostics.Append(mapMCPServerToModel(server, &state)...)
	if resp.Diagnostics.HasError() {
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

	configMap, cfgDiags := configMapToAPI(ctx, plan.Config)
	resp.Diagnostics.Append(cfgDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := coraxclient.MCPServerUpdate{
		Name:   plan.Name.ValueString(),
		URL:    plan.URL.ValueString(),
		Type:   plan.Type.ValueString(),
		Config: configMap,
	}
	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		isPublic := plan.IsPublic.ValueBool()
		payload.IsPublic = &isPublic
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating MCP server with ID: %s", serverID))
	updated, err := r.client.UpdateMCPServer(ctx, serverID, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update MCP server '%s': %s", serverID, err))
		return
	}

	resp.Diagnostics.Append(mapMCPServerToModel(updated, &plan)...)
	if resp.Diagnostics.HasError() {
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
