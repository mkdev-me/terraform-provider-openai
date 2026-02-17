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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ProjectUserResource{}
var _ resource.ResourceWithImportState = &ProjectUserResource{}

type ProjectUserResource struct {
	client *OpenAIClient
}

func NewProjectUserResource() resource.Resource {
	return &ProjectUserResource{}
}

func (r *ProjectUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_user"
}

type ProjectUserResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	UserID    types.String `tfsdk:"user_id"`
	RoleIDs   types.Set    `tfsdk:"role_ids"`
	Email     types.String `tfsdk:"email"`
	AddedAt   types.Int64  `tfsdk:"added_at"`
}

func (r *ProjectUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a user's membership and roles in an OpenAI Project. This resource adds a user to a project and assigns them one or more project-level roles.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the project user (project_id:user_id).",
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
			"user_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_ids": schema.SetAttribute{
				Required:            true,
				MarkdownDescription: "Set of project-level role IDs to assign to the user. Must be project roles (e.g. from the `openai_project_role` data source). At least one role is required.",
				ElementType:         types.StringType,
			},
			"email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The email of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"added_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The Unix timestamp (in seconds) when the user was added to the project.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProjectUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	userID := data.UserID.ValueString()
	roleIDs := roleIDsFromSet(data.RoleIDs)
	httpClient := &http.Client{Timeout: 30 * time.Second}

	if len(roleIDs) == 0 {
		resp.Diagnostics.AddError("Invalid Configuration", "At least one role_id is required in role_ids.")
		return
	}

	// Step 1: Add user to project (membership endpoint requires a role name, not ID)
	body, err := json.Marshal(map[string]string{
		"user_id": userID,
		"role":    "member",
	})
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	addURL := adminBaseURL(r.client) + "/v1/organization/projects/" + projectID + "/users"
	httpReq, err := http.NewRequest("POST", addURL, bytes.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setAdminAuthHeaders(r.client, httpReq)

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error adding user to project", err.Error())
		return
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		// If user already exists in project, that's fine — we'll manage roles below
		if !strings.Contains(string(respBody), "already exists in project") {
			resp.Diagnostics.AddError("API error adding user to project", fmt.Sprintf("%s - %s", httpResp.Status, string(respBody)))
			return
		}
	}

	// Read user details from the project (works whether we just added or already existed)
	getUserURL := adminBaseURL(r.client) + "/v1/organization/projects/" + projectID + "/users/" + userID
	getUserReq, err := http.NewRequest("GET", getUserURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	setAdminAuthHeaders(r.client, getUserReq)

	getUserResp, err := httpClient.Do(getUserReq)
	if err != nil {
		resp.Diagnostics.AddError("Error reading project user", err.Error())
		return
	}
	defer getUserResp.Body.Close()

	getUserBody, _ := io.ReadAll(getUserResp.Body)
	if getUserResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API error reading project user", fmt.Sprintf("%s - %s", getUserResp.Status, string(getUserBody)))
		return
	}

	var userResp ProjectUserResponseFramework
	if err := json.Unmarshal(getUserBody, &userResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	// Step 2: Assign all requested roles via the roles endpoint
	for _, roleID := range roleIDs {
		assignBody, err := json.Marshal(RoleAssignRequest{RoleID: roleID})
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling role assign request", err.Error())
			return
		}

		assignURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/users/" + userID + "/roles"
		assignReq, err := http.NewRequest("POST", assignURL, bytes.NewReader(assignBody))
		if err != nil {
			resp.Diagnostics.AddError("Error creating role assign request", err.Error())
			return
		}
		assignReq.Header.Set("Content-Type", "application/json")
		setAdminAuthHeaders(r.client, assignReq)

		assignResp, err := httpClient.Do(assignReq)
		if err != nil {
			resp.Diagnostics.AddError("Error assigning role to user", err.Error())
			return
		}
		defer assignResp.Body.Close()

		if assignResp.StatusCode != http.StatusOK && assignResp.StatusCode != http.StatusCreated {
			assignRespBody, _ := io.ReadAll(assignResp.Body)
			resp.Diagnostics.AddError("API error assigning role to user", fmt.Sprintf("%s - %s", assignResp.Status, string(assignRespBody)))
			return
		}
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", projectID, userResp.ID))
	data.Email = types.StringValue(userResp.Email)
	data.AddedAt = types.Int64Value(userResp.AddedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be project_id:user_id")
		return
	}
	projectID := idParts[0]
	userID := idParts[1]
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Step 1: Verify the user is still in the project
	userURL := adminBaseURL(r.client) + "/v1/organization/projects/" + projectID + "/users/" + userID
	apiReq, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	setAdminAuthHeaders(r.client, apiReq)

	apiResp, err := httpClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error reading project user", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if apiResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API error reading project user", fmt.Sprintf("API returned: %s", apiResp.Status))
		return
	}

	var userResp ProjectUserResponseFramework
	respBody, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBody, &userResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ProjectID = types.StringValue(projectID)
	data.UserID = types.StringValue(userResp.ID)
	data.Email = types.StringValue(userResp.Email)
	data.AddedAt = types.Int64Value(userResp.AddedAt)

	// Step 2: Read all role assignments from the roles endpoint
	rolesURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/users/" + userID + "/roles"
	var allRoleIDs []string
	cursor := ""

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
			resp.Diagnostics.AddError("Error reading user roles", err.Error())
			return
		}

		if rolesResp.StatusCode != http.StatusOK {
			rolesResp.Body.Close()
			resp.Diagnostics.AddError("API error reading user roles", fmt.Sprintf("API returned: %s", rolesResp.Status))
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
		allRoleIDs = []string{}
	}

	data.RoleIDs = roleIDsToSet(allRoleIDs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectUserResourceModel
	var state ProjectUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(state.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be project_id:user_id")
		return
	}
	projectID := idParts[0]
	userID := idParts[1]
	httpClient := &http.Client{Timeout: 30 * time.Second}

	oldRoleIDs := roleIDsFromSet(state.RoleIDs)
	newRoleIDs := roleIDsFromSet(plan.RoleIDs)

	tflog.Info(ctx, "project_user Update", map[string]interface{}{
		"project_id":   projectID,
		"user_id":      userID,
		"old_role_ids": fmt.Sprintf("%v", oldRoleIDs),
		"new_role_ids": fmt.Sprintf("%v", newRoleIDs),
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
			unassignURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/users/" + userID + "/roles/" + id
			tflog.Info(ctx, "Unassigning role from user", map[string]interface{}{"role_id": id})
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

			assignURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/users/" + userID + "/roles"
			tflog.Info(ctx, "Assigning role to user", map[string]interface{}{"role_id": id})
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
	plan.Email = state.Email
	plan.AddedAt = state.AddedAt

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be project_id:user_id")
		return
	}
	projectID := idParts[0]
	userID := idParts[1]
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Step 1: Unassign all roles
	for _, roleID := range roleIDsFromSet(data.RoleIDs) {
		unassignURL := adminBaseURL(r.client) + "/v1/projects/" + projectID + "/users/" + userID + "/roles/" + roleID
		unassignReq, err := http.NewRequest("DELETE", unassignURL, nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating role unassign request", err.Error())
			return
		}
		setAdminAuthHeaders(r.client, unassignReq)

		unassignResp, err := httpClient.Do(unassignReq)
		if err != nil {
			resp.Diagnostics.AddError("Error unassigning role from user", err.Error())
			return
		}
		unassignResp.Body.Close()
		if unassignResp.StatusCode != http.StatusOK && unassignResp.StatusCode != http.StatusNoContent && unassignResp.StatusCode != http.StatusNotFound {
			resp.Diagnostics.AddError("API error unassigning role", fmt.Sprintf("API returned: %s", unassignResp.Status))
			return
		}
	}

	// Step 2: Remove user from project
	removeURL := adminBaseURL(r.client) + "/v1/organization/projects/" + projectID + "/users/" + userID
	removeReq, err := http.NewRequest("DELETE", removeURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	setAdminAuthHeaders(r.client, removeReq)

	removeResp, err := httpClient.Do(removeReq)
	if err != nil {
		resp.Diagnostics.AddError("Error removing user from project", err.Error())
		return
	}
	defer removeResp.Body.Close()

	if removeResp.StatusCode != http.StatusOK && removeResp.StatusCode != http.StatusNoContent && removeResp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(removeResp.Body)
		resp.Diagnostics.AddError("API error removing user from project", fmt.Sprintf("%s - %s", removeResp.Status, string(body)))
		return
	}
}

func (r *ProjectUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
