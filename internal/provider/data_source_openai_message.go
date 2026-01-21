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

var _ datasource.DataSource = &MessageDataSource{}

func NewMessageDataSource() datasource.DataSource {
	return &MessageDataSource{}
}

type MessageDataSource struct {
	client *OpenAIClient
}

type MessageDataSourceModel struct {
	ID          types.String                       `tfsdk:"id"`
	Object      types.String                       `tfsdk:"object"`
	CreatedAt   types.Int64                        `tfsdk:"created_at"`
	ThreadID    types.String                       `tfsdk:"thread_id"`
	Role        types.String                       `tfsdk:"role"`
	Content     []MessageContentModel              `tfsdk:"content"`
	AssistantID types.String                       `tfsdk:"assistant_id"`
	RunID       types.String                       `tfsdk:"run_id"`
	Metadata    types.Map                          `tfsdk:"metadata"`
	Attachments []MessageDataSourceAttachmentModel `tfsdk:"attachments"`
}

type MessageContentModel struct {
	Type types.String             `tfsdk:"type"`
	Text *MessageContentTextModel `tfsdk:"text"`
	// ImageFile *MessageContentImageFileModel `tfsdk:"image_file"` // Add if supported
}

type MessageContentTextModel struct {
	Value       types.String   `tfsdk:"value"`
	Annotations []types.String `tfsdk:"annotations"` // JSON string list? Or simpler? Annotations are complex.
}

type MessageDataSourceAttachmentModel struct {
	FileID types.String                 `tfsdk:"file_id"`
	Tools  []MessageDataSourceToolModel `tfsdk:"tools"`
}

type MessageDataSourceToolModel struct {
	Type types.String `tfsdk:"type"`
}

func (d *MessageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_message"
}

func (d *MessageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI message.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the message.",
				Required:    true,
			},
			"thread_id": schema.StringAttribute{
				Description: "The ID of the thread this message belongs to.",
				Required:    true,
			},
			"object": schema.StringAttribute{
				Description: "The object type, which is always 'thread.message'.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the message was created.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The role of the entity that produced the message. One of 'user' or 'assistant'.",
				Computed:    true,
			},
			"content": schema.ListNestedAttribute{
				Description: "The content of the message.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The type of content. Can be 'text'.", //  'image_file' later?
							Computed:    true,
						},
						"text": schema.SingleNestedAttribute{
							Description: "The text content.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"value": schema.StringAttribute{
									Description: "The data that makes up the text.",
									Computed:    true,
								},
								"annotations": schema.ListAttribute{
									Description: "Annotations for the text.",
									ElementType: types.StringType,
									Computed:    true,
								},
							},
						},
					},
				},
			},
			"assistant_id": schema.StringAttribute{
				Description: "If applicable, the ID of the assistant that authored this message.",
				Computed:    true,
			},
			"run_id": schema.StringAttribute{
				Description: "If applicable, the ID of the run associated with the authoring of this message.",
				Computed:    true,
			},
			"metadata": schema.MapAttribute{
				Description: "Set of 16 key-value pairs that can be attached to an object.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"attachments": schema.ListNestedAttribute{
				Description: "A list of files attached to the message, and the tools they were added to.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"file_id": schema.StringAttribute{
							Description: "The ID of the file to attach to the message.",
							Computed:    true,
						},
						"tools": schema.ListNestedAttribute{
							Description: "The tools to add this file to.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "The type of tool being defined: code_interpreter or file_search.",
										Computed:    true,
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

func (d *MessageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MessageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MessageDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	threadID := data.ThreadID.ValueString()
	messageID := data.ID.ValueString()
	path := fmt.Sprintf("threads/%s/messages/%s", threadID, messageID)

	apiClient := d.client.OpenAIClient

	respBody, err := apiClient.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Message",
			fmt.Sprintf("Could not read message with ID %s in thread %s: %s", messageID, threadID, err.Error()),
		)
		return
	}

	var messageResponse MessageResponse
	if err := json.Unmarshal(respBody, &messageResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Message Response",
			fmt.Sprintf("Could not parse response for message %s: %s", messageID, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(messageResponse.ID)
	data.Object = types.StringValue(messageResponse.Object)
	data.CreatedAt = types.Int64Value(int64(messageResponse.CreatedAt))
	data.ThreadID = types.StringValue(messageResponse.ThreadID)
	data.Role = types.StringValue(messageResponse.Role)
	data.AssistantID = types.StringValue(messageResponse.AssistantID)
	data.RunID = types.StringValue(messageResponse.RunID)

	// Map Content
	if len(messageResponse.Content) > 0 {
		var content []MessageContentModel
		for _, c := range messageResponse.Content {
			contentModel := MessageContentModel{
				Type: types.StringValue(c.Type),
			}
			if c.Text != nil {
				textModel := &MessageContentTextModel{
					Value: types.StringValue(c.Text.Value),
				}
				// Map Annotations (simplistic for now, assuming JSON strings)
				if len(c.Text.Annotations) > 0 {
					var annotations []types.String
					for _, ann := range c.Text.Annotations {
						annBytes, _ := json.Marshal(ann) // Flatten to JSON string
						annotations = append(annotations, types.StringValue(string(annBytes)))
					}
					textModel.Annotations = annotations
				}
				contentModel.Text = textModel
			}
			content = append(content, contentModel)
		}
		data.Content = content
	} else {
		data.Content = nil
	}

	// Map Metadata
	if len(messageResponse.Metadata) > 0 {
		metadataVals := make(map[string]attr.Value)
		for k, v := range messageResponse.Metadata {
			metadataVals[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		data.Metadata, _ = types.MapValue(types.StringType, metadataVals)
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	// Map Attachments
	if len(messageResponse.Attachments) > 0 {
		var attachments []MessageDataSourceAttachmentModel
		for _, att := range messageResponse.Attachments {
			attModel := MessageDataSourceAttachmentModel{
				FileID: types.StringValue(att.FileID),
			}
			if len(att.Tools) > 0 {
				var tools []MessageDataSourceToolModel
				for _, t := range att.Tools {
					tools = append(tools, MessageDataSourceToolModel{Type: types.StringValue(t.Type)})
				}
				attModel.Tools = tools
			}
			attachments = append(attachments, attModel)
		}
		data.Attachments = attachments
	} else {
		data.Attachments = nil
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
