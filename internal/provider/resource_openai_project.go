package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

type ProjectResource struct {
	client *OpenAIClient
}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

type ProjectResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Status     types.String `tfsdk:"status"`
	CreatedAt  types.String `tfsdk:"created_at"`
	ArchivedAt types.String `tfsdk:"archived_at"`
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI Project.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the project.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the project (e.g. active, archived).",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the project was created.",
			},
			"archived_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the project was archived.",
			},
		},
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerClient, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *provider.OpenAIClient, got: %T", req.ProviderData))
		return
	}

	r.client = providerClient
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := map[string]interface{}{
		"name": data.Name.ValueString(),
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	httpResp, err := doRequestWithRetry(ctx, projectClientHTTP(r.client), r.client, http.MethodPost, adminBaseURL(r.client)+"/v1/organization/projects", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}
	defer httpResp.Body.Close()

	project, err := decodeProjectResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError("API error creating project", err.Error())
		return
	}
	setProjectResourceData(&data, project)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := doRequestWithRetry(ctx, projectClientHTTP(r.client), r.client, http.MethodGet, adminBaseURL(r.client)+"/v1/organization/projects/"+data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	project, err := decodeProjectResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError("API error reading project", err.Error())
		return
	}
	setProjectResourceData(&data, project)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := map[string]interface{}{
		"name": data.Name.ValueString(),
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	httpResp, err := doRequestWithRetry(ctx, projectClientHTTP(r.client), r.client, http.MethodPost, adminBaseURL(r.client)+"/v1/organization/projects/"+data.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}
	defer httpResp.Body.Close()

	project, err := decodeProjectResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError("API error updating project", err.Error())
		return
	}
	setProjectResourceData(&data, project)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := doRequestWithRetry(ctx, projectClientHTTP(r.client), r.client, http.MethodPost, adminBaseURL(r.client)+"/v1/organization/projects/"+data.ID.ValueString()+"/archive", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting (archiving) project", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		return
	}
	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API error deleting (archiving) project", fmt.Sprintf("%s - %s", httpResp.Status, string(body)))
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func decodeProjectResponse(resp *http.Response) (*ProjectResponseFramework, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s - %s", resp.Status, string(body))
	}

	var project ProjectResponseFramework
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project response: %w", err)
	}
	return &project, nil
}

func setProjectResourceData(data *ProjectResourceModel, project *ProjectResponseFramework) {
	data.ID = types.StringValue(project.ID)
	data.Name = types.StringValue(project.Name)
	data.Status = types.StringValue(project.Status)

	if project.CreatedAt != 0 {
		data.CreatedAt = types.StringValue(time.Unix(project.CreatedAt, 0).Format(time.RFC3339))
	}

	if project.ArchivedAt != nil {
		data.ArchivedAt = types.StringValue(time.Unix(*project.ArchivedAt, 0).Format(time.RFC3339))
	} else {
		data.ArchivedAt = types.StringNull()
	}
}
