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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MessageResource{}
var _ resource.ResourceWithImportState = &MessageResource{}

type MessageResource struct {
	client *OpenAIClient
}

func NewMessageResource() resource.Resource {
	return &MessageResource{}
}

func (r *MessageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_message"
}

type MessageResourceModel struct {
	ID          types.String             `tfsdk:"id"`
	ThreadID    types.String             `tfsdk:"thread_id"`
	Role        types.String             `tfsdk:"role"`
	Content     types.String             `tfsdk:"content"` // Input content (simple string)
	Attachments []MessageAttachmentModel `tfsdk:"attachments"`
	Metadata    types.Map                `tfsdk:"metadata"`
	CreatedAt   types.Int64              `tfsdk:"created_at"`
	AssistantID types.String             `tfsdk:"assistant_id"`
	RunID       types.String             `tfsdk:"run_id"`
}

func (r *MessageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a message in a thread.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the message.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thread_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the thread to add the message to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The role of the entity that is creating the message. Currently only 'user' is supported.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The content of the message.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attachments": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "A list of files attached to the message.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"file_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The ID of the file to attach.",
						},
						"tools": schema.ListNestedAttribute{
							Required:            true,
							MarkdownDescription: "The tools that should use this file.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "The type of tool that should use this file.",
									},
								},
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					// Messages are generally immutable in terms of content/attachments once created?
					// API docs say "Modify message" only allows metadata.
					// So attachments should force new.
					// ListRequiresReplace? No standard modifier for List.
					// We'll rely on update logic to error or force replace if we can custom, but for now PlanModifiers on attributes inside?
					// Or just let Update fail/recreate?
					// Standard SDKv2 had ForceNew on attachments.
					// We should enforce ForceNew on the LIST if possible, or attributes.
				},
			},
			"metadata": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Set of key-value pairs that can be attached to the message.",
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The timestamp for when the message was created.",
			},
			"assistant_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "If applicable, the ID of the assistant that authored this message.",
			},
			"run_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "If applicable, the ID of the run associated with this message.",
			},
		},
	}
}

func (r *MessageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MessageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MessageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := MessageCreateRequest{
		Role:    data.Role.ValueString(),
		Content: data.Content.ValueString(),
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

	if data.Attachments != nil {
		attachments := make([]AttachmentRequest, 0, len(data.Attachments))
		for _, attModel := range data.Attachments {
			tools := make([]ToolRequest, 0, len(attModel.Tools))
			for _, t := range attModel.Tools {
				tools = append(tools, ToolRequest{Type: t.Type.ValueString()})
			}
			attachments = append(attachments, AttachmentRequest{
				FileID: attModel.FileID.ValueString(),
				Tools:  tools,
			})
		}
		createRequest.Attachments = attachments
	}

	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing request", err.Error())
		return
	}

	url := fmt.Sprintf("%s/threads/%s/messages", r.client.OpenAIClient.APIURL, data.ThreadID.ValueString())
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

	var messageResponse MessageResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &messageResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(messageResponse.ID)
	data.CreatedAt = types.Int64Value(int64(messageResponse.CreatedAt))
	if messageResponse.AssistantID != "" {
		data.AssistantID = types.StringValue(messageResponse.AssistantID)
	} else {
		data.AssistantID = types.StringNull()
	}
	if messageResponse.RunID != "" {
		data.RunID = types.StringValue(messageResponse.RunID)
	} else {
		data.RunID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MessageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MessageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/threads/%s/messages/%s", r.client.OpenAIClient.APIURL, data.ThreadID.ValueString(), data.ID.ValueString())
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

	var messageResponse MessageResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &messageResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.CreatedAt = types.Int64Value(int64(messageResponse.CreatedAt))
	data.Role = types.StringValue(messageResponse.Role)
	if messageResponse.AssistantID != "" {
		data.AssistantID = types.StringValue(messageResponse.AssistantID)
	}
	if messageResponse.RunID != "" {
		data.RunID = types.StringValue(messageResponse.RunID)
	}

	// Map Content: find text
	// Note: This logic assumes simple text content for now, as existing resource did.
	// If complex content, we might need to adjust logic or schema.
	if len(messageResponse.Content) > 0 {
		for _, content := range messageResponse.Content {
			if content.Type == "text" && content.Text != nil {
				data.Content = types.StringValue(content.Text.Value)
				break
			}
		}
	}

	// Map Metadata
	if len(messageResponse.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range messageResponse.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		data.Metadata, _ = types.MapValueFrom(ctx, types.StringType, metadata)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MessageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MessageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only metadata is updatable for messages
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

	reqBody, _ := json.Marshal(updateRequest)
	url := fmt.Sprintf("%s/threads/%s/messages/%s", r.client.OpenAIClient.APIURL, data.ThreadID.ValueString(), data.ID.ValueString())
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

func (r *MessageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MessageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/threads/%s/messages/%s", r.client.OpenAIClient.APIURL, data.ThreadID.ValueString(), data.ID.ValueString())
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

func (r *MessageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: thread_id:message_id
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected format: thread_id:message_id",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("thread_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
