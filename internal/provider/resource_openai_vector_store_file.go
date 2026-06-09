package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// vectorStoreFileReadMaxRetries bounds the read-after-create verification at
	// 5 attempts (~31s worst case with exponential backoff: 1s, 2s, 4s, 8s, 16s).
	vectorStoreFileReadMaxRetries = 5
	vectorStoreFileReadBaseDelay  = 1 * time.Second
)

var _ resource.Resource = &VectorStoreFileResource{}
var _ resource.ResourceWithImportState = &VectorStoreFileResource{}

type VectorStoreFileResource struct {
	client *OpenAIClient
}

func NewVectorStoreFileResource() resource.Resource {
	return &VectorStoreFileResource{}
}

func (r *VectorStoreFileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_store_file"
}

type VectorStoreFileResourceModel struct {
	ID               types.String             `tfsdk:"id"`
	VectorStoreID    types.String             `tfsdk:"vector_store_id"`
	FileID           types.String             `tfsdk:"file_id"`
	ChunkingStrategy *VSChunkingStrategyModel `tfsdk:"chunking_strategy"` // Reusing from vector store

	// Computed
	Object     types.String      `tfsdk:"object"`
	Status     types.String      `tfsdk:"status"`
	CreatedAt  types.Int64       `tfsdk:"created_at"`
	UsageBytes types.Int64       `tfsdk:"usage_bytes"`
	LastError  *VSLastErrorModel `tfsdk:"last_error"`
}

type VSLastErrorModel struct {
	Code    types.String `tfsdk:"code"`
	Message types.String `tfsdk:"message"`
}

func (r *VectorStoreFileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a file in an OpenAI Vector Store.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the vector store file.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vector_store_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the vector store to add the file to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the file to add.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Computed
			"object":      schema.StringAttribute{Computed: true},
			"status":      schema.StringAttribute{Computed: true},
			"created_at":  schema.Int64Attribute{Computed: true},
			"usage_bytes": schema.Int64Attribute{Computed: true},
			"last_error": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"code":    schema.StringAttribute{Computed: true},
					"message": schema.StringAttribute{Computed: true},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"chunking_strategy": schema.SingleNestedBlock{
				MarkdownDescription: "The chunking strategy used to chunk the file(s).",
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

func (r *VectorStoreFileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VectorStoreFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VectorStoreFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := VectorStoreFileCreateRequest{
		FileID: data.FileID.ValueString(),
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

	url := fmt.Sprintf("%s/vector_stores/%s/files", r.client.OpenAIClient.APIURL, data.VectorStoreID.ValueString())
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

	var vsFileResp VectorStoreFileResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &vsFileResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(vsFileResp.ID)
	data.Object = types.StringValue(vsFileResp.Object)
	data.CreatedAt = types.Int64Value(vsFileResp.CreatedAt)
	data.Status = types.StringValue(vsFileResp.Status)
	data.UsageBytes = types.Int64Value(vsFileResp.UsageBytes)

	tflog.Info(ctx, fmt.Sprintf("Vector store file created: %s in vector store %s", vsFileResp.ID, data.VectorStoreID.ValueString()))

	// The OpenAI API is eventually consistent: a GET issued right after creation can
	// return "no file found", especially when many files are created at once (#35).
	// Verify the file is readable before finishing, retrying with exponential backoff.
	verified, err := r.readVectorStoreFileWithRetry(ctx, data.VectorStoreID.ValueString(), vsFileResp.ID, vectorStoreFileReadMaxRetries)
	if err != nil {
		// The file was created; persist state so it is tracked (tainted) instead of orphaned.
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("Error reading vector store file after creation", err.Error())
		return
	}

	data.Status = types.StringValue(verified.Status)
	data.UsageBytes = types.Int64Value(verified.UsageBytes)
	if verified.LastError != nil {
		data.LastError = &VSLastErrorModel{
			Code:    types.StringValue(verified.LastError.Code),
			Message: types.StringValue(verified.LastError.Message),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// getVectorStoreFile retrieves a single vector store file from the API.
// A 404 is returned as an error whose message matches containsRetriableError,
// alongside the HTTP status code so callers can distinguish "gone" from failure.
func (r *VectorStoreFileResource) getVectorStoreFile(vectorStoreID, fileID string) (*VectorStoreFileResponse, int, error) {
	url := fmt.Sprintf("%s/vector_stores/%s/files/%s", r.client.OpenAIClient.APIURL, vectorStoreID, fileID)
	apiReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
	apiReq.Header.Set("OpenAI-Beta", "assistants=v2")
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		return nil, 0, err
	}
	defer apiResp.Body.Close()

	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if apiResp.StatusCode == http.StatusNotFound {
		return nil, apiResp.StatusCode, fmt.Errorf("no file found with id '%s' in vector store '%s'", fileID, vectorStoreID)
	}
	if apiResp.StatusCode != http.StatusOK {
		return nil, apiResp.StatusCode, fmt.Errorf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes))
	}

	var vsFileResp VectorStoreFileResponse
	if err := json.Unmarshal(respBodyBytes, &vsFileResp); err != nil {
		return nil, apiResp.StatusCode, fmt.Errorf("error parsing response: %s", err)
	}
	return &vsFileResp, apiResp.StatusCode, nil
}

// containsRetriableError checks if an error message indicates a retriable error.
// Uses case-insensitive matching to catch "404 Not Found", "No file found", etc.
func containsRetriableError(message string) bool {
	lowerMsg := strings.ToLower(message)
	return strings.Contains(lowerMsg, "no file found") || strings.Contains(lowerMsg, "not found")
}

// retryVectorStoreFileRead runs read, retrying on "not found" errors to handle
// eventual consistency issues with the OpenAI API.
//
// Retry behavior:
//   - Retries up to maxRetries times (must be >= 1)
//   - Exponential backoff: baseDelay * 1, 2, 4, 8, 16...
//   - Only retries on "not found" errors (case-insensitive); other errors fail immediately
func retryVectorStoreFileRead(ctx context.Context, maxRetries int, baseDelay time.Duration, read func() error) error {
	if maxRetries <= 0 {
		return fmt.Errorf("maxRetries must be at least 1 for vector store file read retries")
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoffDuration := time.Duration(1<<uint(attempt-1)) * baseDelay
			tflog.Info(ctx, fmt.Sprintf("Retrying vector store file read after %v (attempt %d/%d)", backoffDuration, attempt+1, maxRetries))
			time.Sleep(backoffDuration)
		}

		err := read()
		if err == nil {
			return nil
		}
		if !containsRetriableError(err.Error()) {
			return err
		}

		lastErr = err
		tflog.Warn(ctx, fmt.Sprintf("Vector store file not found, will retry (attempt %d/%d)", attempt+1, maxRetries))
	}

	tflog.Error(ctx, fmt.Sprintf("Failed to read vector store file after %d attempts", maxRetries))
	return lastErr
}

func (r *VectorStoreFileResource) readVectorStoreFileWithRetry(ctx context.Context, vectorStoreID, fileID string, maxRetries int) (*VectorStoreFileResponse, error) {
	var result *VectorStoreFileResponse
	err := retryVectorStoreFileRead(ctx, maxRetries, vectorStoreFileReadBaseDelay, func() error {
		vsFile, _, err := r.getVectorStoreFile(vectorStoreID, fileID)
		if err != nil {
			return err
		}
		result = vsFile
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *VectorStoreFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VectorStoreFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vsFile, statusCode, err := r.getVectorStoreFile(data.VectorStoreID.ValueString(), data.ID.ValueString())
	if statusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("API error", err.Error())
		return
	}
	vsFileResp := *vsFile

	data.Status = types.StringValue(vsFileResp.Status)
	data.CreatedAt = types.Int64Value(vsFileResp.CreatedAt)
	data.UsageBytes = types.Int64Value(vsFileResp.UsageBytes)
	data.FileID = types.StringValue(vsFileResp.ID) // Note: The ID of the vector store file object is usually the same as file ID? No, wait.
	// Actually, in vector stores, the returned object has ID = file_id. "The ID of the file."

	if vsFileResp.LastError != nil {
		data.LastError = &VSLastErrorModel{
			Code:    types.StringValue(vsFileResp.LastError.Code),
			Message: types.StringValue(vsFileResp.LastError.Message),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorStoreFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Vector Store Files are immutable mostly, no update endpoint in API usually?
	// Wait, typically metadata?
	// API docs: No update endpoint for vector store file. Only List, Create (add), Retrieive, Delete.
	// So Update should just trigger replacement or return error?
	// Terraform framework handles "RequiresReplace" attributes.
	// non-RequiresReplace attributes that change would trigger Update.
	// But we have no attributes that are updatable.
	// So Update shouldn't be called unless we added attributes that are not ForceNew.
	// Just return.
}

func (r *VectorStoreFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VectorStoreFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/vector_stores/%s/files/%s", r.client.OpenAIClient.APIURL, data.VectorStoreID.ValueString(), data.ID.ValueString())
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

func (r *VectorStoreFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected format: vector_store_id:file_id
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: vector_store_id:file_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vector_store_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("file_id"), idParts[1])...)
}
