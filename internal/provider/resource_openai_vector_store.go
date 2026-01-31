package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &VectorStoreResource{}
var _ resource.ResourceWithImportState = &VectorStoreResource{}

type VectorStoreResource struct {
	client *OpenAIClient
}

func NewVectorStoreResource() resource.Resource {
	return &VectorStoreResource{}
}

func (r *VectorStoreResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_store"
}

type VectorStoreResourceModel struct {
	ID               types.String             `tfsdk:"id"`
	Name             types.String             `tfsdk:"name"`
	FileIDs          []types.String           `tfsdk:"file_ids"`
	Metadata         types.Map                `tfsdk:"metadata"`
	ExpiresAfter     *VSExpiresAfterModel     `tfsdk:"expires_after"`
	ChunkingStrategy *VSChunkingStrategyModel `tfsdk:"chunking_strategy"`

	// Computed
	Object     types.String       `tfsdk:"object"`
	Status     types.String       `tfsdk:"status"`
	CreatedAt  types.Int64        `tfsdk:"created_at"`
	UsageBytes types.Int64        `tfsdk:"usage_bytes"`
	FileCounts *VSFileCountsModel `tfsdk:"file_counts"`
}

type VSExpiresAfterModel struct {
	Anchor types.String `tfsdk:"anchor"`
	Days   types.Int64  `tfsdk:"days"`
}

type VSChunkingStrategyModel struct {
	Type               types.String `tfsdk:"type"`
	MaxChunkSizeTokens types.Int64  `tfsdk:"max_chunk_size_tokens"`
	ChunkOverlapTokens types.Int64  `tfsdk:"chunk_overlap_tokens"`
}

type VSFileCountsModel struct {
	InProgress types.Int64 `tfsdk:"in_progress"`
	Completed  types.Int64 `tfsdk:"completed"`
	Failed     types.Int64 `tfsdk:"failed"`
	Cancelled  types.Int64 `tfsdk:"cancelled"`
	Total      types.Int64 `tfsdk:"total"`
}

func (r *VectorStoreResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI Vector Store.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the vector store.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name of the vector store.",
			},
			"file_ids": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "A list of file IDs to add to the vector store.",
			},
			"metadata": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Metadata.",
			},
			"expires_after": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"anchor": schema.StringAttribute{Required: true},
					"days":   schema.Int64Attribute{Required: true},
				},
			},
			"chunking_strategy": schema.SingleNestedAttribute{
				Optional: true,
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
			// Computed
			"object":      schema.StringAttribute{Computed: true},
			"status":      schema.StringAttribute{Computed: true},
			"created_at":  schema.Int64Attribute{Computed: true},
			"usage_bytes": schema.Int64Attribute{Computed: true},
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
	}
}

func (r *VectorStoreResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VectorStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VectorStoreResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := VectorStoreCreateRequest{}
	if !data.Name.IsNull() {
		createRequest.Name = data.Name.ValueString()
	}

	if !data.Metadata.IsNull() {
		metadata := make(map[string]interface{})
		var metaMap map[string]string
		data.Metadata.ElementsAs(ctx, &metaMap, false)
		for k, v := range metaMap {
			metadata[k] = v
		}
		createRequest.Metadata = metadata
	}

	if len(data.FileIDs) > 0 {
		ids := make([]string, len(data.FileIDs))
		for i, id := range data.FileIDs {
			ids[i] = id.ValueString()
		}
		createRequest.FileIDs = ids
	}

	if data.ExpiresAfter != nil {
		createRequest.ExpiresAfter = &ExpiresAfter{
			Anchor: data.ExpiresAfter.Anchor.ValueString(),
			Days:   int(data.ExpiresAfter.Days.ValueInt64()),
		}
	}

	if data.ChunkingStrategy != nil {
		cs := &ChunkingStrategy{
			Type: data.ChunkingStrategy.Type.ValueString(),
		}
		if cs.Type == "static" {
			// Note: API wrapper structure for static chunking
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

	url := fmt.Sprintf("%s/vector_stores", r.client.OpenAIClient.APIURL)
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

	var vsResp VectorStoreResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &vsResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(vsResp.ID)
	data.Object = types.StringValue(vsResp.Object)
	data.CreatedAt = types.Int64Value(vsResp.CreatedAt)
	data.Status = types.StringValue(vsResp.Status)
	data.UsageBytes = types.Int64Value(vsResp.UsageBytes)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VectorStoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/vector_stores/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
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

	var vsResp VectorStoreResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &vsResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.Status = types.StringValue(vsResp.Status)
	data.CreatedAt = types.Int64Value(vsResp.CreatedAt)
	data.Name = types.StringValue(vsResp.Name)
	data.UsageBytes = types.Int64Value(vsResp.UsageBytes)

	if vsResp.FileCounts != nil {
		data.FileCounts = &VSFileCountsModel{
			InProgress: types.Int64Value(int64(vsResp.FileCounts.InProgress)),
			Completed:  types.Int64Value(int64(vsResp.FileCounts.Completed)),
			Failed:     types.Int64Value(int64(vsResp.FileCounts.Failed)),
			Cancelled:  types.Int64Value(int64(vsResp.FileCounts.Cancelled)),
			Total:      types.Int64Value(int64(vsResp.FileCounts.Total)),
		}
	}

	// Metadata
	if len(vsResp.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range vsResp.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		data.Metadata, _ = types.MapValueFrom(ctx, types.StringType, metadata)
	}

	// Expires After
	if vsResp.ExpiresAfter != nil {
		data.ExpiresAfter = &VSExpiresAfterModel{
			Anchor: types.StringValue(vsResp.ExpiresAfter.Anchor),
			Days:   types.Int64Value(int64(vsResp.ExpiresAfter.Days)),
		}
	}

	// Chunking Stategy not usually returned?

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorStoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VectorStoreResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := map[string]interface{}{}
	if !data.Name.IsNull() {
		updateRequest["name"] = data.Name.ValueString()
	}

	if !data.Metadata.IsNull() {
		metadata := make(map[string]interface{})
		var metaMap map[string]string
		data.Metadata.ElementsAs(ctx, &metaMap, false)
		for k, v := range metaMap {
			metadata[k] = v
		}
		updateRequest["metadata"] = metadata
	}
	if data.ExpiresAfter != nil {
		updateRequest["expires_after"] = map[string]interface{}{
			"anchor": data.ExpiresAfter.Anchor.ValueString(),
			"days":   data.ExpiresAfter.Days.ValueInt64(),
		}
	}

	reqBody, _ := json.Marshal(updateRequest)
	url := fmt.Sprintf("%s/vector_stores/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
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

	if apiResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API error", apiResp.Status)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorStoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VectorStoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/vector_stores/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	apiReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}

	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
	apiReq.Header.Set("OpenAI-Beta", "assistants=v2")
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	http.DefaultClient.Do(apiReq)
}

func (r *VectorStoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
