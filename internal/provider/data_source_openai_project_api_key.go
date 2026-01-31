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

var _ datasource.DataSource = &ProjectAPIKeyDataSource{}

func NewProjectAPIKeyDataSource() datasource.DataSource {
	return &ProjectAPIKeyDataSource{}
}

type ProjectAPIKeyDataSource struct {
	client *OpenAIClient
}

type ProjectAPIKeyDataSourceModel struct {
	ProjectID  types.String `tfsdk:"project_id"`
	APIKeyID   types.String `tfsdk:"api_key_id"`
	AdminKey   types.String `tfsdk:"admin_key"`
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	CreatedAt  types.String `tfsdk:"created_at"`
	LastUsedAt types.String `tfsdk:"last_used_at"`
}

type ProjectAPIKeyDataSourceResponse struct {
	Object     string `json:"object"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreatedAt  int64  `json:"created_at"`
	LastUsedAt int64  `json:"last_used_at,omitempty"`
}

func (d *ProjectAPIKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_api_key"
}

func (d *ProjectAPIKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI project API key.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to which the API key belongs.",
				Required:    true,
			},
			"api_key_id": schema.StringAttribute{
				Description: "The ID of the API key to retrieve.",
				Required:    true,
			},
			"admin_key": schema.StringAttribute{
				Description: "Admin API key for authentication. If not provided, the provider's default Admin API key will be used.",
				Optional:    true,
				Sensitive:   true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the API key resource (composite of project_id:api_key_id).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the API key.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the API key was created.",
				Computed:    true,
			},
			"last_used_at": schema.StringAttribute{
				Description: "Timestamp when the API key was last used.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectAPIKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectAPIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectAPIKeyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	apiKeyID := data.APIKeyID.ValueString()

	// Determine which API key to use
	adminKey := d.client.AdminAPIKey
	if !data.AdminKey.IsNull() {
		adminKey = data.AdminKey.ValueString()
	}

	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API key is required to read project API key details. Please provide it in the configuration or ensure the provider is configured with one.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	// Construct path: /v1/organization/projects/{project_id}/api_keys/{key_id}
	// Similar logic to projects ds
	suffix := fmt.Sprintf("/organization/projects/%s/api_keys/%s", projectID, apiKeyID)

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

	var apiKeyResp ProjectAPIKeyDataSourceResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiKeyResp); err != nil {
		resp.Diagnostics.AddError("Error decoding response", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", projectID, apiKeyID))
	data.Name = types.StringValue(apiKeyResp.Name)

	// Format timestamps
	data.CreatedAt = types.StringValue(time.Unix(apiKeyResp.CreatedAt, 0).Format(time.RFC3339))
	if apiKeyResp.LastUsedAt > 0 {
		data.LastUsedAt = types.StringValue(time.Unix(apiKeyResp.LastUsedAt, 0).Format(time.RFC3339))
	} else {
		data.LastUsedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewProjectAPIKeysDataSource defines the data source implementation.
func NewProjectAPIKeysDataSource() datasource.DataSource {
	return &ProjectAPIKeysDataSource{}
}

type ProjectAPIKeysDataSource struct {
	client *OpenAIClient
}

type ProjectAPIKeysDataSourceModel struct {
	ProjectID types.String               `tfsdk:"project_id"`
	AdminKey  types.String               `tfsdk:"admin_key"`
	APIKeys   []ProjectAPIKeyResultModel `tfsdk:"api_keys"`
	ID        types.String               `tfsdk:"id"`
}

type ProjectAPIKeyResultModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	CreatedAt  types.String `tfsdk:"created_at"`
	LastUsedAt types.String `tfsdk:"last_used_at"`
}

type ProjectAPIKeysListResponse struct {
	Object string                            `json:"object"`
	Data   []ProjectAPIKeyDataSourceResponse `json:"data"`
}

func (d *ProjectAPIKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_api_keys"
}

func (d *ProjectAPIKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of API keys for a specific OpenAI project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to retrieve API keys for.",
				Required:    true,
			},
			"admin_key": schema.StringAttribute{
				Description: "Admin API key for authentication. If not provided, the provider's default Admin API key will be used.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_keys": schema.ListNestedAttribute{
				Description: "List of API keys for the project.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the API key.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the API key.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the API key was created.",
							Computed:    true,
						},
						"last_used_at": schema.StringAttribute{
							Description: "Timestamp when the API key was last used.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ProjectAPIKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectAPIKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectAPIKeysDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	adminKey := d.client.AdminAPIKey
	if !data.AdminKey.IsNull() {
		adminKey = data.AdminKey.ValueString()
	}

	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API key is required. Please provide it in the configuration or ensure the provider is configured with one.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	suffix := fmt.Sprintf("/organization/projects/%s/api_keys", projectID)

	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
	}

	parsedURL, _ := url.Parse(reqURL)
	q := parsedURL.Query()
	q.Set("limit", "100")
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

	var listResp ProjectAPIKeysListResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
		resp.Diagnostics.AddError("Error decoding response", err.Error())
		return
	}

	var allKeys []ProjectAPIKeyResultModel
	for _, k := range listResp.Data {
		keyModel := ProjectAPIKeyResultModel{
			ID:        types.StringValue(k.ID),
			Name:      types.StringValue(k.Name),
			CreatedAt: types.StringValue(time.Unix(k.CreatedAt, 0).Format(time.RFC3339)),
		}
		if k.LastUsedAt > 0 {
			keyModel.LastUsedAt = types.StringValue(time.Unix(k.LastUsedAt, 0).Format(time.RFC3339))
		} else {
			keyModel.LastUsedAt = types.StringNull()
		}
		allKeys = append(allKeys, keyModel)
	}

	data.ID = types.StringValue(projectID)
	data.APIKeys = allKeys

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
