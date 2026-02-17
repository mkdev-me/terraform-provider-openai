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

// RolesDataSource - list data source for organization roles

var _ datasource.DataSource = &RolesDataSource{}

func NewRolesDataSource() datasource.DataSource {
	return &RolesDataSource{}
}

type RolesDataSource struct {
	client *OpenAIClient
}

type RolesDataSourceModel struct {
	Roles     []RoleResultModel `tfsdk:"roles"`
	RoleIDs   []types.String    `tfsdk:"role_ids"`
	RoleCount types.Int64       `tfsdk:"role_count"`
	ID        types.String      `tfsdk:"id"`
}

type RoleResultModel struct {
	RoleID         types.String   `tfsdk:"role_id"`
	Name           types.String   `tfsdk:"name"`
	Description    types.String   `tfsdk:"description"`
	Permissions    []types.String `tfsdk:"permissions"`
	ResourceType   types.String   `tfsdk:"resource_type"`
	PredefinedRole types.Bool     `tfsdk:"predefined_role"`
}

func (d *RolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *RolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all roles configured for the OpenAI organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"roles": schema.ListNestedAttribute{
				Description: "List of roles in the organization.",
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
							Description: "Resource type the role is bound to (e.g., 'api.organization').",
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
				Description: "List of role IDs in the organization.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"role_count": schema.Int64Attribute{
				Description: "Number of roles in the organization.",
				Computed:    true,
			},
		},
	}
}

func (d *RolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RolesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key (admin_key) to read organization roles.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	baseURL := strings.TrimSuffix(apiURL, "/v1")
	baseURL = strings.TrimSuffix(baseURL, "/")
	reqURL := baseURL + "/v1/organization/roles"

	// Initialize slices to empty to avoid nil (which becomes null in state)
	allRoles := make([]RoleResultModel, 0)
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

			roleModel := RoleResultModel{
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

	data.ID = types.StringValue("organization-roles")
	data.Roles = allRoles
	data.RoleCount = types.Int64Value(int64(len(allRoles)))

	data.RoleIDs = make([]types.String, len(roleIDs))
	for i, v := range roleIDs {
		data.RoleIDs[i] = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// RoleDataSource - singular data source for looking up an organization role by name

var _ datasource.DataSource = &RoleDataSource{}

func NewRoleDataSource() datasource.DataSource {
	return &RoleDataSource{}
}

type RoleDataSource struct {
	client *OpenAIClient
}

type RoleDataSourceModel struct {
	Name           types.String   `tfsdk:"name"`
	RoleID         types.String   `tfsdk:"role_id"`
	Description    types.String   `tfsdk:"description"`
	Permissions    []types.String `tfsdk:"permissions"`
	ResourceType   types.String   `tfsdk:"resource_type"`
	PredefinedRole types.Bool     `tfsdk:"predefined_role"`
	ID             types.String   `tfsdk:"id"`
}

func (d *RoleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *RoleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a specific role by name within the OpenAI organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the role to look up (case-insensitive).",
				Required:    true,
			},
			"role_id": schema.StringAttribute{
				Description: "The ID of the role.",
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
				Description: "Resource type the role is bound to (e.g., 'api.organization').",
				Computed:    true,
			},
			"predefined_role": schema.BoolAttribute{
				Description: "Whether the role is predefined and managed by OpenAI.",
				Computed:    true,
			},
		},
	}
}

func (d *RoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoleDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleName := data.Name.ValueString()

	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key (admin_key) to read organization roles.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	baseURL := strings.TrimSuffix(apiURL, "/v1")
	baseURL = strings.TrimSuffix(baseURL, "/")
	reqURL := baseURL + "/v1/organization/roles"

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
			if strings.EqualFold(r.Name, roleName) {
				permissions := make([]types.String, len(r.Permissions))
				for i, p := range r.Permissions {
					permissions[i] = types.StringValue(p)
				}

				description := ""
				if r.Description != nil {
					description = *r.Description
				}

				data.RoleID = types.StringValue(r.ID)
				data.Name = types.StringValue(r.Name)
				data.Description = types.StringValue(description)
				data.Permissions = permissions
				data.ResourceType = types.StringValue(r.ResourceType)
				data.PredefinedRole = types.BoolValue(r.PredefinedRole)
				data.ID = types.StringValue(r.ID)

				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
		}

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	resp.Diagnostics.AddError(
		"Role Not Found",
		fmt.Sprintf("No role with name %q found in the organization.", roleName),
	)
}
