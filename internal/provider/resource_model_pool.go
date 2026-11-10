// Copyright (c) Trifork

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-corax/internal/coraxclient"
	api "terraform-provider-corax/internal/generated"
)

var _ resource.Resource = &ModelPoolResource{}
var _ resource.ResourceWithImportState = &ModelPoolResource{}

func NewModelPoolResource() resource.Resource {
	return &ModelPoolResource{}
}

// ModelPoolResource manages a model deployment pool and its members.
type ModelPoolResource struct {
	client *coraxclient.Client
}

type ModelPoolResourceModel struct {
	ID              types.String                    `tfsdk:"id"`
	Name            types.String                    `tfsdk:"name"`
	Strategy        types.String                    `tfsdk:"strategy"`
	CooldownSeconds types.Int64                     `tfsdk:"cooldown_seconds"`
	FallbackMode    types.String                    `tfsdk:"fallback_mode"`
	Members         map[string]ModelPoolMemberModel `tfsdk:"members"`
}

// ModelPoolMemberModel is one member of a pool. Members are keyed by model
// deployment ID; the API's member ID is looked up when reconciling. A map (not
// a set) so attribute defaults resolve per key.
type ModelPoolMemberModel struct {
	Weight   types.Int64 `tfsdk:"weight"`
	Priority types.Int64 `tfsdk:"priority"`
	IsActive types.Bool  `tfsdk:"is_active"`
}

func enumStrings[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return out
}

func (r *ModelPoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_pool"
}

func (r *ModelPoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Corax Model Pool. A pool load balances requests across several model deployments; capabilities use it through `model_pool_id`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for the model pool (UUID).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A user-defined name for the model pool.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 100)},
			},
			"strategy": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(api.WEIGHTED_RANDOM)),
				MarkdownDescription: "How requests are distributed across members. Defaults to `weighted_random`.",
				Validators:          []validator.String{stringvalidator.OneOf(enumStrings(api.AllowedPoolStrategyEnumValues)...)},
			},
			"cooldown_seconds": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				MarkdownDescription: "How long a failed member is skipped before it is tried again. Defaults to 60.",
				Validators:          []validator.Int64{int64validator.AtLeast(0)},
			},
			"fallback_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(api.TRY_NEXT)),
				MarkdownDescription: "What happens when the selected member fails: `try_next` falls back to another member, `fail_fast` returns the error. Defaults to `try_next`.",
				Validators:          []validator.String{stringvalidator.OneOf(enumStrings(api.AllowedPoolFallbackModeEnumValues)...)},
			},
			"members": schema.MapNestedAttribute{
				Required:            true,
				MarkdownDescription: "The model deployments in the pool, keyed by model deployment UUID. Use `{}` to take the member defaults.",
				Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"weight": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(1),
							MarkdownDescription: "Relative share of traffic for this member. Defaults to 1.",
							Validators:          []validator.Int64{int64validator.AtLeast(1)},
						},
						"priority": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(1),
							MarkdownDescription: "Member priority. Defaults to 1.",
							Validators:          []validator.Int64{int64validator.AtLeast(1)},
						},
						"is_active": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
							MarkdownDescription: "Whether the member receives traffic. Defaults to true.",
						},
					},
				},
			},
		},
	}
}

func (r *ModelPoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func mapModelPoolToModel(pool *api.ModelDeploymentPool, model *ModelPoolResourceModel) {
	model.ID = types.StringValue(pool.Id)
	model.Name = types.StringValue(pool.Name)
	model.Strategy = types.StringValue(string(pool.GetStrategy()))
	model.CooldownSeconds = types.Int64Value(int64(pool.GetCooldownSeconds()))
	model.FallbackMode = types.StringValue(string(pool.GetFallbackMode()))

	members := make(map[string]ModelPoolMemberModel, len(pool.Members))
	for _, m := range pool.Members {
		members[m.ModelDeploymentId] = ModelPoolMemberModel{
			Weight:   types.Int64Value(int64(m.GetWeight())),
			Priority: types.Int64Value(int64(m.GetPriority())),
			IsActive: types.BoolValue(m.GetIsActive()),
		}
	}
	model.Members = members
}

func (r *ModelPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModelPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := api.NewModelDeploymentPoolCreate(plan.Name.ValueString())
	payload.SetStrategy(api.PoolStrategy(plan.Strategy.ValueString()))
	payload.SetCooldownSeconds(int32(plan.CooldownSeconds.ValueInt64()))
	payload.SetFallbackMode(api.PoolFallbackMode(plan.FallbackMode.ValueString()))

	members := make([]api.PoolMemberCreate, 0, len(plan.Members))
	for deploymentID, m := range plan.Members {
		member := api.NewPoolMemberCreate(deploymentID)
		member.SetWeight(int32(m.Weight.ValueInt64()))
		member.SetPriority(int32(m.Priority.ValueInt64()))
		member.SetIsActive(m.IsActive.ValueBool())
		members = append(members, *member)
	}
	payload.SetMembers(members)

	pool, err := r.client.CreateModelPool(ctx, *payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create model pool, got error: %s", err))
		return
	}

	mapModelPoolToModel(pool, &plan)
	tflog.Info(ctx, fmt.Sprintf("Model Pool %s created with ID %s", plan.Name.ValueString(), plan.ID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ModelPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModelPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := r.client.GetModelPool(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, coraxclient.ErrNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Model Pool %s not found, removing from state", state.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read model pool %s: %s", state.ID.ValueString(), err))
		return
	}

	mapModelPoolToModel(pool, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update sets the pool-level fields, then reconciles members against what the
// API currently holds. Removals run last so the pool is never left empty
// mid-update.
func (r *ModelPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ModelPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	poolID := state.ID.ValueString()

	update := api.NewModelDeploymentPoolUpdate()
	update.SetName(plan.Name.ValueString())
	update.SetStrategy(api.PoolStrategy(plan.Strategy.ValueString()))
	update.SetCooldownSeconds(int32(plan.CooldownSeconds.ValueInt64()))
	update.SetFallbackMode(api.PoolFallbackMode(plan.FallbackMode.ValueString()))

	pool, err := r.client.UpdateModelPool(ctx, poolID, *update)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update model pool %s: %s", poolID, err))
		return
	}

	current := make(map[string]api.PoolMember, len(pool.Members))
	for _, m := range pool.Members {
		current[m.ModelDeploymentId] = m
	}

	for deploymentID, m := range plan.Members {
		existing, ok := current[deploymentID]
		delete(current, deploymentID)

		if !ok {
			create := api.NewPoolMemberCreate(deploymentID)
			create.SetWeight(int32(m.Weight.ValueInt64()))
			create.SetPriority(int32(m.Priority.ValueInt64()))
			create.SetIsActive(m.IsActive.ValueBool())
			if _, err := r.client.AddPoolMember(ctx, poolID, *create); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add model deployment %s to pool %s: %s", deploymentID, poolID, err))
				return
			}
			continue
		}

		if int64(existing.GetWeight()) == m.Weight.ValueInt64() &&
			int64(existing.GetPriority()) == m.Priority.ValueInt64() &&
			existing.GetIsActive() == m.IsActive.ValueBool() {
			continue
		}
		memberUpdate := api.NewPoolMemberUpdate()
		memberUpdate.SetWeight(int32(m.Weight.ValueInt64()))
		memberUpdate.SetPriority(int32(m.Priority.ValueInt64()))
		memberUpdate.SetIsActive(m.IsActive.ValueBool())
		if _, err := r.client.UpdatePoolMember(ctx, poolID, existing.Id, *memberUpdate); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update model deployment %s in pool %s: %s", deploymentID, poolID, err))
			return
		}
	}

	// Whatever is left in current is no longer in the plan.
	for deploymentID, stale := range current {
		if err := r.client.RemovePoolMember(ctx, poolID, stale.Id); err != nil && !errors.Is(err, coraxclient.ErrNotFound) {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove model deployment %s from pool %s: %s", deploymentID, poolID, err))
			return
		}
	}

	pool, err = r.client.GetModelPool(ctx, poolID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read model pool %s after update: %s", poolID, err))
		return
	}

	mapModelPoolToModel(pool, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ModelPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ModelPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteModelPool(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, coraxclient.ErrNotFound) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete model pool %s: %s", state.ID.ValueString(), err))
	}
}

func (r *ModelPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
