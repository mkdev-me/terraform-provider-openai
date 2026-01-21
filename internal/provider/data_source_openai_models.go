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

var _ datasource.DataSource = &ModelsDataSource{}

func NewModelsDataSource() datasource.DataSource {
	return &ModelsDataSource{}
}

type ModelsDataSource struct {
	client *OpenAIClient
}

func (d *ModelsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_models"
}

type ModelsDataSourceModel struct {
	ID     types.String         `tfsdk:"id"`
	Models []ModelResponseModel `tfsdk:"models"`
}

type ModelResponseModel struct {
	ID      types.String `tfsdk:"id"`
	Created types.Int64  `tfsdk:"created"`
	OwnedBy types.String `tfsdk:"owned_by"`
	Object  types.String `tfsdk:"object"`
}

func (d *ModelsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Models data source allows you to list all available models.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"models": schema.ListNestedAttribute{
				Description: "List of models.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the model.",
							Computed:    true,
						},
						"created": schema.Int64Attribute{
							Description: "The timestamp when the model was created.",
							Computed:    true,
						},
						"owned_by": schema.StringAttribute{
							Description: "The organization that owns the model.",
							Computed:    true,
						},
						"object": schema.StringAttribute{
							Description: "The object type.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ModelsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *provider.OpenAIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ModelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ModelsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	respBody, err := apiClient.DoRequest(http.MethodGet, "models", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing models", err.Error())
		return
	}

	var listResp struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		resp.Diagnostics.AddError("Error parsing models response", err.Error())
		return
	}

	models := []ModelResponseModel{}
	for _, m := range listResp.Data {
		models = append(models, ModelResponseModel{
			ID:      types.StringValue(m.ID),
			Created: types.Int64Value(m.Created),
			OwnedBy: types.StringValue(m.OwnedBy),
			Object:  types.StringValue(m.Object),
		})
	}

	data.Models = models
	data.ID = types.StringValue("models")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
