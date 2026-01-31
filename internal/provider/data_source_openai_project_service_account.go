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

var _ datasource.DataSource = &ProjectServiceAccountDataSource{}

func NewProjectServiceAccountDataSource() datasource.DataSource {
	return &ProjectServiceAccountDataSource{}
}

type ProjectServiceAccountDataSource struct {
	client *OpenAIClient
}

type ProjectServiceAccountDataSourceModel struct {
	ProjectID        types.String `tfsdk:"project_id"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	CreatedAt        types.Int64  `tfsdk:"created_at"`
	Role             types.String `tfsdk:"role"`
	APIKeyID         types.String `tfsdk:"api_key_id"`
}

func (d *ProjectServiceAccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_service_account"
}

func (d *ProjectServiceAccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI project service account.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to which the service account belongs.",
				Required:    true,
			},
			"service_account_id": schema.StringAttribute{
				Description: "The ID of the service account to retrieve.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the resource (composite of project_id:service_account_id).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the service account.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The timestamp (in Unix time) when the service account was created.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The role of the service account.",
				Computed:    true,
			},
			"api_key_id": schema.StringAttribute{
				Description: "The ID of the API key associated with the service account.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectServiceAccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectServiceAccountDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	serviceAccountID := data.ServiceAccountID.ValueString()

	// API call requires Admin Key.
	// We'll use the client's AdminAPIKey which is configured in provider
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key (admin_key) to manage project service accounts.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	// /v1/organization/projects/{project_id}/service_accounts/{service_account_id}
	suffix := fmt.Sprintf("/organization/projects/%s/service_accounts/%s", projectID, serviceAccountID)

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
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
		return
	}

	var saResp ProjectServiceAccountResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&saResp); err != nil {
		resp.Diagnostics.AddError("Error decoding response", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", projectID, saResp.ID))
	data.Name = types.StringValue(saResp.Name)
	data.CreatedAt = types.Int64Value(saResp.CreatedAt)
	data.Role = types.StringValue(saResp.Role)

	if saResp.APIKey != nil {
		data.APIKeyID = types.StringValue(saResp.APIKey.ID)
	} else {
		data.APIKeyID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ProjectServiceAccountsDataSource

func NewProjectServiceAccountsDataSource() datasource.DataSource {
	return &ProjectServiceAccountsDataSource{}
}

type ProjectServiceAccountsDataSource struct {
	client *OpenAIClient
}

type ProjectServiceAccountsDataSourceModel struct {
	ProjectID       types.String                       `tfsdk:"project_id"`
	ServiceAccounts []ProjectServiceAccountResultModel `tfsdk:"service_accounts"`
	ID              types.String                       `tfsdk:"id"`
}

type ProjectServiceAccountResultModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
	Role      types.String `tfsdk:"role"`
	APIKeyID  types.String `tfsdk:"api_key_id"`
}

type ProjectServiceAccountsListResponse struct {
	Object  string                          `json:"object"`
	Data    []ProjectServiceAccountResponse `json:"data"`
	FirstID string                          `json:"first_id"`
	LastID  string                          `json:"last_id"`
	HasMore bool                            `json:"has_more"`
}

func (d *ProjectServiceAccountsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_service_accounts"
}

func (d *ProjectServiceAccountsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of service accounts for a specific OpenAI project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project from which to retrieve service accounts.",
				Required:    true,
			},
			"service_accounts": schema.ListNestedAttribute{
				Description: "List of service accounts in the project.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the service account.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the service account.",
							Computed:    true,
						},
						"created_at": schema.Int64Attribute{
							Description: "The timestamp (in Unix time) when the service account was created.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "The role of the service account.",
							Computed:    true,
						},
						"api_key_id": schema.StringAttribute{
							Description: "The ID of the API key associated with the service account.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ProjectServiceAccountsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectServiceAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectServiceAccountsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key (admin_key) to list project service accounts.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	// /v1/organization/projects/{project_id}/service_accounts
	suffix := fmt.Sprintf("/organization/projects/%s/service_accounts", projectID)

	var allAccounts []ProjectServiceAccountResultModel
	cursor := ""

	for {
		// Construct full URL with Base + Suffix
		var reqURL string
		if strings.Contains(apiURL, "/v1") {
			reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
		} else {
			reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
		}

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

		var listResp ProjectServiceAccountsListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}

		for _, sa := range listResp.Data {
			saModel := ProjectServiceAccountResultModel{
				ID:        types.StringValue(sa.ID),
				Name:      types.StringValue(sa.Name),
				CreatedAt: types.Int64Value(sa.CreatedAt),
				Role:      types.StringValue(sa.Role),
			}
			if sa.APIKey != nil {
				saModel.APIKeyID = types.StringValue(sa.APIKey.ID)
			} else {
				saModel.APIKeyID = types.StringNull()
			}
			allAccounts = append(allAccounts, saModel)
		}

		if !listResp.HasMore {
			break
		}
		cursor = listResp.LastID
	}

	data.ID = types.StringValue(projectID)
	data.ServiceAccounts = allAccounts

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
