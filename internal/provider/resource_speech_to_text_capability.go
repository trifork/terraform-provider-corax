// Copyright (c) Trifork

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-corax/internal/coraxclient"
	api "terraform-provider-corax/internal/generated"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SpeechToTextCapabilityResource{}
var _ resource.ResourceWithImportState = &SpeechToTextCapabilityResource{}

func NewSpeechToTextCapabilityResource() resource.Resource {
	return &SpeechToTextCapabilityResource{}
}

// SpeechToTextCapabilityResource defines the resource implementation.
type SpeechToTextCapabilityResource struct {
	client *coraxclient.Client
}

// SpeechToTextCapabilityResourceModel describes the resource data model.
type SpeechToTextCapabilityResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	IsPublic     types.Bool   `tfsdk:"is_public"`
	ModelID      types.String `tfsdk:"model_id"`      // Nullable
	ModelPoolID  types.String `tfsdk:"model_pool_id"` // Nullable
	Config       types.Object `tfsdk:"config"`        // Nullable
	ProjectID    types.String `tfsdk:"project_id"`    // Nullable
	SemanticID   types.String `tfsdk:"semantic_id"`   // Nullable
	SystemPrompt types.String `tfsdk:"system_prompt"` // Nullable (optional for STT)
	OutputType   types.String `tfsdk:"output_type"`   // Default "text"
	Owner        types.String `tfsdk:"owner"`         // Computed
	Type         types.String `tfsdk:"type"`          // Computed, should always be "speech_to_text"
	CreatedAt    types.String `tfsdk:"created_at"`    // Computed
	UpdatedAt    types.String `tfsdk:"updated_at"`    // Computed
	CreatedBy    types.String `tfsdk:"created_by"`    // Computed
	UpdatedBy    types.String `tfsdk:"updated_by"`    // Computed
}

func (r *SpeechToTextCapabilityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_speech_to_text_capability"
}

func (r *SpeechToTextCapabilityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Corax Speech-to-Text Capability. Speech-to-text capabilities define configurations for audio transcription models.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for the speech-to-text capability (UUID).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A user-defined name for the speech-to-text capability.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"is_public": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Indicates whether the capability is publicly accessible. Defaults to false.",
			},
			"model_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The UUID of the model deployment to use for this capability. Mutually exclusive with model_pool_id.",
			},
			"model_pool_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The UUID of the model pool to use for this capability. Mutually exclusive with model_id.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The UUID of the project this capability belongs to.",
			},
			"semantic_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A human-readable semantic identifier (lowercase alphanumeric with hyphens, e.g. 'my-stt-capability').",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`),
						"must be lowercase alphanumeric with hyphens",
					),
				},
			},
			"system_prompt": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "An optional system prompt to guide the transcription behavior.",
			},
			"output_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("text"),
				MarkdownDescription: "The output format type. Defaults to 'text'.",
			},
			"config": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Configuration settings for the capability's behavior.",
				Attributes:          capabilityConfigSchemaAttributes(),
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			},
			"owner":      schema.StringAttribute{Computed: true, MarkdownDescription: "Owner of the capability.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"type":       schema.StringAttribute{Computed: true, MarkdownDescription: "Type of the capability (should be 'speech_to_text').", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"created_at": schema.StringAttribute{Computed: true, MarkdownDescription: "The date and time the capability was created (RFC3339 format).", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"updated_at": schema.StringAttribute{Computed: true, MarkdownDescription: "The date and time the capability was last updated (RFC3339 format).", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"created_by": schema.StringAttribute{Computed: true, MarkdownDescription: "The identifier of who created the capability.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"updated_by": schema.StringAttribute{Computed: true, MarkdownDescription: "The identifier of who last updated the capability.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func (r *SpeechToTextCapabilityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// mapSpeechToTextCapabilityRepresentationToModel maps an api.CapabilityRepresentation (from Get/Update) to the TF model.
func mapSpeechToTextCapabilityRepresentationToModel(apiCap *api.CapabilityRepresentation, model *SpeechToTextCapabilityResourceModel, diags *diag.Diagnostics, ctx context.Context) {
	model.ID = types.StringValue(apiCap.Id)
	model.Name = types.StringValue(apiCap.Name)
	model.IsPublic = types.BoolValue(apiCap.GetIsPublic())
	model.Type = types.StringValue(apiCap.Type)

	if modelId, ok := apiCap.GetModelIdOk(); ok && modelId != nil {
		model.ModelID = types.StringValue(*modelId)
	} else {
		model.ModelID = types.StringNull()
	}
	if modelPoolId, ok := apiCap.GetModelPoolIdOk(); ok && modelPoolId != nil {
		model.ModelPoolID = types.StringValue(*modelPoolId)
	} else {
		model.ModelPoolID = types.StringNull()
	}
	if projectId, ok := apiCap.GetProjectIdOk(); ok && projectId != nil {
		model.ProjectID = types.StringValue(*projectId)
	} else {
		model.ProjectID = types.StringNull()
	}

	// SemanticID from CapabilityRepresentation
	if semanticId, ok := apiCap.GetSemanticIdOk(); ok && semanticId != nil {
		model.SemanticID = types.StringValue(*semanticId)
	} else {
		model.SemanticID = types.StringNull()
	}

	// SystemPrompt is in apiCap.Configuration map for CapabilityRepresentation
	if sysPrompt, ok := apiCap.Configuration["system_prompt"].(string); ok && sysPrompt != "" {
		model.SystemPrompt = types.StringValue(sysPrompt)
	} else {
		model.SystemPrompt = types.StringNull()
	}

	// OutputType from Configuration map
	if outputType, ok := apiCap.Configuration["output_type"].(string); ok {
		model.OutputType = types.StringValue(outputType)
	} else {
		model.OutputType = types.StringValue("text") // default
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

// mapSpeechToTextCapabilityCreateResponseToModel maps an api.SpeechToTextCapability (from Create) to the TF model.
func mapSpeechToTextCapabilityCreateResponseToModel(apiCap *api.SpeechToTextCapability, model *SpeechToTextCapabilityResourceModel, diags *diag.Diagnostics, ctx context.Context) {
	model.ID = types.StringValue(apiCap.Id)
	model.Name = types.StringValue(apiCap.Name)
	model.IsPublic = types.BoolValue(apiCap.GetIsPublic())

	if apiCap.Type != nil {
		model.Type = types.StringValue(*apiCap.Type)
	} else {
		model.Type = types.StringValue("speech_to_text")
	}

	if modelId, ok := apiCap.GetModelIdOk(); ok && modelId != nil {
		model.ModelID = types.StringValue(*modelId)
	} else {
		model.ModelID = types.StringNull()
	}
	if modelPoolId, ok := apiCap.GetModelPoolIdOk(); ok && modelPoolId != nil {
		model.ModelPoolID = types.StringValue(*modelPoolId)
	} else {
		model.ModelPoolID = types.StringNull()
	}
	if projectId, ok := apiCap.GetProjectIdOk(); ok && projectId != nil {
		model.ProjectID = types.StringValue(*projectId)
	} else {
		model.ProjectID = types.StringNull()
	}
	if semanticId, ok := apiCap.GetSemanticIdOk(); ok && semanticId != nil {
		model.SemanticID = types.StringValue(*semanticId)
	} else {
		model.SemanticID = types.StringNull()
	}

	if sysPrompt, ok := apiCap.GetSystemPromptOk(); ok && sysPrompt != nil && *sysPrompt != "" {
		model.SystemPrompt = types.StringValue(*sysPrompt)
	} else {
		model.SystemPrompt = types.StringNull()
	}

	if outputType, ok := apiCap.GetOutputTypeOk(); ok && outputType != nil {
		model.OutputType = types.StringValue(*outputType)
	} else {
		model.OutputType = types.StringValue("text")
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

func (r *SpeechToTextCapabilityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SpeechToTextCapabilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating Speech-to-Text Capability: %s", plan.Name.ValueString()))

	apiPayload := api.NewSpeechToTextCapabilityCreate(plan.Name.ValueString(), "speech_to_text")

	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		apiPayload.SetIsPublic(plan.IsPublic.ValueBool())
	}
	if !plan.ModelID.IsNull() && !plan.ModelID.IsUnknown() {
		apiPayload.SetModelId(plan.ModelID.ValueString())
	}
	if !plan.ModelPoolID.IsNull() && !plan.ModelPoolID.IsUnknown() {
		apiPayload.SetModelPoolId(plan.ModelPoolID.ValueString())
	}
	if !plan.ProjectID.IsNull() && !plan.ProjectID.IsUnknown() {
		apiPayload.SetProjectId(plan.ProjectID.ValueString())
	}
	if !plan.SemanticID.IsNull() && !plan.SemanticID.IsUnknown() {
		apiPayload.SetSemanticId(plan.SemanticID.ValueString())
	}
	if !plan.SystemPrompt.IsNull() && !plan.SystemPrompt.IsUnknown() {
		apiPayload.SetSystemPrompt(plan.SystemPrompt.ValueString())
	}
	if !plan.OutputType.IsNull() && !plan.OutputType.IsUnknown() {
		apiPayload.SetOutputType(plan.OutputType.ValueString())
	}

	apiConfig := capabilityConfigModelToAPI(ctx, plan.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if apiConfig != nil {
		apiPayload.SetConfig(*apiConfig)
	}

	createdAPICap, err := r.client.CreateSpeechToTextCapability(ctx, *apiPayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create speech-to-text capability, got error: %s", err))
		return
	}

	mapSpeechToTextCapabilityCreateResponseToModel(createdAPICap, &plan, &resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Speech-to-Text Capability %s created successfully with ID %s", plan.Name.ValueString(), plan.ID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SpeechToTextCapabilityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SpeechToTextCapabilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	capabilityID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Reading Speech-to-Text Capability with ID: %s", capabilityID))

	apiCap, err := r.client.GetCapability(ctx, capabilityID)
	if err != nil {
		if errors.Is(err, coraxclient.ErrNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Speech-to-Text Capability %s not found, removing from state", capabilityID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read speech-to-text capability %s: %s", capabilityID, err))
		return
	}

	if apiCap.Type != "speech_to_text" {
		resp.Diagnostics.AddError("Resource Type Mismatch", fmt.Sprintf("Expected capability type 'speech_to_text' but found '%s' for ID %s. Removing from state.", apiCap.Type, capabilityID))
		resp.State.RemoveResource(ctx)
		return
	}

	mapSpeechToTextCapabilityRepresentationToModel(apiCap, &state, &resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read Speech-to-Text Capability %s", capabilityID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SpeechToTextCapabilityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SpeechToTextCapabilityResourceModel
	var state SpeechToTextCapabilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	capabilityID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Updating Speech-to-Text Capability with ID: %s", capabilityID))

	updatePayload := api.NewSpeechToTextCapabilityUpdate(plan.Name.ValueString(), "speech_to_text")

	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		updatePayload.SetIsPublic(plan.IsPublic.ValueBool())
	} else {
		updatePayload.SetIsPublic(false)
	}
	if !plan.ModelID.IsNull() && !plan.ModelID.IsUnknown() {
		updatePayload.SetModelId(plan.ModelID.ValueString())
	}
	if !plan.ModelPoolID.IsNull() && !plan.ModelPoolID.IsUnknown() {
		updatePayload.SetModelPoolId(plan.ModelPoolID.ValueString())
	}
	if !plan.ProjectID.IsNull() && !plan.ProjectID.IsUnknown() {
		updatePayload.SetProjectId(plan.ProjectID.ValueString())
	}
	if !plan.SemanticID.IsNull() && !plan.SemanticID.IsUnknown() {
		updatePayload.SetSemanticId(plan.SemanticID.ValueString())
	}
	if !plan.SystemPrompt.IsNull() && !plan.SystemPrompt.IsUnknown() {
		updatePayload.SetSystemPrompt(plan.SystemPrompt.ValueString())
	}
	if !plan.OutputType.IsNull() && !plan.OutputType.IsUnknown() {
		updatePayload.SetOutputType(plan.OutputType.ValueString())
	}

	apiConfig := capabilityConfigModelToAPI(ctx, plan.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if apiConfig != nil {
		updatePayload.SetConfig(*apiConfig)
	}

	updatedAPICap, err := r.client.UpdateSpeechToTextCapability(ctx, capabilityID, *updatePayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update speech-to-text capability %s: %s", capabilityID, err))
		return
	}

	mapSpeechToTextCapabilityRepresentationToModel(updatedAPICap, &plan, &resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve immutable computed fields from state
	plan.CreatedAt = state.CreatedAt
	plan.CreatedBy = state.CreatedBy
	plan.UpdatedAt = state.UpdatedAt
	plan.UpdatedBy = state.UpdatedBy

	tflog.Info(ctx, fmt.Sprintf("Speech-to-Text Capability %s updated successfully", capabilityID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SpeechToTextCapabilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SpeechToTextCapabilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	capabilityID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Deleting Speech-to-Text Capability with ID: %s", capabilityID))

	err := r.client.DeleteCapability(ctx, capabilityID)
	if err != nil {
		if errors.Is(err, coraxclient.ErrNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Speech-to-Text Capability %s not found, already deleted", capabilityID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete speech-to-text capability %s: %s", capabilityID, err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Speech-to-Text Capability %s deleted successfully", capabilityID))
}

func (r *SpeechToTextCapabilityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
