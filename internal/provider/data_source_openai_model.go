package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

// Ensure implementation satisfies interfaces.
var _ datasource.DataSource = &ModelDataSource{}

// ModelDataSource defines the data source implementation.
type ModelDataSource struct {
	client *OpenAIClient
}

// ModelDataSourceModel describes the data source data model.
type ModelDataSourceModel struct {
	ModelID types.String `tfsdk:"model_id"`
	ID      types.String `tfsdk:"id"`
	Created types.Int64  `tfsdk:"created"`
	OwnedBy types.String `tfsdk:"owned_by"`
	Object  types.String `tfsdk:"object"`
}

func NewModelDataSource() datasource.DataSource {
	return &ModelDataSource{}
}

func (d *ModelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (d *ModelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The model data source allows you to retrieve information about a specific OpenAI model.",

		Attributes: map[string]schema.Attribute{
			"model_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the model to retrieve information for.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the model.",
				Computed:            true,
			},
			"created": schema.Int64Attribute{
				MarkdownDescription: "The timestamp for when the model was created.",
				Computed:            true,
			},
			"owned_by": schema.StringAttribute{
				MarkdownDescription: "The organization that owns the model.",
				Computed:            true,
			},
			"object": schema.StringAttribute{
				MarkdownDescription: "The object type, which is always 'model'.",
				Computed:            true,
			},
		},
	}
}

func (d *ModelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*OpenAIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *provider.OpenAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ModelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ModelDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Logic to get the client with project key if needed.
	// In the SDKv2 implementation, it used GetOpenAIClientWithProjectKey.
	// Here we need to replicate that logic or just use d.client if it's already configured correctly?
	// The provider configuration injects the client.
	// SDKv2 logic:
	/*
		func GetOpenAIClientWithProjectKey(m interface{}) (*client.OpenAIClient, error) {
			if c, ok := m.(*OpenAIClient); ok {
				if c.ProjectAPIKey != "" {
					// create new client with project key
				}
				return c.OpenAIClient, nil
			}
			return client.NewClient("", "", ""), nil
		}
	*/

	// We can invoke a helper or do it inline.
	// Since OpenAiClient is available in d.client, we can check ProjectAPIKey.
	apiClient := d.client.OpenAIClient
	if d.client.ProjectAPIKey != "" {
		config := client.ClientConfig{
			APIKey:         d.client.ProjectAPIKey,
			OrganizationID: d.client.OpenAIClient.OrganizationID,
			APIURL:         d.client.OpenAIClient.APIURL,
			Timeout:        d.client.OpenAIClient.Timeout,
		}
		apiClient = client.NewClientWithConfig(config)
	}

	modelID := data.ModelID.ValueString()
	url := fmt.Sprintf("models/%s", modelID)

	respBody, reqErr := apiClient.DoRequest(http.MethodGet, url, nil)
	if reqErr != nil {
		resp.Diagnostics.AddError("Error retrieving model", fmt.Sprintf("Could not retrieve model %s: %s", modelID, reqErr))
		return
	}

	var model struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}
	if err := json.Unmarshal(respBody, &model); err != nil {
		resp.Diagnostics.AddError("Error parsing model response", fmt.Sprintf("Could not parse response: %s", err))
		return
	}

	// Set state
	data.ID = types.StringValue(model.ID)
	data.Created = types.Int64Value(model.Created)
	data.OwnedBy = types.StringValue(model.OwnedBy)
	data.Object = types.StringValue(model.Object)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
