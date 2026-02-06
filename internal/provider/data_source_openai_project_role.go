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

// ProjectRolesDataSource - list data source for project roles

var _ datasource.DataSource = &ProjectRolesDataSource{}

func NewProjectRolesDataSource() datasource.DataSource {
	return &ProjectRolesDataSource{}
}

type ProjectRolesDataSource struct {
	client *OpenAIClient
}

type ProjectRolesDataSourceModel struct {
	ProjectID types.String             `tfsdk:"project_id"`
	Roles     []ProjectRoleResultModel `tfsdk:"roles"`
	RoleIDs   []types.String           `tfsdk:"role_ids"`
	RoleCount types.Int64              `tfsdk:"role_count"`
	ID        types.String             `tfsdk:"id"`
}

type ProjectRoleResultModel struct {
	RoleID         types.String   `tfsdk:"role_id"`
	Name           types.String   `tfsdk:"name"`
	Description    types.String   `tfsdk:"description"`
	Permissions    []types.String `tfsdk:"permissions"`
	ResourceType   types.String   `tfsdk:"resource_type"`
	PredefinedRole types.Bool     `tfsdk:"predefined_role"`
}

func (d *ProjectRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_roles"
}

func (d *ProjectRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all roles configured for a specific OpenAI project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to retrieve roles from.",
				Required:    true,
			},
			"roles": schema.ListNestedAttribute{
				Description: "List of roles in the project.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role_id": schema.StringAttribute{
							Description: "The ID of the role.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the role.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the role.",
							Computed:    true,
						},
						"permissions": schema.ListAttribute{
							Description: "Permissions granted by the role.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"resource_type": schema.StringAttribute{
							Description: "Resource type the role is bound to (e.g., 'api.project').",
							Computed:    true,
						},
						"predefined_role": schema.BoolAttribute{
							Description: "Whether the role is predefined and managed by OpenAI.",
							Computed:    true,
						},
					},
				},
			},
			"role_ids": schema.ListAttribute{
				Description: "List of role IDs in the project.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"role_count": schema.Int64Attribute{
				Description: "Number of roles in the project.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectRolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectRolesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API Key is required.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	suffix := fmt.Sprintf("/organization/projects/%s/roles", projectID)

	// Safely construct URL by trimming both /v1 and trailing /
	baseURL := strings.TrimSuffix(apiURL, "/v1")
	baseURL = strings.TrimSuffix(baseURL, "/")
	reqURL := baseURL + "/v1" + suffix

	// Initialize slices to empty to avoid nil (which becomes null in state)
	allRoles := make([]ProjectRoleResultModel, 0)
	roleIDs := make([]string, 0)

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

		var listResp RoleListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			httpResp.Body.Close()
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}
		httpResp.Body.Close()

		for _, r := range listResp.Data {
			permissions := make([]types.String, len(r.Permissions))
			for i, p := range r.Permissions {
				permissions[i] = types.StringValue(p)
			}

			description := ""
			if r.Description != nil {
				description = *r.Description
			}

			roleModel := ProjectRoleResultModel{
				RoleID:         types.StringValue(r.ID),
				Name:           types.StringValue(r.Name),
				Description:    types.StringValue(description),
				Permissions:    permissions,
				ResourceType:   types.StringValue(r.ResourceType),
				PredefinedRole: types.BoolValue(r.PredefinedRole),
			}
			allRoles = append(allRoles, roleModel)
			roleIDs = append(roleIDs, r.ID)
		}

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	data.ID = types.StringValue(projectID)
	data.Roles = allRoles
	data.RoleCount = types.Int64Value(int64(len(allRoles)))

	data.RoleIDs = make([]types.String, len(roleIDs))
	for i, v := range roleIDs {
		data.RoleIDs[i] = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
