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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectGroupResource{}
var _ resource.ResourceWithImportState = &ProjectGroupResource{}

type ProjectGroupResource struct {
	client *OpenAIClient
}

func NewProjectGroupResource() resource.Resource {
	return &ProjectGroupResource{}
}

func (r *ProjectGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_group"
}

type ProjectGroupResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	GroupID   types.String `tfsdk:"group_id"`
	GroupName types.String `tfsdk:"group_name"`
	RoleID    types.String `tfsdk:"role_id"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func (r *ProjectGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a group's membership in an OpenAI Project. Groups are collections of users that can be synced from an identity provider via SCIM.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the project group (project_id:group_id).",
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
			"group_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the group to add to the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The display name of the group.",
			},
			"role_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the project role to grant to the group (e.g., 'role_01J1F8PROJ'). Required when creating. This is write-only - the API does not return this value, so it will be unknown after import.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The Unix timestamp (in seconds) when the group was added to the project.",
			},
		},
	}
}

func (r *ProjectGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add group to project
	// POST /organization/projects/{project_id}/groups
	reqMap := map[string]interface{}{
		"group_id": data.GroupID.ValueString(),
		"role":     data.RoleID.ValueString(),
	}

	reqBody, err := json.Marshal(reqMap)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	apiURL := r.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/projects/" + data.ProjectID.ValueString() + "/groups"
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/projects/" + data.ProjectID.ValueString() + "/groups"
	}

	apiReq, err := http.NewRequest("POST", reqURL, bytes.NewReader(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}
	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	apiResp, err := httpClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK && apiResp.StatusCode != http.StatusCreated {
		respBodyBytes, readErr := io.ReadAll(apiResp.Body)
		if readErr != nil {
			resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s (failed to read body: %s)", apiResp.Status, readErr.Error()))
			return
		}
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	var groupResp ProjectGroupResponseFramework
	respBodyBytes, err := io.ReadAll(apiResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Error reading response", err.Error())
		return
	}
	if err := json.Unmarshal(respBodyBytes, &groupResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", data.ProjectID.ValueString(), groupResp.GroupID))
	data.GroupName = types.StringValue(groupResp.GroupName)
	data.CreatedAt = types.Int64Value(groupResp.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be project_id:group_id")
		return
	}
	projectID := idParts[0]
	groupID := idParts[1]

	// No single-group GET endpoint exists, so we must list and filter
	apiURL := r.client.OpenAIClient.APIURL
	var baseURL string
	if strings.Contains(apiURL, "/v1") {
		baseURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/projects/" + projectID + "/groups"
	} else {
		baseURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/projects/" + projectID + "/groups"
	}

	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}

	var foundGroup *ProjectGroupResponseFramework
	cursor := ""
	httpClient := &http.Client{Timeout: 30 * time.Second}

	for foundGroup == nil {
		parsedURL, err := url.Parse(baseURL)
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

		apiReq.Header.Set("Authorization", "Bearer "+apiKey)
		if r.client.OpenAIClient.OrganizationID != "" {
			apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
		}

		apiResp, err := httpClient.Do(apiReq)
		if err != nil {
			resp.Diagnostics.AddError("Error making request", err.Error())
			return
		}

		if apiResp.StatusCode == http.StatusNotFound {
			apiResp.Body.Close()
			resp.State.RemoveResource(ctx)
			return
		}
		if apiResp.StatusCode != http.StatusOK {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s", apiResp.Status))
			return
		}

		var listResp ProjectGroupListResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&listResp); err != nil {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("Error parsing response", err.Error())
			return
		}
		apiResp.Body.Close()

		for i := range listResp.Data {
			group := listResp.Data[i]
			if group.GroupID == groupID {
				foundGroup = &group
				break
			}
		}

		if foundGroup != nil {
			break
		}

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	if foundGroup == nil {
		// Group not found in project, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	data.ProjectID = types.StringValue(projectID)
	data.GroupID = types.StringValue(foundGroup.GroupID)
	data.GroupName = types.StringValue(foundGroup.GroupName)
	data.CreatedAt = types.Int64Value(foundGroup.CreatedAt)
	// Note: role_id is write-only and not returned by the API, so we preserve the state value

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update endpoint exists for project groups.
	// Changes require delete+recreate, which is handled by RequiresReplace on the attributes.
	resp.Diagnostics.AddError(
		"Update not supported",
		"The OpenAI API does not support updating project group assignments. The group must be removed and re-added.",
	)
}

func (r *ProjectGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be project_id:group_id")
		return
	}
	projectID := idParts[0]
	groupID := idParts[1]

	apiURL := r.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/projects/" + projectID + "/groups/" + groupID
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/projects/" + projectID + "/groups/" + groupID
	}

	apiReq, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}
	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	apiResp, err := httpClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer apiResp.Body.Close()

	// Accept 200, 204, or 404 (already deleted) as success
	if apiResp.StatusCode != http.StatusOK && apiResp.StatusCode != http.StatusNoContent && apiResp.StatusCode != http.StatusNotFound {
		respBodyBytes, readErr := io.ReadAll(apiResp.Body)
		if readErr != nil {
			resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s (failed to read body: %s)", apiResp.Status, readErr.Error()))
			return
		}
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}
}

func (r *ProjectGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
