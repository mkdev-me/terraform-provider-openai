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

var _ datasource.DataSource = &OrganizationUserDataSource{}

func NewOrganizationUserDataSource() datasource.DataSource {
	return &OrganizationUserDataSource{}
}

type OrganizationUserDataSource struct {
	client *OpenAIClient
}

type OrganizationUserDataSourceModel struct {
	UserID  types.String `tfsdk:"user_id"`
	Email   types.String `tfsdk:"email"`
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Role    types.String `tfsdk:"role"`
	AddedAt types.Int64  `tfsdk:"added_at"`
}

func (d *OrganizationUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_user"
}

func (d *OrganizationUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific user in your organization.",
		Attributes: map[string]schema.Attribute{
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to retrieve.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the user to retrieve.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the user.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the user.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The role of the user (owner, member, or reader).",
				Computed:    true,
			},
			"added_at": schema.Int64Attribute{
				Description: "The Unix timestamp when the user was added to the organization.",
				Computed:    true,
			},
		},
	}
}

func (d *OrganizationUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationUserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()
	email := data.Email.ValueString()

	if userID == "" && email == "" {
		resp.Diagnostics.AddError(
			"Missing Identifier",
			"Either user_id or email must be provided.",
		)
		return
	}

	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key to manage organization users.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	// /v1/organization/users
	// or /v1/organization/users/{user_id}

	var foundUser *OrganizationUserResponseFramework

	if userID != "" {
		// Fetch by ID directly
		suffix := fmt.Sprintf("/organization/users/%s", userID)
		var reqURL string
		if strings.Contains(apiURL, "/v1") {
			reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
		} else {
			reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
		}

		httpRequest, err := http.NewRequest("GET", reqURL, nil)
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
			if httpResp.StatusCode == 404 {
				resp.Diagnostics.AddError("User Not Found", fmt.Sprintf("User with ID %s not found", userID))
				return
			}
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
			return
		}

		var user OrganizationUserResponseFramework
		if err := json.NewDecoder(httpResp.Body).Decode(&user); err != nil {
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}
		foundUser = &user
	} else {
		// Find by Email - must list all users
		suffix := "/organization/users"
		var reqURL string
		if strings.Contains(apiURL, "/v1") {
			reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
		} else {
			reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
		}

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

			type OrgUserListResponse struct {
				Object  string                              `json:"object"`
				Data    []OrganizationUserResponseFramework `json:"data"`
				FirstID string                              `json:"first_id"`
				LastID  string                              `json:"last_id"`
				HasMore bool                                `json:"has_more"`
			}
			var listResp OrgUserListResponse
			if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
				resp.Diagnostics.AddError("Error decoding response", err.Error())
				return
			}

			for i := range listResp.Data {
				u := listResp.Data[i]
				if strings.EqualFold(u.Email, email) {
					foundUser = &u
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
			resp.Diagnostics.AddError("User Not Found", fmt.Sprintf("User with email %s not found", email))
			return
		}
	}

	data.ID = types.StringValue(foundUser.ID)
	// Preserve input in state if present, otherwise set from response
	if !data.UserID.IsNull() {
		data.UserID = types.StringValue(foundUser.ID)
	}
	if !data.Email.IsNull() {
		data.Email = types.StringValue(foundUser.Email)
	}

	// If only ID was provided, enrich Email
	if data.Email.IsNull() {
		data.Email = types.StringValue(foundUser.Email)
	}
	// If only Email was provided, enrich ID
	if data.UserID.IsNull() {
		data.UserID = types.StringValue(foundUser.ID)
	}

	data.Name = types.StringValue(foundUser.Name)
	data.Role = types.StringValue(foundUser.Role)
	data.AddedAt = types.Int64Value(foundUser.AddedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// OrganizationUsersDataSource

func NewOrganizationUsersDataSource() datasource.DataSource {
	return &OrganizationUsersDataSource{}
}

type OrganizationUsersDataSource struct {
	client *OpenAIClient
}

type OrganizationUsersDataSourceModel struct {
	UserID types.String                  `tfsdk:"user_id"`
	Emails types.List                    `tfsdk:"emails"` // Changed from []types.String to types.List
	Users  []OrganizationUserResultModel `tfsdk:"users"`
	ID     types.String                  `tfsdk:"id"`
}

type OrganizationUserResultModel struct {
	ID      types.String `tfsdk:"id"`
	Object  types.String `tfsdk:"object"`
	Email   types.String `tfsdk:"email"`
	Name    types.String `tfsdk:"name"`
	Role    types.String `tfsdk:"role"`
	AddedAt types.Int64  `tfsdk:"added_at"`
}

func (d *OrganizationUsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_users"
}

func (d *OrganizationUsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all users in your organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of a specific user to retrieve. If provided, other filter parameters are ignored.",
				Optional:    true,
			},
			"emails": schema.ListAttribute{
				Description: "Filter by the email address of users.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"users": schema.ListNestedAttribute{
				Description: "List of users in the organization.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the user.",
							Computed:    true,
						},
						"object": schema.StringAttribute{
							Description: "The object type, always 'organization.user'.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email address of the user.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the user.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "The role of the user.",
							Computed:    true,
						},
						"added_at": schema.Int64Attribute{
							Description: "The Unix timestamp when the user was added to the organization.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationUsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationUsersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key to manage organization users.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	var allUsers []OrganizationUserResultModel

	// If user_id provided, fetch single user
	if !data.UserID.IsNull() {
		// Similar to single user fetch
		userID := data.UserID.ValueString()
		suffix := fmt.Sprintf("/organization/users/%s", userID)
		var reqURL string
		if strings.Contains(apiURL, "/v1") {
			reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
		} else {
			reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
		}

		httpRequest, _ := http.NewRequest("GET", reqURL, nil)
		httpRequest.Header.Set("Authorization", "Bearer "+adminKey)
		httpClient := &http.Client{}
		httpResp, err := httpClient.Do(httpRequest)
		if err != nil {
			resp.Diagnostics.AddError("Error executing request", err.Error())
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode == 200 {
			var user OrganizationUserResponseFramework
			if err := json.NewDecoder(httpResp.Body).Decode(&user); err == nil {
				allUsers = append(allUsers, OrganizationUserResultModel{
					ID:      types.StringValue(user.ID),
					Object:  types.StringValue(user.Object),
					Email:   types.StringValue(user.Email),
					Name:    types.StringValue(user.Name),
					Role:    types.StringValue(user.Role),
					AddedAt: types.Int64Value(user.AddedAt),
				})
			}
		} else if httpResp.StatusCode != 404 {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
			return
		} else {
			resp.Diagnostics.AddError("User Not Found", "User with ID "+userID+" not found")
			return
		}

		data.ID = types.StringValue(fmt.Sprintf("organization-user-%s", userID))
	} else {
		// List all users
		suffix := "/organization/users"
		var reqURL string
		if strings.Contains(apiURL, "/v1") {
			reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
		} else {
			reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
		}

		targetEmails := make(map[string]bool)
		if !data.Emails.IsNull() {
			var emails []string
			// Convert List to []string
			data.Emails.ElementsAs(ctx, &emails, false)
			for _, e := range emails {
				targetEmails[strings.ToLower(e)] = true
			}
		}

		cursor := ""
		for {
			parsedURL, _ := url.Parse(reqURL)
			q := parsedURL.Query()
			q.Set("limit", "100")
			if cursor != "" {
				q.Set("after", cursor)
			}
			parsedURL.RawQuery = q.Encode()

			httpRequest, _ := http.NewRequest("GET", parsedURL.String(), nil)
			httpRequest.Header.Set("Authorization", "Bearer "+adminKey)
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

			type OrgUserListResponse struct {
				Object  string                              `json:"object"`
				Data    []OrganizationUserResponseFramework `json:"data"`
				FirstID string                              `json:"first_id"`
				LastID  string                              `json:"last_id"`
				HasMore bool                                `json:"has_more"`
			}
			var listResp OrgUserListResponse
			if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
				resp.Diagnostics.AddError("Error decoding response", err.Error())
				return
			}

			for _, u := range listResp.Data {
				if len(targetEmails) > 0 {
					if !targetEmails[strings.ToLower(u.Email)] {
						continue
					}
				}
				allUsers = append(allUsers, OrganizationUserResultModel{
					ID:      types.StringValue(u.ID),
					Object:  types.StringValue(u.Object),
					Email:   types.StringValue(u.Email),
					Name:    types.StringValue(u.Name),
					Role:    types.StringValue(u.Role),
					AddedAt: types.Int64Value(u.AddedAt),
				})
			}

			if !listResp.HasMore || listResp.LastID == "" {
				break
			}
			cursor = listResp.LastID
		}

		data.ID = types.StringValue(fmt.Sprintf("organization-users-all-%d", len(allUsers)))
	}

	data.Users = allUsers

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
