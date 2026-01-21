package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CheckpointPermissionResource{}
var _ resource.ResourceWithImportState = &CheckpointPermissionResource{}

type CheckpointPermissionResource struct {
	client *OpenAIClient
}

func NewCheckpointPermissionResource() resource.Resource {
	return &CheckpointPermissionResource{}
}

func (r *CheckpointPermissionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fine_tuning_checkpoint_permission"
}

type CheckpointPermissionResourceModel struct {
	ID           types.String `tfsdk:"id"`
	CheckpointID types.String `tfsdk:"checkpoint_id"`
	ProjectIDs   types.List   `tfsdk:"project_ids"`
	CreatedAt    types.Int64  `tfsdk:"created_at"`
}

func (r *CheckpointPermissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages permissions for a fine-tuning checkpoint (utility resource).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"checkpoint_id": schema.StringAttribute{
				Description: "The ID of the checkpoint to set permissions for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_ids": schema.ListAttribute{
				Description:   "The list of project IDs to grant access to.",
				Required:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.List{
					// Could be updateable? SDKv2 schema says ForceNew: true?
					// Step 966 line 43 says ForceNew: true.
				},
				// Actually list modifiers need generic or custom.
				// We'll enforce replacement via manual check or Update error if needed.
				// But since it's ForceNew in SDKv2, we should probably support Update or ForceNew?
				// If we want ForceNew behavior in Framework, we use RequiresReplace modifier on the attribute.
			},
			"created_at": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *CheckpointPermissionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *provider.OpenAIClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *CheckpointPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CheckpointPermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// This resource seems to be a utility wrapper.
	// It calls an endpoint to ADD permissions.
	// Does it REPLACES permissions?
	// Checkpoint permissions are typically additive?

	// SDKv2 implemented this.
	// request struct: ProjectIDs []string `json:"project_ids"`
	// It POSTs to... where?
	// I need to know the endpoint.
	// Step 966 didn't show the `resourceOpenAIFineTuningCheckpointPermissionCreate` function body API call.
	// But it likely calls `client.AddCheckpointPermissions`?
	// or `POST /fine_tuning/checkpoints/{id}`?

	// I'll bet it calls `POST /fine_tuning/jobs/{job_id}/checkpoints`? No.
	// Maybe `POST /fine_tuning/checkpoints/{checkpoint_id}`?

	// I will attempt to use a likely path or assume the client method `AddCheckpointPermission` exists.
	// But I don't see client methods.
	// The resource name `fine_tuning_checkpoint_permission` implies it might be using an undocumented or specific API.

	// "This resource requires an admin API key with api.fine_tuning.checkpoints.write scope"

	// I'll search for this endpoint docs.
	// `POST https://api.openai.com/v1/fine_tuning/checkpoints/{checkpoint_id}` ?

	// Actually, let's look at `resourceOpenAIFineTuningCheckpointPermissionCreate` if I can see more lines.
	// I'll do a quick view of lines 100+ of `resource_openai_fine_tuning_checkpoint_permission.go`.

	resp.Diagnostics.AddError("Implementation Pending", "Need to verify API endpoint")
}

func (r *CheckpointPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Dummy read or state only?
}

func (r *CheckpointPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *CheckpointPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *CheckpointPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
