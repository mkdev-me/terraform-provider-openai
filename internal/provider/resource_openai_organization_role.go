package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &OrganizationRoleResource{}
var _ resource.ResourceWithImportState = &OrganizationRoleResource{}

type OrganizationRoleResource struct {
	client *OpenAIClient
}

func NewOrganizationRoleResource() resource.Resource {
	return &OrganizationRoleResource{}
}

func (r *OrganizationRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_role"
}

type OrganizationRoleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Permissions    types.List   `tfsdk:"permissions"`
	Description    types.String `tfsdk:"description"`
	ResourceType   types.String `tfsdk:"resource_type"`
	PredefinedRole types.Bool   `tfsdk:"predefined_role"`
}

func (r *OrganizationRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom organization role in the OpenAI organization.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the organization role.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the organization role.",
			},
			"permissions": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The list of permissions granted by this role.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the organization role.",
			},
			"resource_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource type the role is bound to (e.g., 'api.organization').",
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

func (r *OrganizationRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permissions []string
	resp.Diagnostics.Append(data.Permissions.ElementsAs(ctx, &permissions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := RoleCreateRequest{
		RoleName:    data.Name.ValueString(),
		Permissions: permissions,
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		createReq.Description = data.Description.ValueString()
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	apiURL := adminBaseURL(r.client) + "/v1/organization/roles"
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
		resp.Diagnostics.AddError("Error creating organization role", err.Error())
		return
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("API error creating organization role", fmt.Sprintf("%s - %s", httpResp.Status, string(respBody)))
		return
	}

	var roleResp RoleResponseFramework
	if err := json.Unmarshal(respBody, &roleResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(roleResp.ID)
	data.Name = types.StringValue(roleResp.Name)
	data.ResourceType = types.StringValue(roleResp.ResourceType)
	data.PredefinedRole = types.BoolValue(roleResp.PredefinedRole)

	if roleResp.Description != nil {
		data.Description = types.StringValue(*roleResp.Description)
	}

	if len(roleResp.Permissions) > 0 {
		permList, diags := types.ListValueFrom(ctx, types.StringType, roleResp.Permissions)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Permissions = permList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := data.ID.ValueString()
	rolesURL := adminBaseURL(r.client) + "/v1/organization/roles"
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
			resp.Diagnostics.AddError("Error listing organization roles", err.Error())
			return
		}

		if apiResp.StatusCode == http.StatusNotFound {
			apiResp.Body.Close()
			resp.State.RemoveResource(ctx)
			return
		}
		if apiResp.StatusCode != http.StatusOK {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("API error listing organization roles", fmt.Sprintf("API returned: %s", apiResp.Status))
			return
		}

		var listResp RoleListResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&listResp); err != nil {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("Error parsing organization roles response", err.Error())
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

	data.ID = types.StringValue(foundRole.ID)
	data.Name = types.StringValue(foundRole.Name)
	data.ResourceType = types.StringValue(foundRole.ResourceType)
	data.PredefinedRole = types.BoolValue(foundRole.PredefinedRole)

	if foundRole.Description != nil {
		data.Description = types.StringValue(*foundRole.Description)
	} else {
		data.Description = types.StringNull()
	}

	if len(foundRole.Permissions) > 0 {
		permList, diags := types.ListValueFrom(ctx, types.StringType, foundRole.Permissions)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Permissions = permList
	} else {
		data.Permissions = types.ListValueMust(types.StringType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrganizationRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state OrganizationRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := state.ID.ValueString()

	var permissions []string
	resp.Diagnostics.Append(data.Permissions.ElementsAs(ctx, &permissions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := RoleUpdateRequest{
		RoleName:    data.Name.ValueString(),
		Permissions: permissions,
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		updateReq.Description = data.Description.ValueString()
	}

	body, err := json.Marshal(updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	apiURL := adminBaseURL(r.client) + "/v1/organization/roles/" + roleID
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
		resp.Diagnostics.AddError("Error updating organization role", err.Error())
		return
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API error updating organization role", fmt.Sprintf("%s - %s", httpResp.Status, string(respBody)))
		return
	}

	var roleResp RoleResponseFramework
	if err := json.Unmarshal(respBody, &roleResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(roleResp.ID)
	data.Name = types.StringValue(roleResp.Name)
	data.ResourceType = types.StringValue(roleResp.ResourceType)
	data.PredefinedRole = types.BoolValue(roleResp.PredefinedRole)

	if roleResp.Description != nil {
		data.Description = types.StringValue(*roleResp.Description)
	}

	if len(roleResp.Permissions) > 0 {
		permList, diags := types.ListValueFrom(ctx, types.StringType, roleResp.Permissions)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Permissions = permList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := data.ID.ValueString()

	deleteURL := adminBaseURL(r.client) + "/v1/organization/roles/" + roleID
	deleteReq, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	setAdminAuthHeaders(r.client, deleteReq)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	deleteResp, err := httpClient.Do(deleteReq)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting organization role", err.Error())
		return
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusOK && deleteResp.StatusCode != http.StatusNoContent && deleteResp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(deleteResp.Body)
		resp.Diagnostics.AddError("API error deleting organization role", fmt.Sprintf("%s - %s", deleteResp.Status, string(body)))
		return
	}
}

func (r *OrganizationRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
