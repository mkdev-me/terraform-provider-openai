package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProjectGroupRolesDataSource - list data source for project group role assignments

var _ datasource.DataSource = &ProjectGroupRolesDataSource{}

func NewProjectGroupRolesDataSource() datasource.DataSource {
	return &ProjectGroupRolesDataSource{}
}

type ProjectGroupRolesDataSource struct {
	client *OpenAIClient
}

type ProjectGroupRolesDataSourceModel struct {
	ProjectID       types.String                       `tfsdk:"project_id"`
	GroupID         types.String                       `tfsdk:"group_id"`
	RoleAssignments []ProjectGroupRoleAssignmentModel  `tfsdk:"role_assignments"`
	RoleIDs         []types.String                     `tfsdk:"role_ids"`
	ID              types.String                       `tfsdk:"id"`
}

type ProjectGroupRoleAssignmentModel struct {
	AssignmentID   types.String   `tfsdk:"assignment_id"`
	RoleID         types.String   `tfsdk:"role_id"`
	RoleName       types.String   `tfsdk:"role_name"`
	RoleDescription types.String  `tfsdk:"role_description"`
	Permissions    []types.String `tfsdk:"permissions"`
	ResourceType   types.String   `tfsdk:"resource_type"`
	PredefinedRole types.Bool     `tfsdk:"predefined_role"`
	GroupID        types.String   `tfsdk:"group_id"`
	GroupName      types.String   `tfsdk:"group_name"`
}

func (d *ProjectGroupRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_group_roles"
}

func (d *ProjectGroupRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all roles assigned to a specific group within an OpenAI project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource (project_id:group_id).",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project.",
				Required:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the group to retrieve role assignments for.",
				Required:    true,
			},
			"role_assignments": schema.ListNestedAttribute{
				Description: "List of role assignments for the group.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"assignment_id": schema.StringAttribute{
							Description: "The ID of the role assignment.",
							Computed:    true,
						},
						"role_id": schema.StringAttribute{
							Description: "The ID of the role.",
							Computed:    true,
						},
						"role_name": schema.StringAttribute{
							Description: "The name of the role.",
							Computed:    true,
						},
						"role_description": schema.StringAttribute{
							Description: "The description of the role.",
							Computed:    true,
						},
						"permissions": schema.ListAttribute{
							Description: "Permissions granted by the role.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"resource_type": schema.StringAttribute{
							Description: "Resource type the role is bound to.",
							Computed:    true,
						},
						"predefined_role": schema.BoolAttribute{
							Description: "Whether the role is predefined and managed by OpenAI.",
							Computed:    true,
						},
						"group_id": schema.StringAttribute{
							Description: "The ID of the group.",
							Computed:    true,
						},
						"group_name": schema.StringAttribute{
							Description: "The name of the group.",
							Computed:    true,
						},
					},
				},
			},
			"role_ids": schema.ListAttribute{
				Description: "List of role IDs assigned to the group.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *ProjectGroupRolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*OpenAIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ProjectGroupRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectGroupRolesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	groupID := data.GroupID.ValueString()
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API Key is required.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	suffix := fmt.Sprintf("/projects/%s/groups/%s/roles", projectID, groupID)

	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
	}

	var allAssignments []ProjectGroupRoleAssignmentModel
	var roleIDs []string

	cursor := ""
	httpClient := &http.Client{Timeout: 30 * time.Second}
	for {
		parsedURL, err := url.Parse(reqURL)
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

		httpRequest, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating request", err.Error())
			return
		}
		httpRequest.Header.Set("Authorization", "Bearer "+adminKey)
		httpRequest.Header.Set("Content-Type", "application/json")
		if d.client.OpenAIClient.OrganizationID != "" {
			httpRequest.Header.Set("OpenAI-Organization", d.client.OpenAIClient.OrganizationID)
		}

		httpResp, err := httpClient.Do(httpRequest)
		if err != nil {
			resp.Diagnostics.AddError("Error executing request", err.Error())
			return
		}

		if httpResp.StatusCode != 200 {
			httpResp.Body.Close()
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
			return
		}

		var listResp GroupRoleAssignmentListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			httpResp.Body.Close()
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}
		httpResp.Body.Close()

		for _, a := range listResp.Data {
			permissions := make([]types.String, len(a.Role.Permissions))
			for i, p := range a.Role.Permissions {
				permissions[i] = types.StringValue(p)
			}

			description := ""
			if a.Role.Description != nil {
				description = *a.Role.Description
			}

			assignmentModel := ProjectGroupRoleAssignmentModel{
				AssignmentID:   types.StringValue(a.ID),
				RoleID:         types.StringValue(a.Role.ID),
				RoleName:       types.StringValue(a.Role.Name),
				RoleDescription: types.StringValue(description),
				Permissions:    permissions,
				ResourceType:   types.StringValue(a.Role.ResourceType),
				PredefinedRole: types.BoolValue(a.Role.PredefinedRole),
				GroupID:        types.StringValue(a.Group.ID),
				GroupName:      types.StringValue(a.Group.Name),
			}
			allAssignments = append(allAssignments, assignmentModel)
			roleIDs = append(roleIDs, a.Role.ID)
		}

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", projectID, groupID))
	data.RoleAssignments = allAssignments

	data.RoleIDs = make([]types.String, len(roleIDs))
	for i, v := range roleIDs {
		data.RoleIDs[i] = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
