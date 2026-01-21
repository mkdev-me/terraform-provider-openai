package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectsDataSource{}

func NewProjectsDataSource() datasource.DataSource {
	return &ProjectsDataSource{}
}

type ProjectsDataSource struct {
	client *OpenAIClient
}

type ProjectsDataSourceModel struct {
	AdminKey types.String         `tfsdk:"admin_key"`
	Projects []ProjectResultModel `tfsdk:"projects"`
	ID       types.String         `tfsdk:"id"` // Dummy ID
}

type ProjectResultModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

type ProjectsListResponseFramework struct {
	Object  string                     `json:"object"`
	Data    []ProjectResponseFramework `json:"data"`
	FirstID string                     `json:"first_id"`
	LastID  string                     `json:"last_id"`
	HasMore bool                       `json:"has_more"`
}

func (d *ProjectsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *ProjectsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of available OpenAI projects.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"admin_key": schema.StringAttribute{
				Description: "Admin API key for authentication. If not provided, the provider's default Admin API key will be used.",
				Optional:    true,
				Sensitive:   true,
			},
			"projects": schema.ListNestedAttribute{
				Description: "List of available projects.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
				},
			},
		},
	}
}

func (d *ProjectsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := d.client.AdminAPIKey
	if !data.AdminKey.IsNull() {
		apiKey = data.AdminKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API key is required to list projects. Please provide it in the configuration or ensure the provider is configured with one.",
		)
		return
	}

	var allProjects []ProjectResultModel
	cursor := ""
	limit := 100

	for {
		// Construct the API URL
		apiURL := d.client.OpenAIClient.APIURL

		// Ensure leading slash for construction if needed, but strings.TrimPrefix helps
		// Actually best to form: base + path
		var fullURL string
		if strings.Contains(apiURL, "/v1") {
			fullURL = fmt.Sprintf("%s/organization/projects", strings.TrimSuffix(apiURL, "/v1"))
			// Wait, if apiURL ends in /v1, and we want /v1/organization/projects
			// reusing apiURL (e.g. https://api.openai.com/v1) -> https://api.openai.com/v1/organization/projects
			fullURL = fmt.Sprintf("%s/organization/projects", strings.TrimSuffix(apiURL, "/"))
		} else {
			fullURL = fmt.Sprintf("%s/v1/organization/projects", strings.TrimSuffix(apiURL, "/"))
		}

		// Add query params
		parsedURL, err := url.Parse(fullURL)
		if err != nil {
			resp.Diagnostics.AddError("Error parsing URL", err.Error())
			return
		}
		q := parsedURL.Query()
		q.Set("limit", strconv.Itoa(limit))
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		httpRequest, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating request", err.Error())
			return
		}

		httpRequest.Header.Set("Authorization", "Bearer "+apiKey)
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

		var listResp ProjectsListResponseFramework
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}

		for _, p := range listResp.Data {
			projectModel := ProjectResultModel{
				ID:        types.StringValue(p.ID),
				Name:      types.StringValue(p.Name),
				Status:    types.StringValue(p.Status),
				CreatedAt: types.Int64Value(p.CreatedAt),
			}
			allProjects = append(allProjects, projectModel)
		}

		if !listResp.HasMore {
			break
		}
		if listResp.LastID == "" {
			break
		}
		cursor = listResp.LastID
	}

	data.ID = types.StringValue("projects")
	data.Projects = allProjects

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
