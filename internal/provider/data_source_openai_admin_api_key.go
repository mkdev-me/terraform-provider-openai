package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AdminAPIKeyDataSource{}

func NewAdminAPIKeyDataSource() datasource.DataSource {
	return &AdminAPIKeyDataSource{}
}

type AdminAPIKeyDataSource struct {
	client *OpenAIClient
}

type AdminAPIKeyDataSourceModel struct {
	APIKeyID  types.String   `tfsdk:"api_key_id"`
	ID        types.String   `tfsdk:"id"`
	Name      types.String   `tfsdk:"name"`
	CreatedAt types.String   `tfsdk:"created_at"`
	ExpiresAt types.Int64    `tfsdk:"expires_at"`
	Scopes    []types.String `tfsdk:"scopes"`
	Object    types.String   `tfsdk:"object"`
}

func (d *AdminAPIKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_api_key"
}

func (d *AdminAPIKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI admin API key.",
		Attributes: map[string]schema.Attribute{
			"api_key_id": schema.StringAttribute{
				Description: "The ID of the admin API key to retrieve.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the admin API key.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the admin API key.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the admin API key was created.",
				Computed:    true,
			},
			"expires_at": schema.Int64Attribute{
				Description: "Timestamp when the admin API key expires (optional).",
				Computed:    true,
			},
			"scopes": schema.ListAttribute{
				Description: "Scopes assigned to the admin API key.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"object": schema.StringAttribute{
				Description: "The object type.",
				Computed:    true,
			},
		},
	}
}

func (d *AdminAPIKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AdminAPIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AdminAPIKeyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKeyID := data.APIKeyID.ValueString()
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API Key is required to read admin API Key details.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	// /v1/organization/admin_api_keys/{key_id}
	suffix := fmt.Sprintf("/organization/admin_api_keys/%s", apiKeyID)

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

	var apiKey AdminAPIKeyResponseFramework
	if err := json.NewDecoder(httpResp.Body).Decode(&apiKey); err != nil {
		resp.Diagnostics.AddError("Error decoding response", err.Error())
		return
	}

	data.ID = types.StringValue(apiKey.ID)
	data.Name = types.StringValue(apiKey.Name)
	data.Object = types.StringValue(apiKey.Object)
	data.CreatedAt = types.StringValue(time.Unix(apiKey.CreatedAt, 0).Format(time.RFC3339))

	if apiKey.ExpiresAt != nil {
		data.ExpiresAt = types.Int64Value(*apiKey.ExpiresAt)
	} else {
		data.ExpiresAt = types.Int64Null()
	}

	if apiKey.Scopes != nil {
		scopes := make([]types.String, len(apiKey.Scopes))
		for i, v := range apiKey.Scopes {
			scopes[i] = types.StringValue(v)
		}
		data.Scopes = scopes
	} else {
		data.Scopes = []types.String{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// AdminAPIKeysDataSource

func NewAdminAPIKeysDataSource() datasource.DataSource {
	return &AdminAPIKeysDataSource{}
}

type AdminAPIKeysDataSource struct {
	client *OpenAIClient
}

type AdminAPIKeysDataSourceModel struct {
	Limit   types.Int64              `tfsdk:"limit"`
	After   types.String             `tfsdk:"after"`
	APIKeys []AdminAPIKeyResultModel `tfsdk:"api_keys"`
	HasMore types.Bool               `tfsdk:"has_more"`
	ID      types.String             `tfsdk:"id"`
}

type AdminAPIKeyResultModel struct {
	ID         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	Object     types.String   `tfsdk:"object"`
	CreatedAt  types.String   `tfsdk:"created_at"`
	ExpiresAt  types.Int64    `tfsdk:"expires_at"`
	LastUsedAt types.String   `tfsdk:"last_used_at"`
	Scopes     []types.String `tfsdk:"scopes"`
}

func (d *AdminAPIKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_api_keys"
}

func (d *AdminAPIKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all admin API keys.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Maximum number of API keys to return. Default is 20.",
				Optional:    true,
				Computed:    true,
			},
			"after": schema.StringAttribute{
				Description: "Cursor for pagination, API key ID to fetch results after.",
				Optional:    true,
			},
			"api_keys": schema.ListNestedAttribute{
				Description: "List of admin API keys.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the admin API key.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the admin API key.",
							Computed:    true,
						},
						"object": schema.StringAttribute{
							Description: "The object type.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the admin API key was created.",
							Computed:    true,
						},
						"expires_at": schema.Int64Attribute{
							Description: "Timestamp when the admin API key expires (optional).",
							Computed:    true,
						},
						"last_used_at": schema.StringAttribute{
							Description: "Timestamp when the admin API key was last used.",
							Computed:    true,
						},
						"scopes": schema.ListAttribute{
							Description: "Scopes assigned to the admin API key.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"has_more": schema.BoolAttribute{
				Description: "Whether there are more API keys available beyond the limit.",
				Computed:    true,
			},
		},
	}
}

func (d *AdminAPIKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AdminAPIKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AdminAPIKeysDataSourceModel

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
	suffix := "/organization/admin_api_keys"

	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1" + suffix
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1" + suffix
	}

	parsedURL, _ := url.Parse(reqURL)
	q := parsedURL.Query()
	limit := int64(20)
	if !data.Limit.IsNull() {
		limit = data.Limit.ValueInt64()
	}
	q.Set("limit", strconv.FormatInt(limit, 10))
	if !data.After.IsNull() {
		q.Set("after", data.After.ValueString())
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

	type AdminAPIKeyListResponse struct {
		Data    []AdminAPIKeyResponseFramework `json:"data"` // Uses same struct but might miss LastUsedAt?
		HasMore bool                           `json:"has_more"`
		Object  string                         `json:"object"`
	}
	// Need a custom struct for LastUsedAt as it's not in AdminAPIKeyResponseFramework
	type AdminAPIKeyWithLastUsed struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		CreatedAt  int64    `json:"created_at"`
		ExpiresAt  *int64   `json:"expires_at,omitempty"`
		LastUsedAt *int64   `json:"last_used_at,omitempty"`
		Object     string   `json:"object"`
		Scopes     []string `json:"scopes,omitempty"`
	}
	type ListResp struct {
		Data    []AdminAPIKeyWithLastUsed `json:"data"`
		HasMore bool                      `json:"has_more"`
		Object  string                    `json:"object"`
	}

	var listResp ListResp
	if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
		resp.Diagnostics.AddError("Error decoding response", err.Error())
		return
	}

	var apiKeys []AdminAPIKeyResultModel
	for _, k := range listResp.Data {
		keyModel := AdminAPIKeyResultModel{
			ID:        types.StringValue(k.ID),
			Name:      types.StringValue(k.Name),
			Object:    types.StringValue(k.Object),
			CreatedAt: types.StringValue(time.Unix(k.CreatedAt, 0).Format(time.RFC3339)),
		}

		if k.ExpiresAt != nil {
			keyModel.ExpiresAt = types.Int64Value(*k.ExpiresAt)
		} else {
			keyModel.ExpiresAt = types.Int64Null()
		}

		if k.LastUsedAt != nil {
			keyModel.LastUsedAt = types.StringValue(time.Unix(*k.LastUsedAt, 0).Format(time.RFC3339))
		} else {
			keyModel.LastUsedAt = types.StringNull()
		}

		if k.Scopes != nil {
			scopes := make([]types.String, len(k.Scopes))
			for i, v := range k.Scopes {
				scopes[i] = types.StringValue(v)
			}
			keyModel.Scopes = scopes
		} else {
			keyModel.Scopes = []types.String{}
		}

		apiKeys = append(apiKeys, keyModel)
	}

	data.ID = types.StringValue(fmt.Sprintf("admin_api_keys_%d", time.Now().Unix()))
	data.APIKeys = apiKeys
	data.HasMore = types.BoolValue(listResp.HasMore)
	data.Limit = types.Int64Value(limit)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// AdminAPIKeyResponseFramework defined in types_project_org.go
