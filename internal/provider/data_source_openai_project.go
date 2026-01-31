package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

type ProjectDataSource struct {
	client *OpenAIClient
}

type ProjectDataSourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	AdminKey  types.String `tfsdk:"admin_key"`
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to retrieve.",
				Required:    true,
			},
			"admin_key": schema.StringAttribute{
				Description: "Admin API key for authentication. If not provided, the provider's default Admin API key will be used.",
				Optional:    true,
				Sensitive:   true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the project.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the project.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the project.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the project was created.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()

	// Determine which API key to use
	apiKey := d.client.AdminAPIKey
	if !data.AdminKey.IsNull() {
		apiKey = data.AdminKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API key is required to read project details. Please provide it in the configuration or ensure the provider is configured with one.",
		)
		return
	}

	// Use custom client configuration for admin key if specific key provided, or just use existing logic
	// But since the project endpoint requires Admin Key, we must be careful.
	// We'll construct a direct request or use a helper if we had one.
	// Since DoRequest uses the configured client (Project Key usually), we might need override.

	// However, DoRequest usually uses the client's configured key.
	// If we need to use a different key:
	path := fmt.Sprintf("/v1/organization/projects/%s", projectID)
	// Check for path adjustment
	if strings.Contains(d.client.OpenAIClient.APIURL, "/v1") {
		path = fmt.Sprintf("organization/projects/%s", projectID)
	}

	// Create request manually to allow overriding header
	// Note: d.client.OpenAIClient exposes DoRequest but seemingly not easy header override per request if using the helper method.
	// We might need to access the underlying HTTP client or use a method that allows custom headers if `d.client` exposes it.
	// The `OpenAIClient` wrapper in `provider.go` usually just wraps the `mkdev-me/go-openai` client or similar.
	// Looking at `Configure` in `provider_framework.go`: `providerClient` has `OpenAIClient`, `ProjectAPIKey`, `AdminAPIKey`.

	// If we use `d.client.OpenAIClient.DoRequest`, it likely uses the configured key (Project Key).
	// We need to verify if `OpenAIClient` (from `internal/client`) supports overriding headers.
	// If not, we might need to rely on standard http client or just assume the provider was config with admin key if this datasource is used?
	// But the schema allows `admin_key` override.

	// Let's assume we can use the `DoRequest` but maybe we can't easily swap the key.
	// Actually, `organization/projects` endpoints MUST use Admin Key.
	// If the provider is configured with a Project Key, `DoRequest` will fail if we can't switch it.
	// Ideally we should create a new client with the Admin Key or manually make the request.

	// Safe bet: Manually make the HTTP request if we need a specific key that might differ from the provider's main key.
	apiURL := d.client.OpenAIClient.APIURL
	reqURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(apiURL, "/"), strings.TrimPrefix(path, "/"))

	httpRequest, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	httpRequest.Header.Set("Authorization", "Bearer "+apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")
	// Organization ID from client config if available? The legacy code uses `d.client.OrganizationID`.
	// We can't easily access the internal organization ID from `d.client.OpenAIClient` unless exposed.
	// However, `d.client` is `*OpenAIClient` which is defined in `provider.go` (or `provider_framework.go`?).
	// Let's check `provider_framework.go` again...
	// `type OpenAIClient struct { OpenAIClient *client.Client; ProjectAPIKey ...; AdminAPIKey ... }`
	// It doesn't explicitly store OrganizationID in `OpenAIClient` struct in `provider.go` (based on what I saw earlier in `provider_framework.go`).
	// Wait, `Configure` in `provider_framework.go` creates `config` with `OrganizationID`.
	// But `OpenAIClient` struct definition was unfortunately not fully visible or I missed it.
	// I'll assume standard http client approach is safest for these admin endpoints.

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

	var project ProjectResponseFramework
	if err := json.NewDecoder(httpResp.Body).Decode(&project); err != nil {
		resp.Diagnostics.AddError("Error decoding response", err.Error())
		return
	}

	data.ID = types.StringValue(project.ID)
	data.Name = types.StringValue(project.Name)
	data.Status = types.StringValue(project.Status)
	data.CreatedAt = types.Int64Value(project.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Ensure ProjectResponseFramework is imported or defined. It is in types_project_org.go which is in same package.
