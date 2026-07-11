package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectModelPermissionsResource{}
var _ resource.ResourceWithImportState = &ProjectModelPermissionsResource{}

type ProjectModelPermissionsResource struct {
	client *OpenAIClient
}

type ProjectModelPermissionsResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Mode      types.String `tfsdk:"mode"`
	Models    types.Set    `tfsdk:"models"`
	Object    types.String `tfsdk:"object"`
}

func NewProjectModelPermissionsResource() resource.Resource {
	return &ProjectModelPermissionsResource{}
}

func (r *ProjectModelPermissionsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_model_permissions"
}

func (r *ProjectModelPermissionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the model allowlist or denylist policy for an OpenAI project.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "The project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project_id": schema.StringAttribute{Required: true, MarkdownDescription: "The ID of the project.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"mode": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Whether `models` is an allowlist or denylist.",
				Validators:          []validator.String{stringvalidator.OneOf("allow_list", "deny_list")},
			},
			"models": schema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "The models included in the policy. Sent to OpenAI as `model_ids`."},
			"object": schema.StringAttribute{Computed: true, MarkdownDescription: "The OpenAI object type."},
		},
	}
}

func (r *ProjectModelPermissionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func modelIDsFromSet(value types.Set) []string {
	result := make([]string, 0, len(value.Elements()))
	for _, element := range value.Elements() {
		result = append(result, element.(types.String).ValueString())
	}
	return result
}

func modelIDsToSet(values []string) types.Set {
	elements := make([]types.String, len(values))
	for i, value := range values {
		elements[i] = types.StringValue(value)
	}
	result, _ := types.SetValueFrom(context.Background(), types.StringType, elements)
	return result
}

func (r *ProjectModelPermissionsResource) write(ctx context.Context, data *ProjectModelPermissionsResourceModel) (*ProjectModelPermissionsAPI, error) {
	request := projectModelPermissionsRequest{Mode: data.Mode.ValueString(), ModelIDs: modelIDsFromSet(data.Models)}
	var result ProjectModelPermissionsAPI
	_, err := projectSettingsRequest(ctx, r.client, http.MethodPost, "/v1/organization/projects/"+data.ProjectID.ValueString()+"/model_permissions", request, &result)
	if err == nil {
		invalidateProjectModelPermissionsCache(r.client, data.ProjectID.ValueString())
	}
	return &result, err
}

func setProjectModelPermissionsState(data *ProjectModelPermissionsResourceModel, value *ProjectModelPermissionsAPI) {
	data.ID = types.StringValue(data.ProjectID.ValueString())
	data.Mode = types.StringValue(value.Mode)
	data.Models = modelIDsToSet(value.ModelIDs)
	data.Object = types.StringValue(value.Object)
}

func (r *ProjectModelPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectModelPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.write(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project model permissions", err.Error())
		return
	}
	setProjectModelPermissionsState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectModelPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectModelPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, status, err := cachedProjectModelPermissions(ctx, r.client, data.ProjectID.ValueString())
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading project model permissions", err.Error())
		return
	}
	setProjectModelPermissionsState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectModelPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectModelPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.write(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project model permissions", err.Error())
		return
	}
	setProjectModelPermissionsState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectModelPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectModelPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := projectSettingsRequest(ctx, r.client, http.MethodDelete, "/v1/organization/projects/"+data.ProjectID.ValueString()+"/model_permissions", nil, nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Error deleting project model permissions", err.Error())
		return
	}
	invalidateProjectModelPermissionsCache(r.client, data.ProjectID.ValueString())
}

func (r *ProjectModelPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), req.ID)...)
}
