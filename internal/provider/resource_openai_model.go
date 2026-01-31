package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ModelResource{}
var _ resource.ResourceWithImportState = &ModelResource{}

type ModelResource struct {
	client *OpenAIClient
}

func NewModelResource() resource.Resource {
	return &ModelResource{}
}

func (r *ModelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

type ModelResourceModel struct {
	Model      types.String `tfsdk:"model"`
	OwnedBy    types.String `tfsdk:"owned_by"`
	CreatedAt  types.Int64  `tfsdk:"created_at"`
	Object     types.String `tfsdk:"object"`
	Permission types.List   `tfsdk:"permission"`
}

func (r *ModelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The model resource allows you to pull information about a specific model.",
		Attributes: map[string]schema.Attribute{
			"model": schema.StringAttribute{
				Description: "The ID of the model (e.g., gpt-4, gpt-3.5-turbo, etc.)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"owned_by": schema.StringAttribute{
				Description: "The organization that owns the model",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The timestamp when the model was created",
				Computed:    true,
			},
			"object": schema.StringAttribute{
				Description: "The object type (always 'model')",
				Computed:    true,
			},
			"permission": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                   types.StringType,
						"object":               types.StringType,
						"created":              types.Int64Type,
						"allow_create_engine":  types.BoolType,
						"allow_sampling":       types.BoolType,
						"allow_logprobs":       types.BoolType,
						"allow_search_indices": types.BoolType,
						"allow_view":           types.BoolType,
						"allow_fine_tuning":    types.BoolType,
						"organization":         types.StringType,
						"group":                types.StringType, // Group is interface{} in SDKv2 but usually string or null
						"is_blocking":          types.BoolType,
					},
				},
			},
		},
	}
}

func (r *ModelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Create for openai_model is just setting the ID and reading, as we helpfully
	// adopt an existing model ID into state.
	var data ModelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Just write to state for now, Read will fetch details
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ModelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	modelID := data.Model.ValueString()
	if modelID == "" {
		resp.Diagnostics.AddError("Model ID Missing", "Model ID is required")
		return
	}

	path := fmt.Sprintf("models/%s", modelID)

	respBody, err := r.client.DoRequest("GET", path, nil)
	if err != nil {
		// Handle 404
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading model", err.Error())
		return
	}

	var model ModelInfoResponse
	if err := json.Unmarshal(respBody, &model); err != nil {
		resp.Diagnostics.AddError("Error parsing model response", err.Error())
		return
	}

	data.OwnedBy = types.StringValue(model.OwnedBy)
	data.CreatedAt = types.Int64Value(int64(model.Created))
	data.Object = types.StringValue(model.Object)

	// Permissions
	if len(model.Permission) > 0 {
		perms := []attr.Value{}
		for _, p := range model.Permission {
			grpStr := types.StringNull()
			if p.Group != nil {
				if s, ok := p.Group.(string); ok {
					grpStr = types.StringValue(s)
				}
			}

			obj, _ := types.ObjectValue(
				map[string]attr.Type{
					"id":                   types.StringType,
					"object":               types.StringType,
					"created":              types.Int64Type,
					"allow_create_engine":  types.BoolType,
					"allow_sampling":       types.BoolType,
					"allow_logprobs":       types.BoolType,
					"allow_search_indices": types.BoolType,
					"allow_view":           types.BoolType,
					"allow_fine_tuning":    types.BoolType,
					"organization":         types.StringType,
					"group":                types.StringType,
					"is_blocking":          types.BoolType,
				},
				map[string]attr.Value{
					"id":                   types.StringValue(p.ID),
					"object":               types.StringValue(p.Object),
					"created":              types.Int64Value(int64(p.Created)),
					"allow_create_engine":  types.BoolValue(p.AllowCreateEngine),
					"allow_sampling":       types.BoolValue(p.AllowSampling),
					"allow_logprobs":       types.BoolValue(p.AllowLogprobs),
					"allow_search_indices": types.BoolValue(p.AllowSearchIndices),
					"allow_view":           types.BoolValue(p.AllowView),
					"allow_fine_tuning":    types.BoolValue(p.AllowFineTuning),
					"organization":         types.StringValue(p.Organization),
					"group":                grpStr,
					"is_blocking":          types.BoolValue(p.IsBlocking),
				},
			)
			perms = append(perms, obj)
		}
		permList, _ := types.ListValue(types.ObjectType{AttrTypes: map[string]attr.Type{
			"id":                   types.StringType,
			"object":               types.StringType,
			"created":              types.Int64Type,
			"allow_create_engine":  types.BoolType,
			"allow_sampling":       types.BoolType,
			"allow_logprobs":       types.BoolType,
			"allow_search_indices": types.BoolType,
			"allow_view":           types.BoolType,
			"allow_fine_tuning":    types.BoolType,
			"organization":         types.StringType,
			"group":                types.StringType,
			"is_blocking":          types.BoolType,
		}}, perms)
		data.Permission = permList
	} else {
		data.Permission = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
			"id":                   types.StringType,
			"object":               types.StringType,
			"created":              types.Int64Type,
			"allow_create_engine":  types.BoolType,
			"allow_sampling":       types.BoolType,
			"allow_logprobs":       types.BoolType,
			"allow_search_indices": types.BoolType,
			"allow_view":           types.BoolType,
			"allow_fine_tuning":    types.BoolType,
			"organization":         types.StringType,
			"group":                types.StringType,
			"is_blocking":          types.BoolType,
		}})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Models are read-only / immutable in this context
	resp.Diagnostics.AddError("Operation not supported", "Update is not supported for openai_model")
}

func (r *ModelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Removes from state only. Cannot delete model via API using this resource (unless maybe if fine-tuned, but there's a separate resource for that).
	// Legacy implementation was a no-op on delete.
}

func (r *ModelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("model"), req, resp)
}
