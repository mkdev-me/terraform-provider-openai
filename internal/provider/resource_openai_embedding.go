package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EmbeddingResource{}
var _ resource.ResourceWithImportState = &EmbeddingResource{}

type EmbeddingResource struct {
	client *OpenAIClient
}

func NewEmbeddingResource() resource.Resource {
	return &EmbeddingResource{}
}

func (r *EmbeddingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_embedding"
}

type EmbeddingResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Model          types.String `tfsdk:"model"`
	Input          types.String `tfsdk:"input"`
	User           types.String `tfsdk:"user"`
	Dimensions     types.Int64  `tfsdk:"dimensions"`
	EncodingFormat types.String `tfsdk:"encoding_format"`

	// Computed
	Object    types.String `tfsdk:"object"`
	Embedding types.String `tfsdk:"embedding"` // Return as string representation or maybe handle as text?
	// The embedding vector is large, maybe we shouldn't store it in state by default?
	// But SDKv2 probably did.
	// SDKv2 implemented this?
	// resource_openai_embedding.go defined schema "embedding" as TypeString.
	// Wait, no. SDKv2 schema for embedding was likely computed.
	// Let's check previously viewed `resource_openai_embedding.go` (Step 933). Line 12 says `EmbeddingData` has `Embedding json.RawMessage`.
	// The resource only had `Create` and `Read` (dummy).
	// We'll store it as a stringified json or just the base64 string if that's what it is.
}

func (r *EmbeddingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The embedding resource allows you to generate vector embeddings for text.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the embedding",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model": schema.StringAttribute{
				Description: "ID of the model to use",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input": schema.StringAttribute{
				Description: "The input text to embed",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Description: "A unique identifier representing your end-user",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dimensions": schema.Int64Attribute{
				Description: "The number of dimensions the resulting output embeddings should have",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"encoding_format": schema.StringAttribute{
				Description: "The format to return the embeddings in",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object": schema.StringAttribute{
				Description: "The object type",
				Computed:    true,
			},
			"embedding": schema.StringAttribute{
				Description: "The embedding vector",
				Computed:    true,
			},
		},
	}
}

func (r *EmbeddingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EmbeddingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EmbeddingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := EmbeddingRequest{
		Model: data.Model.ValueString(),
		Input: data.Input.ValueString(),
	}

	if !data.User.IsNull() {
		request.User = data.User.ValueString()
	}
	if !data.Dimensions.IsNull() {
		request.Dimensions = int(data.Dimensions.ValueInt64())
	}
	if !data.EncodingFormat.IsNull() {
		request.EncodingFormat = data.EncodingFormat.ValueString()
	}

	path := "embeddings"
	reqBody, err := json.Marshal(request)
	if err != nil {
		resp.Diagnostics.AddError("Error marshalling request", err.Error())
		return
	}

	respBody, err := r.client.DoRequest("POST", path, reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Error creating embedding", err.Error())
		return
	}

	var embedResp EmbeddingResponse
	if err := json.Unmarshal(respBody, &embedResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	// Embeddings don't have IDs?
	// The response doesn't have a top-level ID field in the struct I defined.
	// But `resource_openai_embedding.go` (Step 933) has `EmbeddingResponse` with `Object`, `Data`, `Model`, `Usage`. No ID.
	// We'll generate a synthetic ID.
	data.ID = types.StringValue(fmt.Sprintf("%s-%d", request.Model, embedResp.Usage.TotalTokens)) // Simple synthetic ID or just UUID?
	// Or use hash of input?
	// Let's use computed ID.

	data.Object = types.StringValue(embedResp.Object)

	if len(embedResp.Data) > 0 {
		// Store embedding as string
		embedBytes, _ := json.Marshal(embedResp.Data[0].Embedding)
		data.Embedding = types.StringValue(string(embedBytes))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EmbeddingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Not retrievable.
}

func (r *EmbeddingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Operation not supported", "Embeddings are immutable")
}

func (r *EmbeddingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op
}

func (r *EmbeddingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Not Supported", "Import is not supported for embeddings")
}
