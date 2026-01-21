package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AssistantDataSource{}

func NewAssistantDataSource() datasource.DataSource {
	return &AssistantDataSource{}
}

type AssistantDataSource struct {
	client *OpenAIClient
}

type AssistantDataSourceModel struct {
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

type AssistantDataSourceToolResourcesModel struct {
	CodeInterpreter *AssistantDataSourceCodeInterpreterResourcesModel `tfsdk:"code_interpreter"`
	FileSearch      *AssistantDataSourceFileSearchResourcesModel      `tfsdk:"file_search"`
}

type AssistantDataSourceCodeInterpreterResourcesModel struct {
	FileIDs []types.String `tfsdk:"file_ids"`
}

type AssistantDataSourceFileSearchResourcesModel struct {
	VectorStoreIDs []types.String `tfsdk:"vector_store_ids"`
}

func (d *AssistantDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assistant"
}

func (d *AssistantDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI assistant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the assistant.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the assistant.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the assistant.",
				Computed:    true,
			},
			"model": schema.StringAttribute{
				Description: "The ID of the model used by the assistant.",
				Computed:    true,
			},
			"instructions": schema.StringAttribute{
				Description: "The system instructions that the assistant uses.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the assistant was created.",
				Computed:    true,
			},
			"object": schema.StringAttribute{
				Description: "The object type, which is always 'assistant'.",
				Computed:    true,
			},
			"metadata": schema.MapAttribute{
				Description: "Set of 16 key-value pairs that can be attached to an object.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"tools": schema.ListNestedAttribute{
				Description: "A list of tools enabled on the assistant.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The type of tool. Can be 'code_interpreter', 'retrieval', or 'function'.",
							Computed:    true,
						},
						"function": schema.ListNestedAttribute{
							Description: "The function definition, if the tool type is 'function'.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The name of the function.",
										Computed:    true,
									},
									"description": schema.StringAttribute{
										Description: "The description of the function.",
										Computed:    true,
									},
									"parameters": schema.StringAttribute{
										Description: "The parameters of the function in JSON format.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"tool_resources": schema.SingleNestedAttribute{
				Description: "A set of resources that are used by the assistant's tools.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"code_interpreter": schema.SingleNestedAttribute{
						Description: "Resources for the code_interpreter tool.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"file_ids": schema.ListAttribute{
								Description: "A list of file IDs attached to this assistant.",
								ElementType: types.StringType,
								Computed:    true,
							},
						},
					},
					"file_search": schema.SingleNestedAttribute{
						Description: "Resources for the file_search tool.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"vector_store_ids": schema.ListAttribute{
								Description: "A list of vector store IDs attached to this assistant.",
								ElementType: types.StringType,
								Computed:    true,
							},
						},
					},
				},
			},
			"temperature": schema.Float64Attribute{
				Description: "What sampling temperature to use, between 0 and 2.",
				Computed:    true,
			},
			"top_p": schema.Float64Attribute{
				Description: "An alternative to sampling with temperature, called nucleus sampling.",
				Computed:    true,
			},
			"response_format": schema.StringAttribute{
				Description: "The format of the response.",
				Computed:    true,
			},
		},
	}
}

func (d *AssistantDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AssistantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AssistantDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	assistantID := data.ID.ValueString()

	path := fmt.Sprintf("assistants/%s", assistantID)

	apiClient := d.client.OpenAIClient

	respBody, err := apiClient.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Assistant",
			fmt.Sprintf("Could not read assistant with ID %s: %s", assistantID, err.Error()),
		)
		return
	}

	var assistantResponse AssistantResponse
	if err := json.Unmarshal(respBody, &assistantResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Assistant Response",
			fmt.Sprintf("Could not parse response for assistant %s: %s", assistantID, err.Error()),
		)
		return
	}

	// Map response to model
	data.ID = types.StringValue(assistantResponse.ID)
	data.Name = types.StringValue(assistantResponse.Name)
	data.Description = types.StringValue(assistantResponse.Description)
	data.Model = types.StringValue(assistantResponse.Model)
	data.Instructions = types.StringValue(assistantResponse.Instructions)
	data.CreatedAt = types.Int64Value(int64(assistantResponse.CreatedAt))
	data.Object = types.StringValue(assistantResponse.Object)

	// Map Metadata
	if len(assistantResponse.Metadata) > 0 {
		metadataVals := make(map[string]attr.Value)
		for k, v := range assistantResponse.Metadata {
			metadataVals[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		data.Metadata, _ = types.MapValue(types.StringType, metadataVals)
	} else {
		data.Metadata = types.MapNull(types.StringType)
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
		data.Tools = tools
	} else {
		data.Tools = nil
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
			data.ToolResources = trModel
		}
	}

	// Map other fields
	if assistantResponse.Temperature != 0 {
		data.Temperature = types.Float64Value(assistantResponse.Temperature)
	} else {
		data.Temperature = types.Float64Null()
	}

	if assistantResponse.TopP != 0 {
		data.TopP = types.Float64Value(assistantResponse.TopP)
	} else {
		data.TopP = types.Float64Null()
	}

	if assistantResponse.ResponseFormat != nil {
		rfStr := fmt.Sprintf("%v", assistantResponse.ResponseFormat)
		data.ResponseFormat = types.StringValue(rfStr)
	} else {
		data.ResponseFormat = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
