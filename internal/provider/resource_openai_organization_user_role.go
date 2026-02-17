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

var _ resource.Resource = &OrganizationUserRoleResource{}
var _ resource.ResourceWithImportState = &OrganizationUserRoleResource{}

type OrganizationUserRoleResource struct {
	client *OpenAIClient
}

func NewOrganizationUserRoleResource() resource.Resource {
	return &OrganizationUserRoleResource{}
}

type OrganizationUserRoleResourceModel struct {
	ID     types.String `tfsdk:"id"`
	UserID types.String `tfsdk:"user_id"`
	RoleID types.String `tfsdk:"role_id"`
}

func (r *OrganizationUserRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_user_role"
}

func (r *OrganizationUserRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assigns a role to a user at the organization level.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the organization user role assignment (user_id:role_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the role to assign.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *OrganizationUserRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationUserRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationUserRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()
	roleID := data.RoleID.ValueString()

	body, err := json.Marshal(RoleAssignRequest{RoleID: roleID})
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	apiURL := adminBaseURL(r.client) + "/v1/organization/users/" + userID + "/roles"
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
		resp.Diagnostics.AddError("Error assigning role to user", err.Error())
		return
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("API error assigning role to user", fmt.Sprintf("%s - %s", httpResp.Status, string(respBody)))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", userID, roleID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationUserRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationUserRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be user_id:role_id")
		return
	}
	userID := idParts[0]
	roleID := idParts[1]

	rolesURL := adminBaseURL(r.client) + "/v1/organization/users/" + userID + "/roles"
	httpClient := &http.Client{Timeout: 30 * time.Second}

	found := false
	cursor := ""

	for !found {
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
			resp.Diagnostics.AddError("Error listing user roles", err.Error())
			return
		}

		if apiResp.StatusCode == http.StatusNotFound {
			apiResp.Body.Close()
			resp.State.RemoveResource(ctx)
			return
		}
		if apiResp.StatusCode != http.StatusOK {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("API error listing user roles", fmt.Sprintf("API returned: %s", apiResp.Status))
			return
		}

		var listResp RoleListResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&listResp); err != nil {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("Error parsing user roles response", err.Error())
			return
		}
		apiResp.Body.Close()

		for _, role := range listResp.Data {
			if role.ID == roleID {
				found = true
				break
			}
		}

		if found {
			break
		}
		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.UserID = types.StringValue(userID)
	data.RoleID = types.StringValue(roleID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationUserRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unexpected Update", "All attributes require replacement; Update should not be called.")
}

func (r *OrganizationUserRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationUserRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be user_id:role_id")
		return
	}
	userID := idParts[0]
	roleID := idParts[1]

	deleteURL := adminBaseURL(r.client) + "/v1/organization/users/" + userID + "/roles/" + roleID
	deleteReq, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	setAdminAuthHeaders(r.client, deleteReq)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	deleteResp, err := httpClient.Do(deleteReq)
	if err != nil {
		resp.Diagnostics.AddError("Error removing role from user", err.Error())
		return
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusOK && deleteResp.StatusCode != http.StatusNoContent && deleteResp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(deleteResp.Body)
		resp.Diagnostics.AddError("API error removing role from user", fmt.Sprintf("%s - %s", deleteResp.Status, string(body)))
		return
	}
}

func (r *OrganizationUserRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "ID must be user_id:role_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), idParts[1])...)
}
