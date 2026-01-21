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

var _ datasource.DataSource = &ThreadDataSource{}

func NewThreadDataSource() datasource.DataSource {
	return &ThreadDataSource{}
}

type ThreadDataSource struct {
	client *OpenAIClient
}

type ThreadDataSourceModel struct {
	ID            types.String                   `tfsdk:"id"`
	Object        types.String                   `tfsdk:"object"`
	CreatedAt     types.Int64                    `tfsdk:"created_at"`
	Metadata      types.Map                      `tfsdk:"metadata"`
	ToolResources *ThreadDataSourceToolResources `tfsdk:"tool_resources"`
}

type ThreadDataSourceToolResources struct {
	CodeInterpreter *ThreadDataSourceCodeInterpreterResources `tfsdk:"code_interpreter"`
	FileSearch      *ThreadDataSourceFileSearchResources      `tfsdk:"file_search"`
}

type ThreadDataSourceCodeInterpreterResources struct {
	FileIDs []types.String `tfsdk:"file_ids"`
}

type ThreadDataSourceFileSearchResources struct {
	VectorStoreIDs []types.String `tfsdk:"vector_store_ids"`
}

func (d *ThreadDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thread"
}

func (d *ThreadDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI thread.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the thread.",
				Required:    true,
			},
			"object": schema.StringAttribute{
				Description: "The object type, which is always 'thread'.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the thread was created.",
				Computed:    true,
			},
			"metadata": schema.MapAttribute{
				Description: "Set of 16 key-value pairs that can be attached to an object.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"tool_resources": schema.SingleNestedAttribute{
				Description: "A set of resources that are used by the thread's tools.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"code_interpreter": schema.SingleNestedAttribute{
						Description: "Resources for the code_interpreter tool.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"file_ids": schema.ListAttribute{
								Description: "A list of file IDs attached to this thread.",
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
								Description: "A list of vector store IDs attached to this thread.",
								ElementType: types.StringType,
								Computed:    true,
							},
						},
					},
				},
			},
		},
	}
}

func (d *ThreadDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ThreadDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ThreadDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	threadID := data.ID.ValueString()
	path := fmt.Sprintf("threads/%s", threadID)

	apiClient := d.client.OpenAIClient

	respBody, err := apiClient.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Thread",
			fmt.Sprintf("Could not read thread with ID %s: %s", threadID, err.Error()),
		)
		return
	}

	var threadResponse ThreadResponse
	if err := json.Unmarshal(respBody, &threadResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Thread Response",
			fmt.Sprintf("Could not parse response for thread %s: %s", threadID, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(threadResponse.ID)
	data.Object = types.StringValue(threadResponse.Object)
	data.CreatedAt = types.Int64Value(int64(threadResponse.CreatedAt))

	// Map Metadata
	if len(threadResponse.Metadata) > 0 {
		metadataVals := make(map[string]attr.Value)
		for k, v := range threadResponse.Metadata {
			metadataVals[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		data.Metadata, _ = types.MapValue(types.StringType, metadataVals)
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	// Map Tool Resources
	if threadResponse.ToolResources != nil {
		trModel := &ThreadDataSourceToolResources{}
		hasResources := false

		if threadResponse.ToolResources.CodeInterpreter != nil {
			fileIDs := []types.String{}
			for _, id := range threadResponse.ToolResources.CodeInterpreter.FileIDs {
				fileIDs = append(fileIDs, types.StringValue(id))
			}
			trModel.CodeInterpreter = &ThreadDataSourceCodeInterpreterResources{
				FileIDs: fileIDs,
			}
			hasResources = true
		}

		if threadResponse.ToolResources.FileSearch != nil {
			vsIDs := []types.String{}
			for _, id := range threadResponse.ToolResources.FileSearch.VectorStoreIDs {
				vsIDs = append(vsIDs, types.StringValue(id))
			}
			trModel.FileSearch = &ThreadDataSourceFileSearchResources{
				VectorStoreIDs: vsIDs,
			}
			hasResources = true
		}

		if hasResources {
			data.ToolResources = trModel
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
