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

// OrganizationUserRolesDataSource - list data source for organization user role assignments

var _ datasource.DataSource = &OrganizationUserRolesDataSource{}

func NewOrganizationUserRolesDataSource() datasource.DataSource {
	return &OrganizationUserRolesDataSource{}
}

type OrganizationUserRolesDataSource struct {
	client *OpenAIClient
}

type OrganizationUserRolesDataSourceModel struct {
	UserID          types.String              `tfsdk:"user_id"`
	RoleAssignments []UserRoleAssignmentModel `tfsdk:"role_assignments"`
	RoleIDs         []types.String            `tfsdk:"role_ids"`
	ID              types.String              `tfsdk:"id"`
}

type UserRoleAssignmentModel struct {
	RoleID          types.String   `tfsdk:"role_id"`
	RoleName        types.String   `tfsdk:"role_name"`
	RoleDescription types.String   `tfsdk:"role_description"`
	Permissions     []types.String `tfsdk:"permissions"`
	ResourceType    types.String   `tfsdk:"resource_type"`
	PredefinedRole  types.Bool     `tfsdk:"predefined_role"`
}

func (d *OrganizationUserRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_user_roles"
}

func (d *OrganizationUserRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all roles assigned to a specific user within the OpenAI organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to retrieve role assignments for.",
				Required:    true,
			},
			"role_assignments": schema.ListNestedAttribute{
				Description: "List of role assignments for the user.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
					},
				},
			},
			"role_ids": schema.ListAttribute{
				Description: "List of role IDs assigned to the user.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *OrganizationUserRolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationUserRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationUserRolesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key (admin_key) to read organization user role assignments.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	baseURL := strings.TrimSuffix(apiURL, "/v1")
	baseURL = strings.TrimSuffix(baseURL, "/")
	reqURL := fmt.Sprintf("%s/v1/organization/users/%s/roles", baseURL, userID)

	// Initialize slices to empty to avoid nil (which becomes null in state)
	allAssignments := make([]UserRoleAssignmentModel, 0)
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

		for _, role := range listResp.Data {
			permissions := make([]types.String, len(role.Permissions))
			for i, p := range role.Permissions {
				permissions[i] = types.StringValue(p)
			}

			description := ""
			if role.Description != nil {
				description = *role.Description
			}

			assignmentModel := UserRoleAssignmentModel{
				RoleID:          types.StringValue(role.ID),
				RoleName:        types.StringValue(role.Name),
				RoleDescription: types.StringValue(description),
				Permissions:     permissions,
				ResourceType:    types.StringValue(role.ResourceType),
				PredefinedRole:  types.BoolValue(role.PredefinedRole),
			}
			allAssignments = append(allAssignments, assignmentModel)
			roleIDs = append(roleIDs, role.ID)
		}

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	data.ID = types.StringValue(userID)
	data.RoleAssignments = allAssignments

	data.RoleIDs = make([]types.String, len(roleIDs))
	for i, v := range roleIDs {
		data.RoleIDs[i] = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
