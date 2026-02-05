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
	Role      types.String `tfsdk:"role"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func (r *ProjectGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a group's access to an OpenAI Project. Groups are collections of users that can be synced from an identity provider via SCIM.",

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
			"role": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The role of the group in the project (e.g., 'owner' or 'member').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the group was added to the project.",
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
		"role":     data.Role.ValueString(),
	}

	reqBody, _ := json.Marshal(reqMap)

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

	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK && apiResp.StatusCode != http.StatusCreated {
		respBodyBytes, _ := io.ReadAll(apiResp.Body)
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	var groupResp ProjectGroupResponseFramework
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
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

	for foundGroup == nil {
		parsedURL, _ := url.Parse(baseURL)
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

		apiResp, err := http.DefaultClient.Do(apiReq)
		if err != nil {
			resp.Diagnostics.AddError("Error making request", err.Error())
			return
		}
		defer apiResp.Body.Close()

		if apiResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		if apiResp.StatusCode != http.StatusOK {
			resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s", apiResp.Status))
			return
		}

		var listResp ProjectGroupListResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&listResp); err != nil {
			resp.Diagnostics.AddError("Error parsing response", err.Error())
			return
		}

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

		if !listResp.HasMore || listResp.LastID == "" {
			break
		}
		cursor = listResp.LastID
	}

	if foundGroup == nil {
		// Group not found in project, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	data.ProjectID = types.StringValue(projectID)
	data.GroupID = types.StringValue(foundGroup.GroupID)
	data.GroupName = types.StringValue(foundGroup.GroupName)
	data.Role = types.StringValue(foundGroup.Role)
	data.CreatedAt = types.Int64Value(foundGroup.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update endpoint exists for project groups.
	// Role changes require delete+recreate, which is handled by RequiresReplace on the role attribute.
	resp.Diagnostics.AddError(
		"Update not supported",
		"The OpenAI API does not support updating project group assignments. To change the role, the group must be removed and re-added.",
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

	http.DefaultClient.Do(apiReq)
}

func (r *ProjectGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
