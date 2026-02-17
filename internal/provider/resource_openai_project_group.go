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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	RoleIDs   types.Set    `tfsdk:"role_ids"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func (r *ProjectGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a group's membership and roles in an OpenAI Project. This resource adds a group to a project and assigns it one or more project-level roles.",

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_ids": schema.SetAttribute{
				Required:            true,
				MarkdownDescription: "Set of project-level role IDs to assign to the group. Must be project roles (e.g. from the `openai_project_role` data source), not organization roles. At least one role is required.",
				ElementType:         types.StringType,
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The Unix timestamp (in seconds) when the group was added to the project.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
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

// roleIDsFromSet extracts a []string from a types.Set of strings.
func roleIDsFromSet(s types.Set) []string {
	elements := s.Elements()
	result := make([]string, len(elements))
	for i, elem := range elements {
		result[i] = elem.(types.String).ValueString()
	}
	return result
}

// roleIDsToSet converts a []string to a types.Set of strings.
func roleIDsToSet(ids []string) types.Set {
	elements := make([]attr.Value, len(ids))
	for i, id := range ids {
		elements[i] = types.StringValue(id)
	}
	setValue, _ := types.SetValue(types.StringType, elements)
	return setValue
}

func (r *ProjectGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	groupID := data.GroupID.ValueString()
	roleIDs := roleIDsFromSet(data.RoleIDs)
	httpClient := &http.Client{Timeout: 30 * time.Second}

	if len(roleIDs) == 0 {
		resp.Diagnostics.AddError("Invalid Configuration", "At least one role_id is required in role_ids.")
		return
	}

	// Step 1: Add group to project (group membership endpoint accepts a role ID)
	body, err := json.Marshal(map[string]string{
		"group_id": groupID,
		"role":     roleIDs[0],
	})
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	addURL := adminBaseURL(r.client) + "/v1/organization/projects/" + projectID + "/groups"
	httpReq, err := http.NewRequest("POST", addURL, bytes.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setAdminAuthHeaders(r.client, httpReq)

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error adding group to project", err.Error())
		return
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		// If group already exists in project, that's fine — we'll manage roles below
		if !strings.Contains(string(respBody), "already exists in project") {
			resp.Diagnostics.AddError("API error adding group to project", fmt.Sprintf("%s - %s", httpResp.Status, string(respBody)))
			return
		}
	}

	// Read group details from the project (works whether we just added or already existed)
	var groupResp ProjectGroupResponseFramework
	groupsURL := adminBaseURL(r.client) + "/v1/organization/projects/" + projectID + "/groups"
	cursor := ""
	for {
		parsedURL, err := url.Parse(groupsURL)
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

		listReq, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating request", err.Error())
			return
		}
		setAdminAuthHeaders(r.client, listReq)

		listResp, err := httpClient.Do(listReq)
		if err != nil {
			resp.Diagnostics.AddError("Error listing project groups", err.Error())
			return
		}

		if listResp.StatusCode != http.StatusOK {
			listResp.Body.Close()
			resp.Diagnostics.AddError("API error listing project groups", fmt.Sprintf("API returned: %s", listResp.Status))
			return
		}

		var listData ProjectGroupListResponse
		if err := json.NewDecoder(listResp.Body).Decode(&listData); err != nil {
			listResp.Body.Close()
			resp.Diagnostics.AddError("Error parsing response", err.Error())
			return
		}
		listResp.Body.Close()

		found := false
		for i := range listData.Data {
			if listData.Data[i].GroupID == groupID {
				groupResp = listData.Data[i]
				found = true
				break
			}
		}
		if found {
			break
		}
		if !listData.HasMore || listData.Next == nil {
			resp.Diagnostics.AddError("Group not found", fmt.Sprintf("Group %s not found in project %s after adding", groupID, projectID))
			return
		}
		cursor = *listData.Next
	}

	// Step 2: Assign additional roles (first one was set via membership endpoint)
	for _, roleID := range roleIDs[1:] {
		assignBody, err := json.Marshal(RoleAssignRequest{RoleID: roleID})
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling role assign request", err.Error())
			return
		}

		assignURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/groups/" + groupID + "/roles"
		assignReq, err := http.NewRequest("POST", assignURL, bytes.NewReader(assignBody))
		if err != nil {
			resp.Diagnostics.AddError("Error creating role assign request", err.Error())
			return
		}
		assignReq.Header.Set("Content-Type", "application/json")
		setAdminAuthHeaders(r.client, assignReq)

		assignResp, err := httpClient.Do(assignReq)
		if err != nil {
			resp.Diagnostics.AddError("Error assigning role to group", err.Error())
			return
		}
		defer assignResp.Body.Close()

		if assignResp.StatusCode != http.StatusOK && assignResp.StatusCode != http.StatusCreated {
			assignRespBody, _ := io.ReadAll(assignResp.Body)
			resp.Diagnostics.AddError("API error assigning role to group", fmt.Sprintf("%s - %s", assignResp.Status, string(assignRespBody)))
			return
		}
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", projectID, groupResp.GroupID))
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
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Step 1: Verify the group is still in the project
	groupsURL := adminBaseURL(r.client) + "/v1/organization/projects/" + projectID + "/groups"

	var foundGroup *ProjectGroupResponseFramework
	cursor := ""

	for foundGroup == nil {
		parsedURL, err := url.Parse(groupsURL)
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
			resp.Diagnostics.AddError("Error listing project groups", err.Error())
			return
		}

		if apiResp.StatusCode == http.StatusNotFound {
			apiResp.Body.Close()
			resp.State.RemoveResource(ctx)
			return
		}
		if apiResp.StatusCode != http.StatusOK {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("API error listing project groups", fmt.Sprintf("API returned: %s", apiResp.Status))
			return
		}

		var listResp ProjectGroupListResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&listResp); err != nil {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("Error parsing project groups response", err.Error())
			return
		}
		apiResp.Body.Close()

		for i := range listResp.Data {
			if listResp.Data[i].GroupID == groupID {
				foundGroup = &listResp.Data[i]
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
		resp.State.RemoveResource(ctx)
		return
	}

	data.ProjectID = types.StringValue(projectID)
	data.GroupID = types.StringValue(foundGroup.GroupID)
	data.GroupName = types.StringValue(foundGroup.GroupName)
	data.CreatedAt = types.Int64Value(foundGroup.CreatedAt)

	// Step 2: Read all role assignments from the roles endpoint
	rolesURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/groups/" + groupID + "/roles"
	var allRoleIDs []string
	cursor = ""

	for {
		parsedURL, err := url.Parse(rolesURL)
		if err != nil {
			resp.Diagnostics.AddError("Error parsing roles URL", err.Error())
			return
		}
		q := parsedURL.Query()
		q.Set("limit", "100")
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		rolesReq, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating roles request", err.Error())
			return
		}
		setAdminAuthHeaders(r.client, rolesReq)

		rolesResp, err := httpClient.Do(rolesReq)
		if err != nil {
			resp.Diagnostics.AddError("Error reading group roles", err.Error())
			return
		}

		if rolesResp.StatusCode != http.StatusOK {
			rolesResp.Body.Close()
			resp.Diagnostics.AddError("API error reading group roles", fmt.Sprintf("API returned: %s", rolesResp.Status))
			return
		}

		var roleListResp RoleListResponse
		if err := json.NewDecoder(rolesResp.Body).Decode(&roleListResp); err != nil {
			rolesResp.Body.Close()
			resp.Diagnostics.AddError("Error parsing roles response", err.Error())
			return
		}
		rolesResp.Body.Close()

		for _, role := range roleListResp.Data {
			allRoleIDs = append(allRoleIDs, role.ID)
		}

		if !roleListResp.HasMore || roleListResp.Next == nil {
			break
		}
		cursor = *roleListResp.Next
	}

	if len(allRoleIDs) == 0 {
		// No roles found — the resource was modified outside Terraform
		allRoleIDs = []string{}
	}

	data.RoleIDs = roleIDsToSet(allRoleIDs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectGroupResourceModel
	var state ProjectGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(state.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be project_id:group_id")
		return
	}
	projectID := idParts[0]
	groupID := idParts[1]
	httpClient := &http.Client{Timeout: 30 * time.Second}

	oldRoleIDs := roleIDsFromSet(state.RoleIDs)
	newRoleIDs := roleIDsFromSet(plan.RoleIDs)

	tflog.Info(ctx, "project_group Update", map[string]interface{}{
		"project_id":    projectID,
		"group_id":      groupID,
		"old_role_ids":  fmt.Sprintf("%v", oldRoleIDs),
		"new_role_ids":  fmt.Sprintf("%v", newRoleIDs),
		"state_null":    state.RoleIDs.IsNull(),
		"state_unknown": state.RoleIDs.IsUnknown(),
		"plan_null":     plan.RoleIDs.IsNull(),
		"plan_unknown":  plan.RoleIDs.IsUnknown(),
	})

	// Build sets for diffing
	oldSet := make(map[string]bool, len(oldRoleIDs))
	for _, id := range oldRoleIDs {
		oldSet[id] = true
	}
	newSet := make(map[string]bool, len(newRoleIDs))
	for _, id := range newRoleIDs {
		newSet[id] = true
	}

	// Roles to remove (in old but not in new)
	for _, id := range oldRoleIDs {
		if !newSet[id] {
			unassignURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/groups/" + groupID + "/roles/" + id
			tflog.Info(ctx, "Unassigning role", map[string]interface{}{"url": unassignURL, "role_id": id})
			unassignReq, err := http.NewRequest("DELETE", unassignURL, nil)
			if err != nil {
				resp.Diagnostics.AddError("Error creating role unassign request", err.Error())
				return
			}
			setAdminAuthHeaders(r.client, unassignReq)

			unassignResp, err := httpClient.Do(unassignReq)
			if err != nil {
				resp.Diagnostics.AddError("Error unassigning role", err.Error())
				return
			}
			respBody, _ := io.ReadAll(unassignResp.Body)
			unassignResp.Body.Close()
			tflog.Info(ctx, "Unassign response", map[string]interface{}{"status": unassignResp.StatusCode, "body": string(respBody)})
			if unassignResp.StatusCode != http.StatusOK && unassignResp.StatusCode != http.StatusNoContent && unassignResp.StatusCode != http.StatusNotFound {
				resp.Diagnostics.AddError("API error unassigning role", fmt.Sprintf("%s - %s", unassignResp.Status, string(respBody)))
				return
			}
		}
	}

	// Roles to add (in new but not in old)
	for _, id := range newRoleIDs {
		if !oldSet[id] {
			assignBody, err := json.Marshal(RoleAssignRequest{RoleID: id})
			if err != nil {
				resp.Diagnostics.AddError("Error marshaling role assign request", err.Error())
				return
			}

			assignURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/groups/" + groupID + "/roles"
			tflog.Info(ctx, "Assigning role", map[string]interface{}{"url": assignURL, "role_id": id})
			assignReq, err := http.NewRequest("POST", assignURL, bytes.NewReader(assignBody))
			if err != nil {
				resp.Diagnostics.AddError("Error creating role assign request", err.Error())
				return
			}
			assignReq.Header.Set("Content-Type", "application/json")
			setAdminAuthHeaders(r.client, assignReq)

			assignResp, err := httpClient.Do(assignReq)
			if err != nil {
				resp.Diagnostics.AddError("Error assigning role", err.Error())
				return
			}
			assignRespBody, _ := io.ReadAll(assignResp.Body)
			assignResp.Body.Close()
			tflog.Info(ctx, "Assign response", map[string]interface{}{"status": assignResp.StatusCode, "body": string(assignRespBody)})

			if assignResp.StatusCode != http.StatusOK && assignResp.StatusCode != http.StatusCreated {
				resp.Diagnostics.AddError("API error assigning role", fmt.Sprintf("%s - %s", assignResp.Status, string(assignRespBody)))
				return
			}
		}
	}

	plan.ID = state.ID
	plan.GroupName = state.GroupName
	plan.CreatedAt = state.CreatedAt

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Step 1: Unassign all roles
	for _, roleID := range roleIDsFromSet(data.RoleIDs) {
		unassignURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/groups/" + groupID + "/roles/" + roleID
		unassignReq, err := http.NewRequest("DELETE", unassignURL, nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating role unassign request", err.Error())
			return
		}
		setAdminAuthHeaders(r.client, unassignReq)

		unassignResp, err := httpClient.Do(unassignReq)
		if err != nil {
			resp.Diagnostics.AddError("Error unassigning role from group", err.Error())
			return
		}
		unassignResp.Body.Close()
		// Ignore 404 — role may already be gone
		if unassignResp.StatusCode != http.StatusOK && unassignResp.StatusCode != http.StatusNoContent && unassignResp.StatusCode != http.StatusNotFound {
			resp.Diagnostics.AddError("API error unassigning role", fmt.Sprintf("API returned: %s", unassignResp.Status))
			return
		}
	}

	// Step 2: Remove group from project
	removeURL := adminBaseURL(r.client) + "/v1/organization/projects/" + projectID + "/groups/" + groupID
	removeReq, err := http.NewRequest("DELETE", removeURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	setAdminAuthHeaders(r.client, removeReq)

	removeResp, err := httpClient.Do(removeReq)
	if err != nil {
		resp.Diagnostics.AddError("Error removing group from project", err.Error())
		return
	}
	defer removeResp.Body.Close()

	if removeResp.StatusCode != http.StatusOK && removeResp.StatusCode != http.StatusNoContent && removeResp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(removeResp.Body)
		resp.Diagnostics.AddError("API error removing group from project", fmt.Sprintf("%s - %s", removeResp.Status, string(body)))
		return
	}
}

func (r *ProjectGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
