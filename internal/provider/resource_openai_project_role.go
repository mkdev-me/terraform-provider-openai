package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectRoleResource{}
var _ resource.ResourceWithImportState = &ProjectRoleResource{}

type ProjectRoleResource struct {
	client *OpenAIClient
}

func NewProjectRoleResource() resource.Resource {
	return &ProjectRoleResource{}
}

func (r *ProjectRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_role"
}

type ProjectRoleResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	ProjectID      types.String   `tfsdk:"project_id"`
	Name           types.String   `tfsdk:"name"`
	Permissions    []types.String `tfsdk:"permissions"`
	Description    types.String   `tfsdk:"description"`
	ResourceType   types.String   `tfsdk:"resource_type"`
	PredefinedRole types.Bool     `tfsdk:"predefined_role"`
}

func (r *ProjectRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom project-level role in an OpenAI project.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the project role (project_id:role_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the role (mapped to `role_name` in the API).",
			},
			"permissions": schema.ListAttribute{
				Required:            true,
				MarkdownDescription: "The list of permissions granted by this role.",
				ElementType:         types.StringType,
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A description of the role.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource type the role is bound to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"predefined_role": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the role is predefined and managed by OpenAI.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProjectRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()

	permissions := make([]string, len(data.Permissions))
	for i, p := range data.Permissions {
		permissions[i] = p.ValueString()
	}

	createReq := RoleCreateRequest{
		RoleName:    data.Name.ValueString(),
		Permissions: permissions,
		Description: data.Description.ValueString(),
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	apiURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/roles"
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setAdminAuthHeaders(r.client, httpReq)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project role", err.Error())
		return
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("API error creating project role", fmt.Sprintf("%s - %s", httpResp.Status, string(respBody)))
		return
	}

	var roleResp RoleResponseFramework
	if err := json.Unmarshal(respBody, &roleResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(projectID + ":" + roleResp.ID)
	data.Name = types.StringValue(roleResp.Name)
	data.ResourceType = types.StringValue(roleResp.ResourceType)
	data.PredefinedRole = types.BoolValue(roleResp.PredefinedRole)

	respPermissions := make([]types.String, len(roleResp.Permissions))
	for i, p := range roleResp.Permissions {
		respPermissions[i] = types.StringValue(p)
	}
	data.Permissions = respPermissions

	if roleResp.Description != nil {
		data.Description = types.StringValue(*roleResp.Description)
	} else {
		data.Description = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be project_id:role_id")
		return
	}
	projectID := idParts[0]
	roleID := idParts[1]

	rolesURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/roles"
	httpClient := &http.Client{Timeout: 30 * time.Second}

	var foundRole *RoleResponseFramework
	cursor := ""

	for foundRole == nil {
		parsedURL, err := url.Parse(rolesURL)
		if err != nil {
			resp.Diagnostics.AddError("Error parsing URL", err.Error())
			return
		}
		q := parsedURL.Query()
		q.Set("limit", "100")
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		apiReq, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating request", err.Error())
			return
		}
		setAdminAuthHeaders(r.client, apiReq)

		apiResp, err := httpClient.Do(apiReq)
		if err != nil {
			resp.Diagnostics.AddError("Error listing project roles", err.Error())
			return
		}

		if apiResp.StatusCode == http.StatusNotFound {
			apiResp.Body.Close()
			resp.State.RemoveResource(ctx)
			return
		}
		if apiResp.StatusCode != http.StatusOK {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("API error listing project roles", fmt.Sprintf("API returned: %s", apiResp.Status))
			return
		}

		var listResp RoleListResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&listResp); err != nil {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("Error parsing project roles response", err.Error())
			return
		}
		apiResp.Body.Close()

		for i := range listResp.Data {
			if listResp.Data[i].ID == roleID {
				foundRole = &listResp.Data[i]
				break
			}
		}

		if foundRole != nil {
			break
		}
		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	if foundRole == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ProjectID = types.StringValue(projectID)
	data.Name = types.StringValue(foundRole.Name)
	data.ResourceType = types.StringValue(foundRole.ResourceType)
	data.PredefinedRole = types.BoolValue(foundRole.PredefinedRole)

	permissions := make([]types.String, len(foundRole.Permissions))
	for i, p := range foundRole.Permissions {
		permissions[i] = types.StringValue(p)
	}
	data.Permissions = permissions

	if foundRole.Description != nil {
		data.Description = types.StringValue(*foundRole.Description)
	} else {
		data.Description = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be project_id:role_id")
		return
	}
	projectID := idParts[0]
	roleID := idParts[1]

	permissions := make([]string, len(data.Permissions))
	for i, p := range data.Permissions {
		permissions[i] = p.ValueString()
	}

	updateReq := RoleUpdateRequest{
		RoleName:    data.Name.ValueString(),
		Permissions: permissions,
		Description: data.Description.ValueString(),
	}

	body, err := json.Marshal(updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	apiURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/roles/" + roleID
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setAdminAuthHeaders(r.client, httpReq)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project role", err.Error())
		return
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API error updating project role", fmt.Sprintf("%s - %s", httpResp.Status, string(respBody)))
		return
	}

	var roleResp RoleResponseFramework
	if err := json.Unmarshal(respBody, &roleResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(projectID + ":" + roleResp.ID)
	data.Name = types.StringValue(roleResp.Name)
	data.ResourceType = types.StringValue(roleResp.ResourceType)
	data.PredefinedRole = types.BoolValue(roleResp.PredefinedRole)

	respPermissions := make([]types.String, len(roleResp.Permissions))
	for i, p := range roleResp.Permissions {
		respPermissions[i] = types.StringValue(p)
	}
	data.Permissions = respPermissions

	if roleResp.Description != nil {
		data.Description = types.StringValue(*roleResp.Description)
	} else {
		data.Description = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be project_id:role_id")
		return
	}
	projectID := idParts[0]
	roleID := idParts[1]

	deleteURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/roles/" + roleID
	deleteReq, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	setAdminAuthHeaders(r.client, deleteReq)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	deleteResp, err := httpClient.Do(deleteReq)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting project role", err.Error())
		return
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusOK && deleteResp.StatusCode != http.StatusNoContent && deleteResp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(deleteResp.Body)
		resp.Diagnostics.AddError("API error deleting project role", fmt.Sprintf("%s - %s", deleteResp.Status, string(body)))
		return
	}
}

func (r *ProjectRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format project_id:role_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), idParts[0])...)
}
