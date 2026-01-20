// Copyright (c) Trifork

//nolint:staticcheck // using json.Marshal on framework types for normalization
package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-corax/internal/coraxclient"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CompletionCapabilityResource{}
var _ resource.ResourceWithImportState = &CompletionCapabilityResource{}

func NewCompletionCapabilityResource() resource.Resource {
	return &CompletionCapabilityResource{}
}

// CompletionCapabilityResource defines the resource implementation.
type CompletionCapabilityResource struct {
	client *coraxclient.Client
}

// CompletionCapabilityResourceModel describes the resource data model.
type CompletionCapabilityResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	SemanticID       types.String `tfsdk:"semantic_id"` // Optional
	IsPublic         types.Bool   `tfsdk:"is_public"`
	ModelID          types.String `tfsdk:"model_id"`      // Nullable
	Config           types.Object `tfsdk:"config"`        // Nullable, uses CapabilityConfigModel from chat_capability.go
	ProjectID        types.String `tfsdk:"project_id"`    // Nullable
	SystemPrompt     types.String `tfsdk:"system_prompt"` // Shared with Chat, but also in Completion
	CompletionPrompt types.String `tfsdk:"completion_prompt"`
	Variables        types.Set    `tfsdk:"variables"`   // Nullable, set of strings
	OutputType       types.String `tfsdk:"output_type"` // "schema" or "text"
	SchemaDef        types.String `tfsdk:"schema_def"`  // Nullable, JSON string for structured output definition
	Owner            types.String `tfsdk:"owner"`       // Computed
	Type             types.String `tfsdk:"type"`        // Computed, should always be "completion"
	CreatedAt        types.String `tfsdk:"created_at"`  // Computed
	UpdatedAt        types.String `tfsdk:"updated_at"`  // Computed
	CreatedBy        types.String `tfsdk:"created_by"`  // Computed
	UpdatedBy        types.String `tfsdk:"updated_by"`  // Computed
}

// Note: CapabilityConfigModel, BlobConfigModel, DataRetentionModel, TimedDataRetentionModel, InfiniteDataRetentionModel
// are already defined in resource_chat_capability.go and can be reused.

func (r *CompletionCapabilityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_completion_capability"
}

func (r *CompletionCapabilityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Corax Completion Capability. Completion capabilities define configurations for generating text completions, potentially with structured output.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for the completion capability (UUID).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A user-defined name for the completion capability.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"semantic_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A semantic identifier for the completion capability that can be used for referencing.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"is_public": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Indicates whether the capability is publicly accessible. Defaults to false.",
			},
			"model_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The UUID of the model deployment to use for this capability. If not provided, a default model for 'completion' type may be used by the API.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The UUID of the project this capability belongs to.",
			},
			"system_prompt": schema.StringAttribute{
				Required:            true, // API spec shows this for CompletionCapability too
				MarkdownDescription: "The system prompt that provides context or instructions to the completion model.",
			},
			"completion_prompt": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The main prompt for which a completion is generated. May include placeholders for variables.",
			},
			"variables": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "A set of variable names (strings) that can be interpolated into the `completion_prompt`. Order is not significant.",
			},
			"output_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Defines the expected output format. Must be either 'text' or 'schema'.",
				Validators:          []validator.String{stringvalidator.OneOf("text", "schema")},
			},
			"schema_def": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Defines the structure of the output when `output_type` is 'schema'. A JSON-encoded string (use `jsonencode()`) defining the schema fields. Required if `output_type` is 'schema', must be null or omitted if `output_type` is 'text'.",
			},
			"config": schema.SingleNestedAttribute{ // Reusing the same config structure as chat
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Configuration settings for the capability's behavior.",
				Attributes:          capabilityConfigSchemaAttributes(), // Defined in chat_capability_resource.go (or move to a common place)
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			},
			"owner": schema.StringAttribute{Computed: true, MarkdownDescription: "Owner of the capability.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"type":  schema.StringAttribute{Computed: true, MarkdownDescription: "Type of the capability (should be 'completion').", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the capability was created (RFC3339 format).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the capability was last updated (RFC3339 format).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of who created the capability.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of who last updated the capability.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

// capabilityConfigSchemaAttributes, capabilityConfigModelToAPI, capabilityConfigAPItoModel
// and their underlying attribute type helpers are defined in common_capability_config.go
// No need to redefine them here.

// --- Helper functions for mapping (specific to Completion Capability) ---

// schemaDefToAPI converts a types.String (JSON string) to a map[string]interface{} for the API.
func schemaDefToAPI(ctx context.Context, schemaDef types.String, diags *diag.Diagnostics) map[string]interface{} {
	if schemaDef.IsNull() || schemaDef.IsUnknown() {
		return nil
	}

	jsonStr := schemaDef.ValueString()
	if jsonStr == "" {
		return nil
	}

	var goMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &goMap); err != nil {
		diags.AddError("SchemaDef JSON Parse Error",
			fmt.Sprintf("Failed to parse schema_def JSON: %s. Value: %s", err.Error(), jsonStr))
		return nil
	}

	return goMap
}

// schemaDefAPIToString converts API response map[string]interface{} to types.String (JSON string).
func schemaDefAPIToString(apiSchemaDef map[string]interface{}, diags *diag.Diagnostics) types.String {
	if apiSchemaDef == nil || len(apiSchemaDef) == 0 {
		return types.StringNull()
	}

	jsonBytes, err := json.Marshal(apiSchemaDef)
	if err != nil {
		diags.AddError("SchemaDef API Conversion Error",
			fmt.Sprintf("Failed to marshal schema_def from API to JSON: %s", err))
		return types.StringNull()
	}

	return types.StringValue(string(jsonBytes))
}

func mapAPICompletionCapabilityToModel(apiCap *coraxclient.CapabilityRepresentation, model *CompletionCapabilityResourceModel, diags *diag.Diagnostics, ctx context.Context) {
	model.ID = types.StringValue(apiCap.ID)
	model.SemanticID = types.StringValue(apiCap.SemanticID)
	model.Name = types.StringValue(apiCap.Name)
	model.IsPublic = types.BoolValue(apiCap.IsPublic != nil && *apiCap.IsPublic)
	model.Type = types.StringValue(apiCap.Type)

	if apiCap.ModelID != nil {
		model.ModelID = types.StringValue(*apiCap.ModelID)
	} else {
		model.ModelID = types.StringNull()
	}
	if apiCap.ProjectID != nil {
		model.ProjectID = types.StringValue(*apiCap.ProjectID)
	} else {
		model.ProjectID = types.StringNull()
	}

	// Populate SystemPrompt and CompletionPrompt from apiCap.Configuration
	if apiCap.Configuration != nil {
		if sysPrompt, ok := apiCap.Configuration["system_prompt"].(string); ok {
			model.SystemPrompt = types.StringValue(sysPrompt)
		} else {
			// If key is missing or not a string, treat as unknown.
			// Per schema, system_prompt is required, so Unknown is appropriate if not found/convertible.
			model.SystemPrompt = types.StringUnknown()
		}

		if compPrompt, ok := apiCap.Configuration["completion_prompt"].(string); ok {
			model.CompletionPrompt = types.StringValue(compPrompt)
		} else {
			// Per schema, completion_prompt is required.
			model.CompletionPrompt = types.StringUnknown()
		}
	} else {
		// apiCap.Configuration map itself is nil
		model.SystemPrompt = types.StringUnknown()
		model.CompletionPrompt = types.StringUnknown()
		tflog.Debug(ctx, fmt.Sprintf("apiCap.Configuration is nil for capability %s. SystemPrompt and CompletionPrompt will be unknown.", apiCap.ID))
	}

	// Populate OutputType and SchemaDef from apiCap.Output
	if apiCap.Output != nil {
		if outputTypeVal, ok := apiCap.Output["type"].(string); ok {
			model.OutputType = types.StringValue(outputTypeVal)
		} else {
			// Per schema, output_type is required.
			model.OutputType = types.StringUnknown()
		}

		// schema_def is sourced from apiCap.Output["result"]
		// Only populate schema_def if output_type is "schema" - for "text" type, we ignore the API value
		// because users don't set it in HCL and we don't want drift.
		outputType := model.OutputType.ValueString()
		if outputType == "schema" {
			if schemaDefVal, ok := apiCap.Output["result"].(map[string]interface{}); ok {
				model.SchemaDef = schemaDefAPIToString(schemaDefVal, diags)
			} else {
				// If "result" is not found, or not a map[string]interface{}, treat SchemaDef as null.
				if _, found := apiCap.Output["result"]; found && !ok {
					diags.AddAttributeWarning(
						path.Root("schema_def"),
						"Invalid Type for Schema Definition",
						fmt.Sprintf("Expected 'result' in API output to be a map, but got %T. Treating schema_def as null.", apiCap.Output["result"]),
					)
				}
				model.SchemaDef = types.StringNull()
			}
		} else {
			// For "text" output type, schema_def should be null
			model.SchemaDef = types.StringNull()
		}
	} else {
		// apiCap.Output map itself is nil
		model.OutputType = types.StringUnknown()
		model.SchemaDef = types.StringNull()
		tflog.Debug(ctx, fmt.Sprintf("apiCap.Output is nil for capability %s. OutputType will be unknown and SchemaDef null.", apiCap.ID))
	}

	// Populate Variables from apiCap.Input
	if apiCap.Input != nil {
		if varsData, found := apiCap.Input["variables"]; found && varsData != nil {
			if vars, ok := varsData.([]interface{}); ok {
				// If empty array, treat as null to avoid drift when user doesn't set variables
				if len(vars) == 0 {
					model.Variables = types.SetNull(types.StringType)
				} else {
					strVars := make([]string, len(vars))
					allStrings := true
					for i, v := range vars {
						if strV, isString := v.(string); isString {
							strVars[i] = strV
						} else {
							allStrings = false
							diags.AddAttributeWarning(
								path.Root("variables"),
								"Invalid Variable Type in API Response",
								fmt.Sprintf("Variable at index %d is not a string (actual type: %T). Treating variables as null.", i, v),
							)
							break
						}
					}
					if allStrings {
						setValue, conversionDiags := types.SetValueFrom(ctx, types.StringType, strVars)
						diags.Append(conversionDiags...)
						if !conversionDiags.HasError() {
							model.Variables = setValue
						} else {
							model.Variables = types.SetNull(types.StringType)
						}
					} else {
						model.Variables = types.SetNull(types.StringType)
					}
				}
			} else if varsMap, ok := varsData.(map[string]interface{}); ok { // Handle map from GET
				// If empty map, treat as null to avoid drift when user doesn't set variables
				if len(varsMap) == 0 {
					model.Variables = types.SetNull(types.StringType)
				} else {
					strVarKeys := make([]string, 0, len(varsMap))
					for k := range varsMap {
						strVarKeys = append(strVarKeys, k)
					}

					setValue, conversionDiags := types.SetValueFrom(ctx, types.StringType, strVarKeys)
					diags.Append(conversionDiags...)
					if !conversionDiags.HasError() {
						model.Variables = setValue
					} else {
						model.Variables = types.SetNull(types.StringType)
						diags.AddAttributeError(
							path.Root("variables"),
							"Variable Conversion Error (Map to Set)",
							fmt.Sprintf("Failed to convert variable keys from API map to set: %v", conversionDiags),
						)
					}
				}
			} else { // apiCap.Input["variables"] is present but not []interface{} and not map[string]interface{}
				diags.AddAttributeWarning(
					path.Root("variables"),
					"Incorrect Type for Variables in API Response",
					fmt.Sprintf("Expected 'variables' in API input to be a list or map of strings, but got %T. Treating variables as null.", varsData),
				)
				model.Variables = types.SetNull(types.StringType)
			}
		} else { // "variables" key not found in apiCap.Input or its value is JSON null
			if model.Variables.IsNull() || model.Variables.IsUnknown() {
				model.Variables = types.SetNull(types.StringType)
			}
		}
	} else { // apiCap.Input map itself is nil
		if model.Variables.IsNull() || model.Variables.IsUnknown() {
			model.Variables = types.SetNull(types.StringType)
		}
		tflog.Debug(ctx, fmt.Sprintf("apiCap.Input is nil for capability %s. Variables will be null.", apiCap.ID))
	}

	model.Config = capabilityConfigAPItoModel(ctx, apiCap.Config, diags) // Common config

	model.Owner = types.StringValue(apiCap.Owner)
	model.CreatedAt = types.StringValue(apiCap.CreatedAt)
	model.UpdatedAt = types.StringValue(apiCap.UpdatedAt)
	model.CreatedBy = types.StringValue(apiCap.CreatedBy)
	model.UpdatedBy = types.StringValue(apiCap.UpdatedBy)
}

func (r *CompletionCapabilityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CompletionCapabilityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CompletionCapabilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating Completion Capability: %s", plan.Name.ValueString()))

	apiPayload := coraxclient.CompletionCapabilityCreate{
		Name:             plan.Name.ValueString(),
		Type:             "completion", // Hardcoded
		SystemPrompt:     plan.SystemPrompt.ValueString(),
		CompletionPrompt: plan.CompletionPrompt.ValueString(),
		OutputType:       plan.OutputType.ValueString(),
	}

	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		isPublic := plan.IsPublic.ValueBool()
		apiPayload.IsPublic = &isPublic
	}
	if !plan.SemanticID.IsNull() && !plan.SemanticID.IsUnknown() {
		semanticID := plan.SemanticID.ValueString()
		apiPayload.SemanticID = &semanticID
	}
	if !plan.ModelID.IsNull() && !plan.ModelID.IsUnknown() {
		modelID := plan.ModelID.ValueString()
		apiPayload.ModelID = &modelID
	}
	if !plan.ProjectID.IsNull() && !plan.ProjectID.IsUnknown() {
		projectID := plan.ProjectID.ValueString()
		apiPayload.ProjectID = &projectID
	}
	if !plan.Variables.IsNull() && !plan.Variables.IsUnknown() {
		resp.Diagnostics.Append(plan.Variables.ElementsAs(ctx, &apiPayload.Variables, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	outputType := plan.OutputType.ValueString()
	if outputType == "schema" {
		if plan.SchemaDef.IsNull() || plan.SchemaDef.IsUnknown() {
			resp.Diagnostics.AddError("Validation Error", "schema_def is required when output_type is 'schema'")
			return
		}
		apiPayload.SchemaDef = schemaDefToAPI(ctx, plan.SchemaDef, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Create: schema_def converted to: %+v", apiPayload.SchemaDef))
		if apiPayload.SchemaDef == nil || len(apiPayload.SchemaDef) == 0 {
			resp.Diagnostics.AddError("SchemaDef Conversion Error", "schema_def was provided but conversion resulted in nil or empty map")
			return
		}
	} else if outputType == "text" {
		if !plan.SchemaDef.IsNull() && !plan.SchemaDef.IsUnknown() {
			resp.Diagnostics.AddError("Validation Error", "schema_def must not be set when output_type is 'text'")
			return
		}
	} else {
		resp.Diagnostics.AddError("Validation Error", fmt.Sprintf("unsupported output_type '%s', must be either 'text' or 'schema'", outputType))
		return
	}

	// Common config mapping (reuse from chat capability if moved to common, or define here)
	// For now, assuming capabilityConfigModelToAPI is available (defined in chat_capability.go or common)
	apiPayload.Config = capabilityConfigModelToAPI(ctx, plan.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createdAPICap, err := r.client.CreateCapability(ctx, apiPayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create completion capability, got error: %s", err))
		return
	}

	mapAPICompletionCapabilityToModel(createdAPICap, &plan, &resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Completion Capability %s created successfully with ID %s", plan.Name.ValueString(), plan.ID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CompletionCapabilityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CompletionCapabilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	capabilityID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Reading Completion Capability with ID: %s", capabilityID))

	apiCap, err := r.client.GetCapability(ctx, capabilityID)
	if err != nil {
		if errors.Is(err, coraxclient.ErrNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Completion Capability %s not found, removing from state", capabilityID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read completion capability %s: %s", capabilityID, err))
		return
	}

	if apiCap.Type != "completion" {
		resp.Diagnostics.AddError("Resource Type Mismatch", fmt.Sprintf("Expected capability type 'completion' but found '%s' for ID %s. Removing from state.", apiCap.Type, capabilityID))
		resp.State.RemoveResource(ctx)
		return
	}

	mapAPICompletionCapabilityToModel(apiCap, &state, &resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read Completion Capability %s", capabilityID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CompletionCapabilityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CompletionCapabilityResourceModel
	var state CompletionCapabilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	capabilityID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Updating Completion Capability with ID: %s using full plan payload", capabilityID))

	// --- Construct full update payload from plan ---
	nameValue := plan.Name.ValueString()
	typeValue := "completion" // Type is fixed for this resource
	systemPromptValue := plan.SystemPrompt.ValueString()
	completionPromptValue := plan.CompletionPrompt.ValueString()
	outputTypeValue := plan.OutputType.ValueString()

	updatePayload := coraxclient.CompletionCapabilityUpdate{
		Name:             &nameValue,
		Type:             &typeValue,
		SystemPrompt:     &systemPromptValue,
		CompletionPrompt: &completionPromptValue,
		OutputType:       &outputTypeValue,
	}

	// IsPublic
	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		isPublicVal := plan.IsPublic.ValueBool()
		updatePayload.IsPublic = &isPublicVal
	} else {
		defaultIsPublic := false // As per schema default
		updatePayload.IsPublic = &defaultIsPublic
	}

	// SemanticID
	if !plan.SemanticID.IsNull() && !plan.SemanticID.IsUnknown() {
		semanticIDVal := plan.SemanticID.ValueString()
		updatePayload.SemanticID = &semanticIDVal
	} else {
		updatePayload.SemanticID = nil
	}

	// ModelID
	if !plan.ModelID.IsNull() && !plan.ModelID.IsUnknown() {
		modelIDVal := plan.ModelID.ValueString()
		updatePayload.ModelID = &modelIDVal
	} else {
		updatePayload.ModelID = nil
	}

	// ProjectID
	if !plan.ProjectID.IsNull() && !plan.ProjectID.IsUnknown() {
		projectIDVal := plan.ProjectID.ValueString()
		updatePayload.ProjectID = &projectIDVal
	} else {
		updatePayload.ProjectID = nil
	}

	// Variables
	if !plan.Variables.IsNull() && !plan.Variables.IsUnknown() {
		var vars []string
		resp.Diagnostics.Append(plan.Variables.ElementsAs(ctx, &vars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updatePayload.Variables = vars // Assign directly, omitempty handles if vars is nil/empty based on API needs
	} else {
		// If API expects an empty list to clear, send []string{}. If omitempty on nil is preferred, send nil.
		// Assuming omitempty on nil is fine for now.
		updatePayload.Variables = nil
	}

	// SchemaDef
	if outputTypeValue == "schema" {
		if plan.SchemaDef.IsNull() || plan.SchemaDef.IsUnknown() {
			resp.Diagnostics.AddError("Validation Error", "schema_def is required when output_type is 'schema'")
			return
		}
		updatePayload.SchemaDef = schemaDefToAPI(ctx, plan.SchemaDef, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	} else if outputTypeValue == "text" {
		if !plan.SchemaDef.IsNull() && !plan.SchemaDef.IsUnknown() {
			resp.Diagnostics.AddError("Validation Error", "schema_def must not be set when output_type is 'text'")
			return
		}
		updatePayload.SchemaDef = nil
	} else {
		resp.Diagnostics.AddError("Validation Error", fmt.Sprintf("unsupported output_type '%s', must be either 'text' or 'schema'", outputTypeValue))
		return
	}

	// Config
	updatePayload.Config = capabilityConfigModelToAPI(ctx, plan.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// --- End of payload construction ---

	updatedAPICap, err := r.client.UpdateCapability(ctx, capabilityID, updatePayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update completion capability %s: %s", capabilityID, err))
		return
	}

	mapAPICompletionCapabilityToModel(updatedAPICap, &plan, &resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve immutable computed fields from state to avoid "inconsistent result" errors
	// caused by timestamp precision differences or server-side timing.
	// The next Read operation will refresh these from the API.
	plan.CreatedAt = state.CreatedAt
	plan.CreatedBy = state.CreatedBy
	plan.UpdatedAt = state.UpdatedAt
	plan.UpdatedBy = state.UpdatedBy

	tflog.Info(ctx, fmt.Sprintf("Completion Capability %s updated successfully", capabilityID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CompletionCapabilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CompletionCapabilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	capabilityID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Deleting Completion Capability with ID: %s", capabilityID))

	err := r.client.DeleteCapability(ctx, capabilityID)
	if err != nil {
		if errors.Is(err, coraxclient.ErrNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Completion Capability %s not found, already deleted", capabilityID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete completion capability %s: %s", capabilityID, err))
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Completion Capability %s deleted successfully", capabilityID))
}

func (r *CompletionCapabilityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
