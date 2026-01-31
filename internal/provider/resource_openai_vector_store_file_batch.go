package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &VectorStoreFileBatchResource{}
var _ resource.ResourceWithImportState = &VectorStoreFileBatchResource{}

type VectorStoreFileBatchResource struct {
	client *OpenAIClient
}

func NewVectorStoreFileBatchResource() resource.Resource {
	return &VectorStoreFileBatchResource{}
}

func (r *VectorStoreFileBatchResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_store_file_batch"
}

type VectorStoreFileBatchResourceModel struct {
	ID               types.String             `tfsdk:"id"`
	VectorStoreID    types.String             `tfsdk:"vector_store_id"`
	FileIDs          []types.String           `tfsdk:"file_ids"`
	ChunkingStrategy *VSChunkingStrategyModel `tfsdk:"chunking_strategy"` // Reusing from vector store

	// Computed
	Object     types.String       `tfsdk:"object"`
	Status     types.String       `tfsdk:"status"`
	CreatedAt  types.Int64        `tfsdk:"created_at"`
	FileCounts *VSFileCountsModel `tfsdk:"file_counts"`
}

func (r *VectorStoreFileBatchResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a file batch in an OpenAI Vector Store.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the vector store file batch.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vector_store_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the vector store to add the batch to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_ids": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "A list of file IDs to add to the vector store.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},

			// Computed
			"object":     schema.StringAttribute{Computed: true},
			"status":     schema.StringAttribute{Computed: true},
			"created_at": schema.Int64Attribute{Computed: true},
			"file_counts": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"in_progress": schema.Int64Attribute{Computed: true},
					"completed":   schema.Int64Attribute{Computed: true},
					"failed":      schema.Int64Attribute{Computed: true},
					"cancelled":   schema.Int64Attribute{Computed: true},
					"total":       schema.Int64Attribute{Computed: true},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"chunking_strategy": schema.SingleNestedBlock{
				MarkdownDescription: "The chunking strategy used to chunk the files.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{Required: true},
					"max_chunk_size_tokens": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "The maximum number of tokens in each chunk. The default is 800. The minimum is 100 and the maximum is 4096.",
					},
					"chunk_overlap_tokens": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "The number of tokens that overlap between chunks. The default is 400. The maximum is half of max_chunk_size_tokens.",
					},
				},
			},
		},
	}
}

func (r *VectorStoreFileBatchResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *provider.OpenAIClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *VectorStoreFileBatchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VectorStoreFileBatchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := VectorStoreFileBatchCreateRequest{}

	if len(data.FileIDs) > 0 {
		ids := make([]string, len(data.FileIDs))
		for i, id := range data.FileIDs {
			ids[i] = id.ValueString()
		}
		createRequest.FileIDs = ids
	}

	if data.ChunkingStrategy != nil {
		cs := &ChunkingStrategy{
			Type: data.ChunkingStrategy.Type.ValueString(),
		}
		if cs.Type == "static" {
			sc := &StaticChunking{}
			if !data.ChunkingStrategy.MaxChunkSizeTokens.IsNull() {
				sc.MaxChunkSizeTokens = int(data.ChunkingStrategy.MaxChunkSizeTokens.ValueInt64())
			}
			if !data.ChunkingStrategy.ChunkOverlapTokens.IsNull() {
				sc.ChunkOverlapTokens = int(data.ChunkingStrategy.ChunkOverlapTokens.ValueInt64())
			}
			cs.Static = sc
		}
		createRequest.ChunkingStrategy = cs
	}

	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing request", err.Error())
		return
	}

	url := fmt.Sprintf("%s/vector_stores/%s/file_batches", r.client.OpenAIClient.APIURL, data.VectorStoreID.ValueString())
	apiReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
	apiReq.Header.Set("OpenAI-Beta", "assistants=v2")
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK && apiResp.StatusCode != http.StatusCreated {
		respBodyBytes, _ := io.ReadAll(apiResp.Body)
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	var vsBatchResp VectorStoreFileBatchResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &vsBatchResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(vsBatchResp.ID)
	data.Object = types.StringValue(vsBatchResp.Object)
	data.CreatedAt = types.Int64Value(vsBatchResp.CreatedAt)
	data.Status = types.StringValue(vsBatchResp.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorStoreFileBatchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VectorStoreFileBatchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/vector_stores/%s/file_batches/%s", r.client.OpenAIClient.APIURL, data.VectorStoreID.ValueString(), data.ID.ValueString())
	apiReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
	apiReq.Header.Set("OpenAI-Beta", "assistants=v2")
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if apiResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s", apiResp.Status))
		return
	}

	var vsBatchResp VectorStoreFileBatchResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &vsBatchResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.Status = types.StringValue(vsBatchResp.Status)
	data.CreatedAt = types.Int64Value(vsBatchResp.CreatedAt)

	if vsBatchResp.FileCounts != nil {
		data.FileCounts = &VSFileCountsModel{
			InProgress: types.Int64Value(int64(vsBatchResp.FileCounts.InProgress)),
			Completed:  types.Int64Value(int64(vsBatchResp.FileCounts.Completed)),
			Failed:     types.Int64Value(int64(vsBatchResp.FileCounts.Failed)),
			Cancelled:  types.Int64Value(int64(vsBatchResp.FileCounts.Cancelled)),
			Total:      types.Int64Value(int64(vsBatchResp.FileCounts.Total)),
		}
	}

	// Note: file_ids might not be returned in the batch object GET response, or they might be.
	// The API ref says the response object has "file_counts" but not "file_ids".
	// So we rely on what's in state for file_ids, as they are immutable.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorStoreFileBatchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Immutable
}

func (r *VectorStoreFileBatchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// API docs: "You can't delete a file batch object."
	// But we can just remove it from state.
	// Legacy provider did nothing (SetId("")).
	// We also do nothing, just let it be removed from state.
}

func (r *VectorStoreFileBatchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected format: vector_store_id:batch_id
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: vector_store_id:batch_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vector_store_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
