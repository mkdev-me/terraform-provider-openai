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

var _ resource.Resource = &ThreadResource{}
var _ resource.ResourceWithImportState = &ThreadResource{}

type ThreadResource struct {
	client *OpenAIClient
}

func NewThreadResource() resource.Resource {
	return &ThreadResource{}
}

func (r *ThreadResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thread"
}

type ThreadResourceModel struct {
	ID            types.String         `tfsdk:"id"`
	Messages      []ThreadMessageModel `tfsdk:"messages"`
	Metadata      types.Map            `tfsdk:"metadata"`
	CreatedAt     types.Int64          `tfsdk:"created_at"`
	ToolResources []ToolResourcesModel `tfsdk:"tool_resources"`
}

type ThreadMessageModel struct {
	Role        types.String             `tfsdk:"role"`
	Content     types.String             `tfsdk:"content"`
	Attachments []MessageAttachmentModel `tfsdk:"attachments"`
	FileIDs     []types.String           `tfsdk:"file_ids"` // Legacy
	Metadata    types.Map                `tfsdk:"metadata"`
}

type MessageAttachmentModel struct {
	FileID types.String       `tfsdk:"file_id"`
	Tools  []MessageToolModel `tfsdk:"tools"`
}

type MessageToolModel struct {
	Type types.String `tfsdk:"type"`
}

type ToolResourcesModel struct {
	FileSearch      []FileSearchResourcesModel      `tfsdk:"file_search"`
	CodeInterpreter []CodeInterpreterResourcesModel `tfsdk:"code_interpreter"`
}

type FileSearchResourcesModel struct {
	VectorStoreIDs []types.String `tfsdk:"vector_store_ids"`
}

type CodeInterpreterResourcesModel struct {
	FileIDs []types.String `tfsdk:"file_ids"`
}

func (r *ThreadResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a thread that the assistant can interact with.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the thread.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The timestamp for when the thread was created.",
			},
			"metadata": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Set of key-value pairs that can be attached to the thread.",
			},
			// Messages used to be optional list. In V2 it's same.
			"messages": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "A list of messages to start the thread with.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The role of the entity that is creating the message. Currently only 'user' is supported.",
						},
						"content": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The content of the message.",
						},
						"file_ids": schema.ListAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "A list of file IDs that the message should use (Legacy).",
						},
						"attachments": schema.ListNestedAttribute{
							Optional:            true,
							MarkdownDescription: "A list of files attached to the message, and the tools they should be added to.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"file_id": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "The ID of the file to attach.",
									},
									"tools": schema.ListNestedAttribute{
										Required: true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Required:            true,
													MarkdownDescription: "The type of tool.",
												},
											},
										},
									},
								},
							},
						},
						"metadata": schema.MapAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Set of key-value pairs that can be attached to the message.",
						},
					},
				},
			},
			"tool_resources": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"file_search": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"vector_store_ids": schema.ListAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"code_interpreter": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"file_ids": schema.ListAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *ThreadResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ThreadResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ThreadResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := ThreadCreateRequest{}

	if !data.Metadata.IsNull() {
		metadata := make(map[string]interface{})
		var metaMap map[string]string
		data.Metadata.ElementsAs(ctx, &metaMap, false)
		for k, v := range metaMap {
			metadata[k] = v
		}
		createRequest.Metadata = metadata
	}

	if data.Messages != nil {
		messages := make([]ThreadMessage, 0, len(data.Messages))
		for _, msgModel := range data.Messages {
			msg := ThreadMessage{
				Role:    msgModel.Role.ValueString(),
				Content: msgModel.Content.ValueString(),
			}

			if !msgModel.Metadata.IsNull() {
				metadata := make(map[string]interface{})
				var metaMap map[string]string
				msgModel.Metadata.ElementsAs(ctx, &metaMap, false)
				for k, v := range metaMap {
					metadata[k] = v
				}
				msg.Metadata = metadata
			}

			// Handle Attachments (V2)
			if msgModel.Attachments != nil {
				attachments := make([]AttachmentRequest, 0, len(msgModel.Attachments))
				for _, attModel := range msgModel.Attachments {
					tools := make([]ToolRequest, 0, len(attModel.Tools))
					for _, t := range attModel.Tools {
						tools = append(tools, ToolRequest{Type: t.Type.ValueString()})
					}
					attachments = append(attachments, AttachmentRequest{
						FileID: attModel.FileID.ValueString(),
						Tools:  tools,
					})
				}
				msg.Attachments = attachments
			}

			// Handle Legacy FileIDs if provided (and no attachments?)
			// OpenAI V2 API prefers Attachments.
			if msgModel.FileIDs != nil {
				fileIDs := make([]string, 0, len(msgModel.FileIDs))
				for _, id := range msgModel.FileIDs {
					fileIDs = append(fileIDs, id.ValueString())
				}
				msg.FileIDs = fileIDs
			}

			messages = append(messages, msg)
		}
		createRequest.Messages = messages
	}

	// Tool Resources
	if len(data.ToolResources) > 0 {
		trModel := data.ToolResources[0]
		tr := &ToolResources{}
		if len(trModel.FileSearch) > 0 {
			fsModel := trModel.FileSearch[0]
			ids := make([]string, 0, len(fsModel.VectorStoreIDs))
			for _, id := range fsModel.VectorStoreIDs {
				ids = append(ids, id.ValueString())
			}
			tr.FileSearch = &FileSearchResources{VectorStoreIDs: ids}
		}
		if len(trModel.CodeInterpreter) > 0 {
			ciModel := trModel.CodeInterpreter[0]
			ids := make([]string, 0, len(ciModel.FileIDs))
			for _, id := range ciModel.FileIDs {
				ids = append(ids, id.ValueString())
			}
			tr.CodeInterpreter = &CodeInterpreterResources{FileIDs: ids}
		}
		createRequest.ToolResources = tr
	}

	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing request", err.Error())
		return
	}

	url := fmt.Sprintf("%s/threads", r.client.OpenAIClient.APIURL)
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

	var threadResponse ThreadResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &threadResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(threadResponse.ID)
	data.CreatedAt = types.Int64Value(int64(threadResponse.CreatedAt))
	// We don't read back messages or tool resources into state from Create response usually,
	// because they are often not fully echoed or slightly different.
	// For Thread, messages are initial only.
	// We should rely on configuration for the state of optional parameters unless computed.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ThreadResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ThreadResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/threads/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
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

	var threadResponse ThreadResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &threadResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.CreatedAt = types.Int64Value(int64(threadResponse.CreatedAt))

	// Map metadata
	if len(threadResponse.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range threadResponse.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		data.Metadata, _ = types.MapValueFrom(ctx, types.StringType, metadata)
	}

	// Note: OpenAI Thread GET response does NOT contain the messages.
	// Messages are fetched via /messages endpoint.
	// Since `messages` input in Create is "initial messages", we might not need to read them back and force sync.
	// The previous SDKv2 implementation did NOT read messages back in `Read`. It only read created_at and metadata.
	// So we follow suit.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ThreadResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Implement Update (Metadata and maybe tool resources?)
	var data ThreadResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := map[string]interface{}{}

	if !data.Metadata.IsNull() {
		metadata := make(map[string]interface{})
		var metaMap map[string]string
		data.Metadata.ElementsAs(ctx, &metaMap, false)
		for k, v := range metaMap {
			metadata[k] = v
		}
		updateRequest["metadata"] = metadata
	}
	// ToolResources update supported? Docs say yes for modify thread.
	// Implementation needed if we want full support.

	reqBody, _ := json.Marshal(updateRequest)
	url := fmt.Sprintf("%s/threads/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
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

func (r *ThreadResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ThreadResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/threads/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	apiReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
	apiReq.Header.Set("OpenAI-Beta", "assistants=v2")
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	http.DefaultClient.Do(apiReq)
}

func (r *ThreadResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
