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

// GroupUserDataSource - singular data source

var _ datasource.DataSource = &GroupUserDataSource{}

func NewGroupUserDataSource() datasource.DataSource {
	return &GroupUserDataSource{}
}

type GroupUserDataSource struct {
	client *OpenAIClient
}

type GroupUserDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	GroupID  types.String `tfsdk:"group_id"`
	UserID   types.String `tfsdk:"user_id"`
	UserName types.String `tfsdk:"user_name"`
	Email    types.String `tfsdk:"email"`
	Role     types.String `tfsdk:"role"`
	AddedAt  types.Int64  `tfsdk:"added_at"`
}

func (d *GroupUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_user"
}

func (d *GroupUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific user in an OpenAI organization group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource (group_id:user_id).",
				Computed:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the group.",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user. Either user_id or email must be provided.",
				Optional:    true,
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email of the user to search for. Either user_id or email must be provided.",
				Optional:    true,
				Computed:    true,
			},
			"user_name": schema.StringAttribute{
				Description: "The name of the user.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The user's organization role.",
				Computed:    true,
			},
			"added_at": schema.Int64Attribute{
				Description: "Unix timestamp when the user was added to the organization.",
				Computed:    true,
			},
		},
	}
}

func (d *GroupUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenAIClient, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *GroupUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupUserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := data.GroupID.ValueString()
	userID := data.UserID.ValueString()
	email := data.Email.ValueString()

	if userID == "" && email == "" {
		resp.Diagnostics.AddError(
			"Missing User Identifier",
			"Either user_id or email must be provided.",
		)
		return
	}

	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key (admin_key) to read group users.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups/" + groupID + "/users"
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups/" + groupID + "/users"
	}

	var foundUser *OrganizationUserResponseFramework
	cursor := ""

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
		if d.client.OpenAIClient.OrganizationID != "" {
			httpRequest.Header.Set("OpenAI-Organization", d.client.OpenAIClient.OrganizationID)
		}

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

		var listResp GroupUserListResponse
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

		if !listResp.HasMore {
			break
		}
		if len(listResp.Data) > 0 {
			cursor = listResp.Data[len(listResp.Data)-1].ID
		} else {
			break
		}
	}

	if foundUser == nil {
		identifier := userID
		if identifier == "" {
			identifier = email
		}
		resp.Diagnostics.AddError(
			"User Not Found",
			fmt.Sprintf("User %s not found in group %s", identifier, groupID),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", groupID, foundUser.ID))
	data.UserID = types.StringValue(foundUser.ID)
	data.UserName = types.StringValue(foundUser.Name)
	data.Email = types.StringValue(foundUser.Email)
	data.Role = types.StringValue(foundUser.Role)
	data.AddedAt = types.Int64Value(foundUser.AddedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// GroupUsersDataSource - list data source

func NewGroupUsersDataSource() datasource.DataSource {
	return &GroupUsersDataSource{}
}

type GroupUsersDataSource struct {
	client *OpenAIClient
}

type GroupUsersDataSourceModel struct {
	ID        types.String           `tfsdk:"id"`
	GroupID   types.String           `tfsdk:"group_id"`
	Users     []GroupUserResultModel `tfsdk:"users"`
	UserIDs   []types.String         `tfsdk:"user_ids"`
	UserCount types.Int64            `tfsdk:"user_count"`
}

type GroupUserResultModel struct {
	UserID   types.String `tfsdk:"user_id"`
	UserName types.String `tfsdk:"user_name"`
	Email    types.String `tfsdk:"email"`
	Role     types.String `tfsdk:"role"`
	AddedAt  types.Int64  `tfsdk:"added_at"`
}

func (d *GroupUsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_users"
}

func (d *GroupUsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all users in an OpenAI organization group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the group to retrieve users from.",
				Required:    true,
			},
			"users": schema.ListNestedAttribute{
				Description: "List of users in the group.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							Description: "The ID of the user.",
							Computed:    true,
						},
						"user_name": schema.StringAttribute{
							Description: "The name of the user.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email of the user.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "The user's organization role.",
							Computed:    true,
						},
						"added_at": schema.Int64Attribute{
							Description: "Unix timestamp when the user was added to the organization.",
							Computed:    true,
						},
					},
				},
			},
			"user_ids": schema.ListAttribute{
				Description: "List of user IDs in the group.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"user_count": schema.Int64Attribute{
				Description: "Number of users in the group.",
				Computed:    true,
			},
		},
	}
}

func (d *GroupUsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenAIClient, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *GroupUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupUsersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups/" + groupID + "/users"
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups/" + groupID + "/users"
	}

	var allUsers []GroupUserResultModel
	var userIDs []string

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
		if d.client.OpenAIClient.OrganizationID != "" {
			httpRequest.Header.Set("OpenAI-Organization", d.client.OpenAIClient.OrganizationID)
		}

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

		var listResp GroupUserListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}

		for _, u := range listResp.Data {
			userModel := GroupUserResultModel{
				UserID:   types.StringValue(u.ID),
				UserName: types.StringValue(u.Name),
				Email:    types.StringValue(u.Email),
				Role:     types.StringValue(u.Role),
				AddedAt:  types.Int64Value(u.AddedAt),
			}
			allUsers = append(allUsers, userModel)
			userIDs = append(userIDs, u.ID)
		}

		if !listResp.HasMore {
			break
		}
		if len(listResp.Data) > 0 {
			cursor = listResp.Data[len(listResp.Data)-1].ID
		} else {
			break
		}
	}

	data.ID = types.StringValue(groupID)
	data.Users = allUsers
	data.UserCount = types.Int64Value(int64(len(allUsers)))

	data.UserIDs = make([]types.String, len(userIDs))
	for i, v := range userIDs {
		data.UserIDs[i] = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
