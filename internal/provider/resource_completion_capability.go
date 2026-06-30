// Copyright (c) Trifork

//nolint:staticcheck // using json.Marshal on framework types for normalization
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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
	api "terraform-provider-corax/internal/generated"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CompletionCapabilityResource{}
var _ resource.ResourceWithImportState = &CompletionCapabilityResource{}
var _ resource.ResourceWithUpgradeState = &CompletionCapabilityResource{}

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
		// Version 1 introduces the schema_def state upgrade (pre-1.0 dynamic value -> JSON string).
		Version:             1,
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
func schemaDefToAPI(_ context.Context, schemaDef types.String, diags *diag.Diagnostics) map[string]interface{} {
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

// mapCompletionCapabilityRepresentationToModel maps an api.CapabilityRepresentation (from Get/Update) to the TF model.
func mapCompletionCapabilityRepresentationToModel(apiCap *api.CapabilityRepresentation, model *CompletionCapabilityResourceModel, diags *diag.Diagnostics, ctx context.Context) {
	model.ID = types.StringValue(apiCap.Id)
	model.SemanticID = types.StringValue(apiCap.SemanticId)
	model.Name = types.StringValue(apiCap.Name)
	model.IsPublic = types.BoolValue(apiCap.GetIsPublic())
	model.Type = types.StringValue(apiCap.Type)

	if modelId, ok := apiCap.GetModelIdOk(); ok && modelId != nil {
		model.ModelID = types.StringValue(*modelId)
	} else {
		model.ModelID = types.StringNull()
	}
	if projectId, ok := apiCap.GetProjectIdOk(); ok && projectId != nil {
		model.ProjectID = types.StringValue(*projectId)
	} else {
		model.ProjectID = types.StringNull()
	}

	// Populate SystemPrompt and CompletionPrompt from apiCap.Configuration
	if apiCap.Configuration != nil {
		if sysPrompt, ok := apiCap.Configuration["system_prompt"].(string); ok {
			model.SystemPrompt = types.StringValue(sysPrompt)
		} else {
			model.SystemPrompt = types.StringUnknown()
		}

		if compPrompt, ok := apiCap.Configuration["completion_prompt"].(string); ok {
			model.CompletionPrompt = types.StringValue(compPrompt)
		} else {
			model.CompletionPrompt = types.StringUnknown()
		}
	} else {
		model.SystemPrompt = types.StringUnknown()
		model.CompletionPrompt = types.StringUnknown()
		tflog.Debug(ctx, fmt.Sprintf("apiCap.Configuration is nil for capability %s. SystemPrompt and CompletionPrompt will be unknown.", apiCap.Id))
	}

	// Populate OutputType and SchemaDef from apiCap.Output
	if apiCap.Output != nil {
		if outputTypeVal, ok := apiCap.Output["type"].(string); ok {
			model.OutputType = types.StringValue(outputTypeVal)
		} else {
			model.OutputType = types.StringUnknown()
		}

		outputType := model.OutputType.ValueString()
		if outputType == "schema" {
			if schemaDefVal, ok := apiCap.Output["result"].(map[string]interface{}); ok {
				model.SchemaDef = schemaDefAPIToString(schemaDefVal, diags)
			} else {
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
			model.SchemaDef = types.StringNull()
		}
	} else {
		model.OutputType = types.StringUnknown()
		model.SchemaDef = types.StringNull()
		tflog.Debug(ctx, fmt.Sprintf("apiCap.Output is nil for capability %s. OutputType will be unknown and SchemaDef null.", apiCap.Id))
	}

	// Populate Variables from apiCap.Input
	if apiCap.Input != nil {
		if varsData, found := apiCap.Input["variables"]; found && varsData != nil {
			if vars, ok := varsData.([]interface{}); ok {
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
			} else if varsMap, ok := varsData.(map[string]interface{}); ok {
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
			} else if vars, ok := varsData.([]string); ok {
				if len(vars) == 0 {
					model.Variables = types.SetNull(types.StringType)
				} else {
					setValue, conversionDiags := types.SetValueFrom(ctx, types.StringType, vars)
					diags.Append(conversionDiags...)
					if !conversionDiags.HasError() {
						model.Variables = setValue
					} else {
						model.Variables = types.SetNull(types.StringType)
					}
				}
			} else {
				diags.AddAttributeWarning(
					path.Root("variables"),
					"Incorrect Type for Variables in API Response",
					fmt.Sprintf("Expected 'variables' in API input to be a list or map of strings, but got %T. Treating variables as null.", varsData),
				)
				model.Variables = types.SetNull(types.StringType)
			}
		} else {
			if model.Variables.IsNull() || model.Variables.IsUnknown() {
				model.Variables = types.SetNull(types.StringType)
			}
		}
	} else {
		if model.Variables.IsNull() || model.Variables.IsUnknown() {
			model.Variables = types.SetNull(types.StringType)
		}
		tflog.Debug(ctx, fmt.Sprintf("apiCap.Input is nil for capability %s. Variables will be null.", apiCap.Id))
	}

	// Extract config from NullableCapabilityConfig
	var cfgPtr *api.CapabilityConfig
	if configVal, ok := apiCap.GetConfigOk(); ok {
		cfgPtr = configVal
	}
	model.Config = capabilityConfigAPItoModel(ctx, cfgPtr, diags)

	model.Owner = types.StringValue(apiCap.Owner)
	model.CreatedAt = types.StringValue(apiCap.CreatedAt.Format(time.RFC3339))
	model.UpdatedAt = types.StringValue(apiCap.UpdatedAt.Format(time.RFC3339))
	model.CreatedBy = types.StringValue(apiCap.CreatedBy)
	model.UpdatedBy = types.StringValue(apiCap.UpdatedBy)
}

// mapCompletionCapabilityCreateResponseToModel maps an api.CompletionCapability (from Create) to the TF model.
func mapCompletionCapabilityCreateResponseToModel(apiCap *api.CompletionCapability, model *CompletionCapabilityResourceModel, diags *diag.Diagnostics, ctx context.Context) {
	model.ID = types.StringValue(apiCap.Id)
	model.Name = types.StringValue(apiCap.Name)
	model.IsPublic = types.BoolValue(apiCap.GetIsPublic())

	if apiCap.Type != nil {
		model.Type = types.StringValue(*apiCap.Type)
	} else {
		model.Type = types.StringValue("completion")
	}

	if modelId, ok := apiCap.GetModelIdOk(); ok && modelId != nil {
		model.ModelID = types.StringValue(*modelId)
	} else {
		model.ModelID = types.StringNull()
	}
	if projectId, ok := apiCap.GetProjectIdOk(); ok && projectId != nil {
		model.ProjectID = types.StringValue(*projectId)
	} else {
		model.ProjectID = types.StringNull()
	}
	if semanticId, ok := apiCap.GetSemanticIdOk(); ok && semanticId != nil {
		model.SemanticID = types.StringValue(*semanticId)
	} else {
		model.SemanticID = types.StringValue("")
	}

	model.SystemPrompt = types.StringValue(apiCap.SystemPrompt)
	model.CompletionPrompt = types.StringValue(apiCap.CompletionPrompt)
	model.OutputType = types.StringValue(apiCap.OutputType)

	// Variables from typed []string
	if len(apiCap.Variables) == 0 {
		model.Variables = types.SetNull(types.StringType)
	} else {
		setValue, conversionDiags := types.SetValueFrom(ctx, types.StringType, apiCap.Variables)
		diags.Append(conversionDiags...)
		if !conversionDiags.HasError() {
			model.Variables = setValue
		} else {
			model.Variables = types.SetNull(types.StringType)
		}
	}

	// SchemaDef from typed map[string]CompletionCapabilitySchemaDefValue
	if apiCap.OutputType == "schema" && len(apiCap.SchemaDef) > 0 {
		// Convert typed schema def to map[string]interface{} then to JSON string
		genericMap := make(map[string]interface{}, len(apiCap.SchemaDef))
		for k, v := range apiCap.SchemaDef {
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				diags.AddError("SchemaDef Conversion Error",
					fmt.Sprintf("Failed to marshal schema_def value for key '%s': %s", k, err))
				model.SchemaDef = types.StringNull()
				break
			}
			var generic interface{}
			if err := json.Unmarshal(jsonBytes, &generic); err != nil {
				diags.AddError("SchemaDef Conversion Error",
					fmt.Sprintf("Failed to unmarshal schema_def value for key '%s': %s", k, err))
				model.SchemaDef = types.StringNull()
				break
			}
			genericMap[k] = generic
		}
		if !diags.HasError() {
			model.SchemaDef = schemaDefAPIToString(genericMap, diags)
		}
	} else {
		model.SchemaDef = types.StringNull()
	}

	// Extract config from NullableCapabilityConfig
	var cfgPtr *api.CapabilityConfig
	if configVal, ok := apiCap.GetConfigOk(); ok {
		cfgPtr = configVal
	}
	model.Config = capabilityConfigAPItoModel(ctx, cfgPtr, diags)

	model.Owner = types.StringValue(apiCap.Owner)
	model.CreatedAt = types.StringValue(apiCap.CreatedAt.Format(time.RFC3339))
	model.UpdatedAt = types.StringValue(apiCap.UpdatedAt.Format(time.RFC3339))
	model.CreatedBy = types.StringValue(apiCap.CreatedBy)
	model.UpdatedBy = types.StringValue(apiCap.UpdatedBy)
}

// schemaDefToTypedAPI converts a map[string]interface{} to map[string]api.CompletionCapabilityCreateSchemaDefValue.
// Each value is marshaled to JSON and then unmarshaled into the union type which has UnmarshalJSON
// that tries ArrayPropertyInput, BasicProperty, EnumProperty, ObjectPropertyInput.
func schemaDefToTypedAPI(genericMap map[string]interface{}, diags *diag.Diagnostics) map[string]api.CompletionCapabilityCreateSchemaDefValue {
	if genericMap == nil || len(genericMap) == 0 {
		return nil
	}
	result := make(map[string]api.CompletionCapabilityCreateSchemaDefValue, len(genericMap))
	for k, v := range genericMap {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			diags.AddError("SchemaDef Conversion Error",
				fmt.Sprintf("Failed to marshal schema_def value for key '%s': %s", k, err))
			return nil
		}
		var typedVal api.CompletionCapabilityCreateSchemaDefValue
		if err := json.Unmarshal(jsonBytes, &typedVal); err != nil {
			diags.AddError("SchemaDef Conversion Error",
				fmt.Sprintf("Failed to unmarshal schema_def value for key '%s' into API type: %s", k, err))
			return nil
		}
		result[k] = typedVal
	}
	return result
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

	apiPayload := api.NewCompletionCapabilityCreate(
		plan.Name.ValueString(),
		"completion",
		plan.SystemPrompt.ValueString(),
		plan.CompletionPrompt.ValueString(),
		plan.OutputType.ValueString(),
	)

	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		apiPayload.SetIsPublic(plan.IsPublic.ValueBool())
	}
	if !plan.SemanticID.IsNull() && !plan.SemanticID.IsUnknown() {
		apiPayload.SetSemanticId(plan.SemanticID.ValueString())
	}
	if !plan.ModelID.IsNull() && !plan.ModelID.IsUnknown() {
		apiPayload.SetModelId(plan.ModelID.ValueString())
	}
	if !plan.ProjectID.IsNull() && !plan.ProjectID.IsUnknown() {
		apiPayload.SetProjectId(plan.ProjectID.ValueString())
	}
	if !plan.Variables.IsNull() && !plan.Variables.IsUnknown() {
		var vars []string
		resp.Diagnostics.Append(plan.Variables.ElementsAs(ctx, &vars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiPayload.SetVariables(vars)
	}
	outputType := plan.OutputType.ValueString()
	if outputType == "schema" {
		if plan.SchemaDef.IsNull() || plan.SchemaDef.IsUnknown() {
			resp.Diagnostics.AddError("Validation Error", "schema_def is required when output_type is 'schema'")
			return
		}
		genericSchemaDef := schemaDefToAPI(ctx, plan.SchemaDef, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Create: schema_def converted to: %+v", genericSchemaDef))
		if genericSchemaDef == nil || len(genericSchemaDef) == 0 {
			resp.Diagnostics.AddError("SchemaDef Conversion Error", "schema_def was provided but conversion resulted in nil or empty map")
			return
		}
		typedSchemaDef := schemaDefToTypedAPI(genericSchemaDef, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		apiPayload.SetSchemaDef(typedSchemaDef)
	} else if outputType == "text" {
		if !plan.SchemaDef.IsNull() && !plan.SchemaDef.IsUnknown() {
			resp.Diagnostics.AddError("Validation Error", "schema_def must not be set when output_type is 'text'")
			return
		}
	} else {
		resp.Diagnostics.AddError("Validation Error", fmt.Sprintf("unsupported output_type '%s', must be either 'text' or 'schema'", outputType))
		return
	}

	// Common config mapping
	apiConfig := capabilityConfigModelToAPI(ctx, plan.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if apiConfig != nil {
		apiPayload.SetConfig(*apiConfig)
	}

	createdAPICap, err := r.client.CreateCompletionCapability(ctx, *apiPayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create completion capability, got error: %s", err))
		return
	}

	mapCompletionCapabilityCreateResponseToModel(createdAPICap, &plan, &resp.Diagnostics, ctx)
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

	mapCompletionCapabilityRepresentationToModel(apiCap, &state, &resp.Diagnostics, ctx)
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
	outputTypeValue := plan.OutputType.ValueString()

	updatePayload := api.NewCompletionCapabilityUpdate(plan.Name.ValueString(), "completion")

	// IsPublic
	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		updatePayload.SetIsPublic(plan.IsPublic.ValueBool())
	} else {
		updatePayload.SetIsPublic(false) // default
	}

	// SemanticID
	if !plan.SemanticID.IsNull() && !plan.SemanticID.IsUnknown() {
		updatePayload.SetSemanticId(plan.SemanticID.ValueString())
	}

	// ModelID
	if !plan.ModelID.IsNull() && !plan.ModelID.IsUnknown() {
		updatePayload.SetModelId(plan.ModelID.ValueString())
	}

	// ProjectID
	if !plan.ProjectID.IsNull() && !plan.ProjectID.IsUnknown() {
		updatePayload.SetProjectId(plan.ProjectID.ValueString())
	}

	// SystemPrompt, CompletionPrompt, OutputType
	updatePayload.SetSystemPrompt(plan.SystemPrompt.ValueString())
	updatePayload.SetCompletionPrompt(plan.CompletionPrompt.ValueString())
	updatePayload.SetOutputType(outputTypeValue)

	// Variables
	if !plan.Variables.IsNull() && !plan.Variables.IsUnknown() {
		var vars []string
		resp.Diagnostics.Append(plan.Variables.ElementsAs(ctx, &vars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updatePayload.SetVariables(vars)
	}

	// SchemaDef
	if outputTypeValue == "schema" {
		if plan.SchemaDef.IsNull() || plan.SchemaDef.IsUnknown() {
			resp.Diagnostics.AddError("Validation Error", "schema_def is required when output_type is 'schema'")
			return
		}
		genericSchemaDef := schemaDefToAPI(ctx, plan.SchemaDef, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		typedSchemaDef := schemaDefToTypedAPI(genericSchemaDef, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		updatePayload.SetSchemaDef(typedSchemaDef)
	} else if outputTypeValue == "text" {
		if !plan.SchemaDef.IsNull() && !plan.SchemaDef.IsUnknown() {
			resp.Diagnostics.AddError("Validation Error", "schema_def must not be set when output_type is 'text'")
			return
		}
	} else {
		resp.Diagnostics.AddError("Validation Error", fmt.Sprintf("unsupported output_type '%s', must be either 'text' or 'schema'", outputTypeValue))
		return
	}

	// Config
	apiConfig := capabilityConfigModelToAPI(ctx, plan.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if apiConfig != nil {
		updatePayload.SetConfig(*apiConfig)
	}
	// --- End of payload construction ---

	updatedAPICap, err := r.client.UpdateCompletionCapability(ctx, capabilityID, *updatePayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update completion capability %s: %s", capabilityID, err))
		return
	}

	mapCompletionCapabilityRepresentationToModel(updatedAPICap, &plan, &resp.Diagnostics, ctx)
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

// --- State upgrade: schema_def dynamic (pre-1.0) -> JSON string (1.x) ---
//
// Releases before 1.0 (the bjarkehs/corax provider and the 0.x line) modeled
// schema_def as a dynamic attribute, so existing state stores it as a structured
// JSON value (typically an object). The 1.x schema models schema_def as a
// JSON-encoded string. Without an upgrader, reading that prior state fails with
// "unsupported type json.Delim sent as tftypes.String", which blocks every plan
// and apply against pre-existing state.
//
// SchemaVersion is therefore 1 (see Schema), and this upgrader rewrites a
// version-0 instance by carrying the prior structured schema_def across as its
// JSON text. The value is corrected on the next Read regardless, so the only goal
// here is to produce a state that decodes cleanly under the 1.x schema.

// completionCapabilityResourceModelV0 mirrors the current model except schema_def,
// which was a dynamic value in version-0 state.
type completionCapabilityResourceModelV0 struct {
	ID               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	SemanticID       types.String  `tfsdk:"semantic_id"`
	IsPublic         types.Bool    `tfsdk:"is_public"`
	ModelID          types.String  `tfsdk:"model_id"`
	Config           types.Object  `tfsdk:"config"`
	ProjectID        types.String  `tfsdk:"project_id"`
	SystemPrompt     types.String  `tfsdk:"system_prompt"`
	CompletionPrompt types.String  `tfsdk:"completion_prompt"`
	Variables        types.Set     `tfsdk:"variables"`
	OutputType       types.String  `tfsdk:"output_type"`
	SchemaDef        types.Dynamic `tfsdk:"schema_def"`
	Owner            types.String  `tfsdk:"owner"`
	Type             types.String  `tfsdk:"type"`
	CreatedAt        types.String  `tfsdk:"created_at"`
	UpdatedAt        types.String  `tfsdk:"updated_at"`
	CreatedBy        types.String  `tfsdk:"created_by"`
	UpdatedBy        types.String  `tfsdk:"updated_by"`
}

func (r *CompletionCapabilityResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// The prior schema is the current schema with schema_def restored to a dynamic
	// attribute, so version-0 state (where schema_def is a structured value) decodes
	// cleanly. Reusing the current attributes keeps the two schemas in lock-step.
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	priorAttributes := make(map[string]schema.Attribute, len(schemaResp.Schema.Attributes))
	for name, attribute := range schemaResp.Schema.Attributes {
		priorAttributes[name] = attribute
	}
	priorAttributes["schema_def"] = schema.DynamicAttribute{
		Optional:            true,
		MarkdownDescription: "Pre-1.0 dynamic representation of the output schema.",
	}

	priorSchema := schema.Schema{
		Version:    0,
		Attributes: priorAttributes,
	}

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &priorSchema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior completionCapabilityResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				var rawJSON []byte
				if req.RawState != nil {
					rawJSON = req.RawState.JSON
				}

				upgraded := CompletionCapabilityResourceModel{
					ID:               prior.ID,
					Name:             prior.Name,
					SemanticID:       prior.SemanticID,
					IsPublic:         prior.IsPublic,
					ModelID:          prior.ModelID,
					Config:           prior.Config,
					ProjectID:        prior.ProjectID,
					SystemPrompt:     prior.SystemPrompt,
					CompletionPrompt: prior.CompletionPrompt,
					Variables:        prior.Variables,
					OutputType:       prior.OutputType,
					SchemaDef:        upgradeSchemaDefFromRawState(rawJSON, &resp.Diagnostics),
					Owner:            prior.Owner,
					Type:             prior.Type,
					CreatedAt:        prior.CreatedAt,
					UpdatedAt:        prior.UpdatedAt,
					CreatedBy:        prior.CreatedBy,
					UpdatedBy:        prior.UpdatedBy,
				}
				if resp.Diagnostics.HasError() {
					return
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgraded)...)
			},
		},
	}
}

// upgradeSchemaDefFromRawState reads schema_def out of the raw prior-state JSON and
// returns it as the JSON-string value the 1.x schema expects. A structured value
// (object/array) is carried across verbatim as its JSON text; a value already stored
// as a JSON string is decoded so the result isn't double-encoded; null/absent maps to
// null. The value is authoritative only until the next Read refreshes it from the API.
func upgradeSchemaDefFromRawState(rawJSON []byte, diags *diag.Diagnostics) types.String {
	if len(rawJSON) == 0 {
		return types.StringNull()
	}

	var attributes map[string]json.RawMessage
	if err := json.Unmarshal(rawJSON, &attributes); err != nil {
		diags.AddError(
			"Unable to upgrade schema_def",
			fmt.Sprintf("Could not parse prior state JSON while upgrading schema_def: %s", err),
		)
		return types.StringNull()
	}

	raw, ok := attributes["schema_def"]
	if !ok {
		return types.StringNull()
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return types.StringNull()
	}

	// A leading quote means the prior value was already a JSON string; decode it so the
	// result is the string's content rather than a double-encoded string.
	if trimmed[0] == '"' {
		var decoded string
		if err := json.Unmarshal(trimmed, &decoded); err != nil {
			diags.AddError(
				"Unable to upgrade schema_def",
				fmt.Sprintf("Could not decode prior schema_def string value: %s", err),
			)
			return types.StringNull()
		}
		return types.StringValue(decoded)
	}

	// Structured value (object/array): carry it across as its JSON text.
	return types.StringValue(string(trimmed))
}
