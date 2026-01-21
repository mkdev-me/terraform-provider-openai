package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AssistantsDataSource{}

func NewAssistantsDataSource() datasource.DataSource {
	return &AssistantsDataSource{}
}

type AssistantsDataSource struct {
	client *OpenAIClient
}

type AssistantsDataSourceModel struct {
	ID         types.String           `tfsdk:"id"`
	Assistants []AssistantResultModel `tfsdk:"assistants"`
}

// assistantResultModel mirrors AssistantDataSourceModel but for use in a list
type AssistantResultModel struct {
	ID             types.String                           `tfsdk:"id"`
	Name           types.String                           `tfsdk:"name"`
	Description    types.String                           `tfsdk:"description"`
	Model          types.String                           `tfsdk:"model"`
	Instructions   types.String                           `tfsdk:"instructions"`
	CreatedAt      types.Int64                            `tfsdk:"created_at"`
	Object         types.String                           `tfsdk:"object"`
	Metadata       types.Map                              `tfsdk:"metadata"`
	Tools          []AssistantToolModel                   `tfsdk:"tools"`
	ToolResources  *AssistantDataSourceToolResourcesModel `tfsdk:"tool_resources"`
	Temperature    types.Float64                          `tfsdk:"temperature"`
	TopP           types.Float64                          `tfsdk:"top_p"`
	ResponseFormat types.String                           `tfsdk:"response_format"`
}

func (d *AssistantsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assistants"
}

func (d *AssistantsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Reusing attributes from AssistantDataSource for the nested list
	// We need to define the nested schema explicitly.

	assistantNestedAttributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"description": schema.StringAttribute{
			Computed: true,
		},
		"model": schema.StringAttribute{
			Computed: true,
		},
		"instructions": schema.StringAttribute{
			Computed: true,
		},
		"created_at": schema.Int64Attribute{
			Computed: true,
		},
		"object": schema.StringAttribute{
			Computed: true,
		},
		"metadata": schema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"tools": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed: true,
					},
					"function": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed: true,
								},
								"description": schema.StringAttribute{
									Computed: true,
								},
								"parameters": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
		"tool_resources": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"code_interpreter": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"file_ids": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
				"file_search": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"vector_store_ids": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
		},
		"temperature": schema.Float64Attribute{
			Computed: true,
		},
		"top_p": schema.Float64Attribute{
			Computed: true,
		},
		"response_format": schema.StringAttribute{
			Computed: true,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of OpenAI assistants.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"assistants": schema.ListNestedAttribute{
				Description: "List of assistants.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: assistantNestedAttributes,
				},
			},
		},
	}
}

func (d *AssistantsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AssistantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AssistantsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := d.client.OpenAIClient
	var allAssistants []AssistantResultModel
	cursor := ""

	for {
		queryParams := url.Values{}
		queryParams.Set("limit", "100")
		if cursor != "" {
			queryParams.Set("after", cursor)
		}

		path := fmt.Sprintf("assistants?%s", queryParams.Encode())

		respBody, err := apiClient.DoRequest(http.MethodGet, path, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Assistants",
				fmt.Sprintf("Could not list assistants: %s", err.Error()),
			)
			return
		}

		var listResp ListAssistantsResponse
		if err := json.Unmarshal(respBody, &listResp); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Assistants Response",
				fmt.Sprintf("Could not parse response: %s", err.Error()),
			)
			return
		}

		for _, assistantResponse := range listResp.Data {
			assistantModel := AssistantResultModel{
				ID:           types.StringValue(assistantResponse.ID),
				Name:         types.StringValue(assistantResponse.Name),
				Description:  types.StringValue(assistantResponse.Description),
				Model:        types.StringValue(assistantResponse.Model),
				Instructions: types.StringValue(assistantResponse.Instructions),
				CreatedAt:    types.Int64Value(int64(assistantResponse.CreatedAt)),
				Object:       types.StringValue(assistantResponse.Object),
			}

			// Map Metadata
			if len(assistantResponse.Metadata) > 0 {
				metadataVals := make(map[string]attr.Value)
				for k, v := range assistantResponse.Metadata {
					metadataVals[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
				assistantModel.Metadata, _ = types.MapValue(types.StringType, metadataVals)
			} else {
				assistantModel.Metadata = types.MapNull(types.StringType)
			}

			// Map Tools
			if len(assistantResponse.Tools) > 0 {
				var tools []AssistantToolModel
				for _, t := range assistantResponse.Tools {
					toolModel := AssistantToolModel{
						Type: types.StringValue(t.Type),
					}
					if t.Function != nil {
						paramStr := string(t.Function.Parameters)
						toolModel.Function = []AssistantToolFunctionModel{{
							Name:        types.StringValue(t.Function.Name),
							Description: types.StringValue(t.Function.Description),
							Parameters:  types.StringValue(paramStr),
						}}
					}
					tools = append(tools, toolModel)
				}
				assistantModel.Tools = tools
			} else {
				assistantModel.Tools = nil
			}

			// Map Tool Resources
			if assistantResponse.ToolResources != nil {
				trModel := &AssistantDataSourceToolResourcesModel{}
				hasResources := false

				if assistantResponse.ToolResources.CodeInterpreter != nil {
					fileIDs := []types.String{}
					for _, id := range assistantResponse.ToolResources.CodeInterpreter.FileIDs {
						fileIDs = append(fileIDs, types.StringValue(id))
					}
					trModel.CodeInterpreter = &AssistantDataSourceCodeInterpreterResourcesModel{
						FileIDs: fileIDs,
					}
					hasResources = true
				}

				if assistantResponse.ToolResources.FileSearch != nil {
					vsIDs := []types.String{}
					for _, id := range assistantResponse.ToolResources.FileSearch.VectorStoreIDs {
						vsIDs = append(vsIDs, types.StringValue(id))
					}
					trModel.FileSearch = &AssistantDataSourceFileSearchResourcesModel{
						VectorStoreIDs: vsIDs,
					}
					hasResources = true
				}

				if hasResources {
					assistantModel.ToolResources = trModel
				}
			}

			// Map other fields
			if assistantResponse.Temperature != 0 {
				assistantModel.Temperature = types.Float64Value(assistantResponse.Temperature)
			} else {
				assistantModel.Temperature = types.Float64Null()
			}

			if assistantResponse.TopP != 0 {
				assistantModel.TopP = types.Float64Value(assistantResponse.TopP)
			} else {
				assistantModel.TopP = types.Float64Null()
			}

			if assistantResponse.ResponseFormat != nil {
				rfStr := fmt.Sprintf("%v", assistantResponse.ResponseFormat)
				assistantModel.ResponseFormat = types.StringValue(rfStr)
			} else {
				assistantModel.ResponseFormat = types.StringNull()
			}

			allAssistants = append(allAssistants, assistantModel)
		}

		if !listResp.HasMore {
			break
		}
		cursor = listResp.LastID
	}

	data.ID = types.StringValue("assistants")
	data.Assistants = allAssistants

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
