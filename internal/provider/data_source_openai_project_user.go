package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectUserDataSource{}

func NewProjectUserDataSource() datasource.DataSource {
	return &ProjectUserDataSource{}
}

type ProjectUserDataSource struct {
	client *OpenAIClient
}

type ProjectUserDataSourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	UserID    types.String `tfsdk:"user_id"`
	Email     types.String `tfsdk:"email"`
	ID        types.String `tfsdk:"id"`
	Role      types.String `tfsdk:"role"`
	AddedAt   types.Int64  `tfsdk:"added_at"`
}

func (d *ProjectUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_user"
}

func (d *ProjectUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific user in an OpenAI project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to retrieve the user from.",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to retrieve.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the user to retrieve.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the resource (composite of project_id:user_id).",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The role of the user in the project (owner or member).",
				Computed:    true,
			},
			"added_at": schema.Int64Attribute{
				Description: "Timestamp when the user was added to the project.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectUserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	userID := data.UserID.ValueString()
	email := data.Email.ValueString()

	if userID == "" && email == "" {
		resp.Diagnostics.AddError(
			"Missing User Identifier",
			"Either user_id or email must be provided.",
		)
		return
	}

	// We need Admin Key
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key (admin_key) to read project users.",
		)
		return
	}

	// Helper to fetch all users and find logic
	// Since we don't have a direct "Get Project User by Email" API, we likely need to list and filter.
	// Even for ID, checking existence via list is common if Get endpoint doesn't exist or we want to be safe.
	// Actually typical API is GET /v1/organization/projects/{project_id}/users/{user_id}
	// Let's check if that endpoint exists.
	// Legacy code uses `ListProjectUsers` for both ID and Email lookup.
	// This implies there might NOT be a direct GET endpoint or legacy implementation preferred listing.
	// Checking `dataSourceFindProjectUser` in legacy: it loops `ListProjectUsers`.
	// So we must replicate that.

	apiURL := d.client.OpenAIClient.APIURL
	// /v1/organization/projects/{project_id}/users
	suffix := fmt.Sprintf("/organization/projects/%s/users", projectID)

	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
	}

	var foundUser *ProjectUserResponseFramework
	cursor := ""

	// Loop until found or done
	for foundUser == nil {
		parsedURL, _ := url.Parse(reqURL)
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

		httpClient := &http.Client{}
		httpResp, err := httpClient.Do(httpRequest)
		if err != nil {
			resp.Diagnostics.AddError("Error executing request", err.Error())
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != 200 {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
			return
		}

		// We need a list response struct for Project Users
		// I'll define it here locally or check if I defined it earlier.
		// ProjectUserResponseFramework is in types_project_org.go.
		type ProjectUserListResponse struct {
			Object  string                         `json:"object"`
			Data    []ProjectUserResponseFramework `json:"data"`
			FirstID string                         `json:"first_id"`
			LastID  string                         `json:"last_id"`
			HasMore bool                           `json:"has_more"`
		}

		var listResp ProjectUserListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}

		for i := range listResp.Data {
			user := listResp.Data[i]
			if userID != "" && user.ID == userID {
				foundUser = &user
				break
			}
			if email != "" && strings.EqualFold(user.Email, email) {
				foundUser = &user
				break
			}
		}

		if foundUser != nil {
			break
		}

		if !listResp.HasMore || listResp.LastID == "" {
			break
		}
		cursor = listResp.LastID
	}

	if foundUser == nil {
		identifier := userID
		if identifier == "" {
			identifier = email
		}
		resp.Diagnostics.AddError(
			"User Not Found",
			fmt.Sprintf("User %s not found in project %s", identifier, projectID),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", projectID, foundUser.ID))
	data.UserID = types.StringValue(foundUser.ID) // Store actual ID even if lookup was by email
	data.Email = types.StringValue(foundUser.Email)
	data.Role = types.StringValue(foundUser.Role)
	data.AddedAt = types.Int64Value(foundUser.AddedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ProjectUsersDataSource

func NewProjectUsersDataSource() datasource.DataSource {
	return &ProjectUsersDataSource{}
}

type ProjectUsersDataSource struct {
	client *OpenAIClient
}

type ProjectUsersDataSourceModel struct {
	ProjectID types.String             `tfsdk:"project_id"`
	Users     []ProjectUserResultModel `tfsdk:"users"`
	UserIDs   []types.String           `tfsdk:"user_ids"`
	UserCount types.Int64              `tfsdk:"user_count"`
	OwnerIDs  []types.String           `tfsdk:"owner_ids"`
	MemberIDs []types.String           `tfsdk:"member_ids"`
	ID        types.String             `tfsdk:"id"`
}

type ProjectUserResultModel struct {
	ID      types.String `tfsdk:"id"`
	Email   types.String `tfsdk:"email"`
	Role    types.String `tfsdk:"role"`
	AddedAt types.Int64  `tfsdk:"added_at"`
}

func (d *ProjectUsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_users"
}

func (d *ProjectUsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all users in a specific OpenAI project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to retrieve users from.",
				Required:    true,
			},
			"users": schema.ListNestedAttribute{
				Description: "List of users in the project.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the user.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email address of the user.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "The role of the user (owner or member).",
							Computed:    true,
						},
						"added_at": schema.Int64Attribute{
							Description: "Timestamp when the user was added to the project.",
							Computed:    true,
						},
					},
				},
			},
			"user_ids": schema.ListAttribute{
				Description: "List of user IDs in the project.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"user_count": schema.Int64Attribute{
				Description: "Number of users in the project.",
				Computed:    true,
			},
			"owner_ids": schema.ListAttribute{
				Description: "List of user IDs with owner role.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"member_ids": schema.ListAttribute{
				Description: "List of user IDs with member role.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *ProjectUsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectUsersDataSourceModel

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
	// /v1/organization/projects/{project_id}/users
	suffix := fmt.Sprintf("/organization/projects/%s/users", projectID)

	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
	}

	var allUsers []ProjectUserResultModel
	var userIDs []string
	var ownerIDs []string
	var memberIDs []string

	cursor := ""
	for {
		parsedURL, _ := url.Parse(reqURL)
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

		httpClient := &http.Client{}
		httpResp, err := httpClient.Do(httpRequest)
		if err != nil {
			resp.Diagnostics.AddError("Error executing request", err.Error())
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != 200 {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
			return
		}

		type ProjectUserListResponse struct {
			Object  string                         `json:"object"`
			Data    []ProjectUserResponseFramework `json:"data"`
			FirstID string                         `json:"first_id"`
			LastID  string                         `json:"last_id"`
			HasMore bool                           `json:"has_more"`
		}

		var listResp ProjectUserListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}

		for _, u := range listResp.Data {
			userModel := ProjectUserResultModel{
				ID:      types.StringValue(u.ID),
				Email:   types.StringValue(u.Email),
				Role:    types.StringValue(u.Role),
				AddedAt: types.Int64Value(u.AddedAt),
			}
			allUsers = append(allUsers, userModel)
			userIDs = append(userIDs, u.ID)

			if u.Role == "owner" {
				ownerIDs = append(ownerIDs, u.ID)
			} else if u.Role == "member" {
				memberIDs = append(memberIDs, u.ID)
			}
		}

		if !listResp.HasMore || listResp.LastID == "" {
			break
		}
		cursor = listResp.LastID
	}

	data.ID = types.StringValue(projectID)
	data.Users = allUsers
	data.UserCount = types.Int64Value(int64(len(allUsers)))

	data.UserIDs = make([]types.String, len(userIDs))
	for i, v := range userIDs {
		data.UserIDs[i] = types.StringValue(v)
	}

	data.OwnerIDs = make([]types.String, len(ownerIDs))
	for i, v := range ownerIDs {
		data.OwnerIDs[i] = types.StringValue(v)
	}

	data.MemberIDs = make([]types.String, len(memberIDs))
	for i, v := range memberIDs {
		data.MemberIDs[i] = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ProjectUserResponseFramework defined in types_project_org.go
