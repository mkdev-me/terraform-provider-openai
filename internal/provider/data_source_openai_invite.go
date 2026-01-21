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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

var _ datasource.DataSource = &InviteDataSource{}

func NewInviteDataSource() datasource.DataSource {
	return &InviteDataSource{}
}

type InviteDataSource struct {
	client *OpenAIClient
}

type InviteDataSourceModel struct {
	InviteID  types.String                   `tfsdk:"invite_id"`
	ID        types.String                   `tfsdk:"id"`
	Email     types.String                   `tfsdk:"email"`
	Role      types.String                   `tfsdk:"role"`
	Status    types.String                   `tfsdk:"status"`
	CreatedAt types.Int64                    `tfsdk:"created_at"`
	ExpiresAt types.Int64                    `tfsdk:"expires_at"`
	Projects  []InviteDataSourceProjectModel `tfsdk:"projects"`
}

type InviteDataSourceProjectModel struct {
	ID   types.String `tfsdk:"id"`
	Role types.String `tfsdk:"role"`
}

func (d *InviteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

func (d *InviteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific invitation in an OpenAI organization.",
		Attributes: map[string]schema.Attribute{
			"invite_id": schema.StringAttribute{
				Description: "The ID of the invitation to retrieve.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the invitation.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the invited user.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The role assigned to the invited user (owner or reader).",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the invitation.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "When the invitation was created (Unix timestamp).",
				Computed:    true,
			},
			"expires_at": schema.Int64Attribute{
				Description: "When the invitation expires (Unix timestamp).",
				Computed:    true,
			},
			"projects": schema.ListNestedAttribute{
				Description: "Projects assigned to the invited user.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the project.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "The role assigned to the user within the project (owner or member).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *InviteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InviteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InviteDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inviteID := data.InviteID.ValueString()
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API Key is required.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	// /v1/organization/invites/{invite_id}
	suffix := fmt.Sprintf("/organization/invites/%s", inviteID)

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
			resp.Diagnostics.AddError("Invite Not Found", fmt.Sprintf("Invite with ID %s not found", inviteID))
			return
		}
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
		return
	}

	type InviteProject struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	type InviteResponse struct {
		ID        string          `json:"id"`
		Email     string          `json:"email"`
		Role      string          `json:"role"`
		Status    string          `json:"status"`
		CreatedAt int64           `json:"created_at"`
		ExpiresAt int64           `json:"expires_at"`
		Projects  []InviteProject `json:"projects"`
	}

	var invite InviteResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&invite); err != nil {
		resp.Diagnostics.AddError("Error decoding response", err.Error())
		return
	}

	data.ID = types.StringValue(invite.ID)
	data.Email = types.StringValue(invite.Email)
	data.Role = types.StringValue(invite.Role)
	data.Status = types.StringValue(invite.Status)
	data.CreatedAt = types.Int64Value(invite.CreatedAt)
	data.ExpiresAt = types.Int64Value(invite.ExpiresAt)

	if len(invite.Projects) > 0 {
		projects := make([]InviteDataSourceProjectModel, len(invite.Projects))
		for i, p := range invite.Projects {
			projects[i] = InviteDataSourceProjectModel{
				ID:   types.StringValue(p.ID),
				Role: types.StringValue(p.Role),
			}
		}
		data.Projects = projects
	} else {
		data.Projects = []InviteDataSourceProjectModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// InvitesDataSource

func NewInvitesDataSource() datasource.DataSource {
	return &InvitesDataSource{}
}

type InvitesDataSource struct {
	client *OpenAIClient
}

type InvitesDataSourceModel struct {
	Invites []InviteResultModel `tfsdk:"invites"`
	ID      types.String        `tfsdk:"id"`
}

type InviteResultModel struct {
	ID        types.String                   `tfsdk:"id"`
	Email     types.String                   `tfsdk:"email"`
	Role      types.String                   `tfsdk:"role"`
	Status    types.String                   `tfsdk:"status"`
	CreatedAt types.String                   `tfsdk:"created_at"` // String in plural
	ExpiresAt types.String                   `tfsdk:"expires_at"` // String in plural
	Projects  []InviteDataSourceProjectModel `tfsdk:"projects"`
}

func (d *InvitesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invites"
}

func (d *InvitesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all pending invitations in an OpenAI organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"invites": schema.ListNestedAttribute{
				Description: "List of pending invitations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the invitation.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email address of the invited user.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "The role assigned to the invited user.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The status of the invitation.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the invitation was created.",
							Computed:    true,
						},
						"expires_at": schema.StringAttribute{
							Description: "Timestamp when the invitation expires.",
							Computed:    true,
						},
						"projects": schema.ListNestedAttribute{
							Description: "Projects assigned to the invited user.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description: "The ID of the project.",
										Computed:    true,
									},
									"role": schema.StringAttribute{
										Description: "The role assigned to the user within the project.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *InvitesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InvitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InvitesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API Key is required.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	// /v1/organization/invites
	suffix := "/organization/invites"

	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
	}

	var allInvites []InviteResultModel
	cursor := ""

	for {
		// Retry wrapper for request
		var listResp *struct {
			Data []struct {
				ID        string `json:"id"`
				Email     string `json:"email"`
				Role      string `json:"role"`
				Status    string `json:"status"`
				CreatedAt int64  `json:"created_at"`
				ExpiresAt int64  `json:"expires_at"`
				Projects  []struct {
					ID   string `json:"id"`
					Role string `json:"role"`
				} `json:"projects"`
			} `json:"data"`
			HasMore bool   `json:"has_more"`
			LastID  string `json:"last_id"`
		}

		opErr := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
			parsedURL, _ := url.Parse(reqURL)
			q := parsedURL.Query()
			q.Set("limit", "100")
			if cursor != "" {
				q.Set("after", cursor)
			}
			parsedURL.RawQuery = q.Encode()

			httpRequest, err := http.NewRequest("GET", parsedURL.String(), nil)
			if err != nil {
				return retry.NonRetryableError(err)
			}
			httpRequest.Header.Set("Authorization", "Bearer "+adminKey)
			httpRequest.Header.Set("Content-Type", "application/json")

			httpClient := &http.Client{Timeout: 30 * time.Second}
			httpResp, err := httpClient.Do(httpRequest)
			if err != nil {
				return retry.RetryableError(err)
			}

			if httpResp.StatusCode == 504 || httpResp.StatusCode == 408 {
				httpResp.Body.Close()
				return retry.RetryableError(fmt.Errorf("API timeout: %s", httpResp.Status))
			}
			if httpResp.StatusCode != 200 {
				defer httpResp.Body.Close()
				return retry.NonRetryableError(fmt.Errorf("API Error: %s", httpResp.Status))
			}

			defer httpResp.Body.Close()
			var pageResp struct {
				Data []struct {
					ID        string `json:"id"`
					Email     string `json:"email"`
					Role      string `json:"role"`
					Status    string `json:"status"`
					CreatedAt int64  `json:"created_at"`
					ExpiresAt int64  `json:"expires_at"`
					Projects  []struct {
						ID   string `json:"id"`
						Role string `json:"role"`
					} `json:"projects"`
				} `json:"data"`
				HasMore bool   `json:"has_more"`
				LastID  string `json:"last_id"`
			}
			if err := json.NewDecoder(httpResp.Body).Decode(&pageResp); err != nil {
				return retry.NonRetryableError(err)
			}
			listResp = &pageResp
			return nil
		})

		if opErr != nil {
			if strings.Contains(opErr.Error(), "timeout") || strings.Contains(opErr.Error(), "504") {
				resp.Diagnostics.AddWarning("OpenAI API Timeout", "The OpenAI API timed out while retrieving invitations.")
				data.Invites = []InviteResultModel{} // Return empty
				data.ID = types.StringValue(fmt.Sprintf("invites_%d", time.Now().Unix()))
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			resp.Diagnostics.AddError("Error listing invites", opErr.Error())
			return
		}

		for _, inv := range listResp.Data {
			model := InviteResultModel{
				ID:        types.StringValue(inv.ID),
				Email:     types.StringValue(inv.Email),
				Role:      types.StringValue(inv.Role),
				Status:    types.StringValue(inv.Status),
				CreatedAt: types.StringValue(time.Unix(inv.CreatedAt, 0).Format(time.RFC3339)),
				ExpiresAt: types.StringValue(time.Unix(inv.ExpiresAt, 0).Format(time.RFC3339)),
			}
			if len(inv.Projects) > 0 {
				projs := make([]InviteDataSourceProjectModel, len(inv.Projects))
				for i, p := range inv.Projects {
					projs[i] = InviteDataSourceProjectModel{
						ID:   types.StringValue(p.ID),
						Role: types.StringValue(p.Role),
					}
				}
				model.Projects = projs
			} else {
				model.Projects = []InviteDataSourceProjectModel{}
			}
			allInvites = append(allInvites, model)
		}

		if !listResp.HasMore || listResp.LastID == "" {
			break
		}
		cursor = listResp.LastID
	}

	data.ID = types.StringValue(fmt.Sprintf("invites_%d", time.Now().Unix()))
	data.Invites = allInvites

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
