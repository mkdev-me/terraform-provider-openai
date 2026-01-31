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

var _ datasource.DataSource = &VectorStoreDataSource{}

func NewVectorStoreDataSource() datasource.DataSource {
	return &VectorStoreDataSource{}
}

type VectorStoreDataSource struct {
	client *OpenAIClient
}

type VectorStoreDataSourceModel struct {
	ID           types.String                  `tfsdk:"id"`
	Object       types.String                  `tfsdk:"object"`
	CreatedAt    types.Int64                   `tfsdk:"created_at"`
	Name         types.String                  `tfsdk:"name"`
	UsageBytes   types.Int64                   `tfsdk:"usage_bytes"`
	Status       types.String                  `tfsdk:"status"`
	FileCounts   *VectorStoreFileCountsModel   `tfsdk:"file_counts"`
	ExpiresAfter *VectorStoreExpiresAfterModel `tfsdk:"expires_after"`
	ExpiresAt    types.Int64                   `tfsdk:"expires_at"`
	LastActiveAt types.Int64                   `tfsdk:"last_active_at"`
	Metadata     types.Map                     `tfsdk:"metadata"`
}

type VectorStoreFileCountsModel struct {
	InProgress types.Int64 `tfsdk:"in_progress"`
	Completed  types.Int64 `tfsdk:"completed"`
	Failed     types.Int64 `tfsdk:"failed"`
	Cancelled  types.Int64 `tfsdk:"cancelled"`
	Total      types.Int64 `tfsdk:"total"`
}

type VectorStoreExpiresAfterModel struct {
	Anchor types.String `tfsdk:"anchor"`
	Days   types.Int64  `tfsdk:"days"`
}

func (d *VectorStoreDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_store"
}

func (d *VectorStoreDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI vector store.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the vector store.",
				Required:    true,
			},
			"object": schema.StringAttribute{
				Description: "The object type, which is always 'vector_store'.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the vector store was created.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the vector store.",
				Computed:    true,
			},
			"usage_bytes": schema.Int64Attribute{
				Description: "The total number of bytes used by the files in the vector store.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the vector store, which can be 'expired', 'in_progress', or 'completed'.",
				Computed:    true,
			},
			"file_counts": schema.SingleNestedAttribute{
				Description: "Counts of files in various states within the vector store.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"in_progress": schema.Int64Attribute{
						Description: "Number of files in progress.",
						Computed:    true,
					},
					"completed": schema.Int64Attribute{
						Description: "Number of files completed.",
						Computed:    true,
					},
					"failed": schema.Int64Attribute{
						Description: "Number of files failed.",
						Computed:    true,
					},
					"cancelled": schema.Int64Attribute{
						Description: "Number of files cancelled.",
						Computed:    true,
					},
					"total": schema.Int64Attribute{
						Description: "Total number of files.",
						Computed:    true,
					},
				},
			},
			"expires_after": schema.SingleNestedAttribute{
				Description: "The expiration policy for a vector store.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"anchor": schema.StringAttribute{
						Description: "Anchor timestamp type for the expiration policy.",
						Computed:    true,
					},
					"days": schema.Int64Attribute{
						Description: "The number of days after the anchor time that the vector store will expire.",
						Computed:    true,
					},
				},
			},
			"expires_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the vector store will expire.",
				Computed:    true,
			},
			"last_active_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the vector store was last active.",
				Computed:    true,
			},
			"metadata": schema.MapAttribute{
				Description: "Set of key-value pairs that can be attached to an object.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *VectorStoreDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VectorStoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VectorStoreDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vectorStoreID := data.ID.ValueString()
	path := fmt.Sprintf("vector_stores/%s", vectorStoreID)

	apiClient := d.client.OpenAIClient

	respBody, err := apiClient.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Vector Store",
			fmt.Sprintf("Could not read vector store with ID %s: %s", vectorStoreID, err.Error()),
		)
		return
	}

	var vsResponse VectorStoreResponse
	if err := json.Unmarshal(respBody, &vsResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Vector Store Response",
			fmt.Sprintf("Could not parse response for vector store %s: %s", vectorStoreID, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(vsResponse.ID)
	data.Object = types.StringValue(vsResponse.Object)
	data.CreatedAt = types.Int64Value(vsResponse.CreatedAt)
	data.Name = types.StringValue(vsResponse.Name)
	data.UsageBytes = types.Int64Value(vsResponse.UsageBytes)
	data.Status = types.StringValue(vsResponse.Status)

	if vsResponse.ExpiresAt != nil {
		data.ExpiresAt = types.Int64Value(*vsResponse.ExpiresAt)
	} else {
		data.ExpiresAt = types.Int64Null()
	}

	if vsResponse.LastActiveAt != nil {
		data.LastActiveAt = types.Int64Value(*vsResponse.LastActiveAt)
	} else {
		data.LastActiveAt = types.Int64Null()
	}

	// Map File Counts
	if vsResponse.FileCounts != nil {
		data.FileCounts = &VectorStoreFileCountsModel{
			InProgress: types.Int64Value(int64(vsResponse.FileCounts.InProgress)),
			Completed:  types.Int64Value(int64(vsResponse.FileCounts.Completed)),
			Failed:     types.Int64Value(int64(vsResponse.FileCounts.Failed)),
			Cancelled:  types.Int64Value(int64(vsResponse.FileCounts.Cancelled)),
			Total:      types.Int64Value(int64(vsResponse.FileCounts.Total)),
		}
	} else {
		data.FileCounts = nil
	}

	// Map Expires After
	if vsResponse.ExpiresAfter != nil {
		data.ExpiresAfter = &VectorStoreExpiresAfterModel{
			Anchor: types.StringValue(vsResponse.ExpiresAfter.Anchor),
			Days:   types.Int64Value(int64(vsResponse.ExpiresAfter.Days)),
		}
	} else {
		data.ExpiresAfter = nil
	}

	// Map Metadata
	if len(vsResponse.Metadata) > 0 {
		metadataVals := make(map[string]attr.Value)
		for k, v := range vsResponse.Metadata {
			metadataVals[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		data.Metadata, _ = types.MapValue(types.StringType, metadataVals)
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
