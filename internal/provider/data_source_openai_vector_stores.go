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

var _ datasource.DataSource = &VectorStoresDataSource{}

func NewVectorStoresDataSource() datasource.DataSource {
	return &VectorStoresDataSource{}
}

type VectorStoresDataSource struct {
	client *OpenAIClient
}

type VectorStoresDataSourceModel struct {
	Result []VectorStoreResultModel `tfsdk:"vector_stores"`
	ID     types.String             `tfsdk:"id"` // Dummy ID
}

// VectorStoreResultModel mirrors VectorStoreDataSourceModel but for use in a list
type VectorStoreResultModel struct {
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

func (d *VectorStoresDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_stores"
}

func (d *VectorStoresDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {

	vsNestedAttributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"object": schema.StringAttribute{
			Computed: true,
		},
		"created_at": schema.Int64Attribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"usage_bytes": schema.Int64Attribute{
			Computed: true,
		},
		"status": schema.StringAttribute{
			Computed: true,
		},
		"file_counts": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"in_progress": schema.Int64Attribute{
					Computed: true,
				},
				"completed": schema.Int64Attribute{
					Computed: true,
				},
				"failed": schema.Int64Attribute{
					Computed: true,
				},
				"cancelled": schema.Int64Attribute{
					Computed: true,
				},
				"total": schema.Int64Attribute{
					Computed: true,
				},
			},
		},
		"expires_after": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"anchor": schema.StringAttribute{
					Computed: true,
				},
				"days": schema.Int64Attribute{
					Computed: true,
				},
			},
		},
		"expires_at": schema.Int64Attribute{
			Computed: true,
		},
		"last_active_at": schema.Int64Attribute{
			Computed: true,
		},
		"metadata": schema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of OpenAI vector stores.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"vector_stores": schema.ListNestedAttribute{
				Description: "List of vector stores.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: vsNestedAttributes,
				},
			},
		},
	}
}

func (d *VectorStoresDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VectorStoresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VectorStoresDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := d.client.OpenAIClient
	var allVectorStores []VectorStoreResultModel
	cursor := ""

	for {
		queryParams := url.Values{}
		queryParams.Set("limit", "100")
		if cursor != "" {
			queryParams.Set("after", cursor)
		}

		path := fmt.Sprintf("vector_stores?%s", queryParams.Encode())

		respBody, err := apiClient.DoRequest(http.MethodGet, path, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Vector Stores",
				fmt.Sprintf("Could not list vector stores: %s", err.Error()),
			)
			return
		}

		var listResp ListVectorStoresResponse
		if err := json.Unmarshal(respBody, &listResp); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Vector Stores Response",
				fmt.Sprintf("Could not parse response: %s", err.Error()),
			)
			return
		}

		for _, vsResponse := range listResp.Data {
			vsModel := VectorStoreResultModel{
				ID:         types.StringValue(vsResponse.ID),
				Object:     types.StringValue(vsResponse.Object),
				CreatedAt:  types.Int64Value(vsResponse.CreatedAt),
				Name:       types.StringValue(vsResponse.Name),
				UsageBytes: types.Int64Value(vsResponse.UsageBytes),
				Status:     types.StringValue(vsResponse.Status),
			}

			if vsResponse.ExpiresAt != nil {
				vsModel.ExpiresAt = types.Int64Value(*vsResponse.ExpiresAt)
			} else {
				vsModel.ExpiresAt = types.Int64Null()
			}

			if vsResponse.LastActiveAt != nil {
				vsModel.LastActiveAt = types.Int64Value(*vsResponse.LastActiveAt)
			} else {
				vsModel.LastActiveAt = types.Int64Null()
			}

			// Map File Counts
			if vsResponse.FileCounts != nil {
				vsModel.FileCounts = &VectorStoreFileCountsModel{
					InProgress: types.Int64Value(int64(vsResponse.FileCounts.InProgress)),
					Completed:  types.Int64Value(int64(vsResponse.FileCounts.Completed)),
					Failed:     types.Int64Value(int64(vsResponse.FileCounts.Failed)),
					Cancelled:  types.Int64Value(int64(vsResponse.FileCounts.Cancelled)),
					Total:      types.Int64Value(int64(vsResponse.FileCounts.Total)),
				}
			} else {
				vsModel.FileCounts = nil
			}

			// Map Expires After
			if vsResponse.ExpiresAfter != nil {
				vsModel.ExpiresAfter = &VectorStoreExpiresAfterModel{
					Anchor: types.StringValue(vsResponse.ExpiresAfter.Anchor),
					Days:   types.Int64Value(int64(vsResponse.ExpiresAfter.Days)),
				}
			} else {
				vsModel.ExpiresAfter = nil
			}

			// Map Metadata
			if len(vsResponse.Metadata) > 0 {
				metadataVals := make(map[string]attr.Value)
				for k, v := range vsResponse.Metadata {
					metadataVals[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
				vsModel.Metadata, _ = types.MapValue(types.StringType, metadataVals)
			} else {
				vsModel.Metadata = types.MapNull(types.StringType)
			}

			allVectorStores = append(allVectorStores, vsModel)
		}

		if !listResp.HasMore {
			break
		}
		cursor = listResp.LastID
	}

	data.ID = types.StringValue("vector_stores") // Use static ID or based on params (none here really)
	data.Result = allVectorStores

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
