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

var _ datasource.DataSource = &MessagesDataSource{}

func NewMessagesDataSource() datasource.DataSource {
	return &MessagesDataSource{}
}

type MessagesDataSource struct {
	client *OpenAIClient
}

type MessagesDataSourceModel struct {
	ThreadID types.String         `tfsdk:"thread_id"`
	Messages []MessageResultModel `tfsdk:"messages"`
	ID       types.String         `tfsdk:"id"` // Dummy ID for Terraform
}

// MessageResultModel mirrors MessageDataSourceModel but for use in a list
type MessageResultModel struct {
	ID          types.String             `tfsdk:"id"`
	Object      types.String             `tfsdk:"object"`
	CreatedAt   types.Int64              `tfsdk:"created_at"`
	ThreadID    types.String             `tfsdk:"thread_id"`
	Role        types.String             `tfsdk:"role"`
	Content     []MessageContentModel    `tfsdk:"content"`
	AssistantID types.String             `tfsdk:"assistant_id"`
	RunID       types.String             `tfsdk:"run_id"`
	Metadata    types.Map                `tfsdk:"metadata"`
	Attachments []MessageAttachmentModel `tfsdk:"attachments"`
}

func (d *MessagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_messages"
}

func (d *MessagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Reusing attributes from MessageDataSource for the nested list
	// Need to manually recreate schema as we can't easily reuse struct schema definition code without refactoring to shared function.

	messageNestedAttributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"thread_id": schema.StringAttribute{
			Computed: true,
		},
		"object": schema.StringAttribute{
			Computed: true,
		},
		"created_at": schema.Int64Attribute{
			Computed: true,
		},
		"role": schema.StringAttribute{
			Computed: true,
		},
		"content": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed: true,
					},
					"text": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"value": schema.StringAttribute{
								Computed: true,
							},
							"annotations": schema.ListAttribute{
								ElementType: types.StringType,
								Computed:    true,
							},
						},
					},
				},
			},
		},
		"assistant_id": schema.StringAttribute{
			Computed: true,
		},
		"run_id": schema.StringAttribute{
			Computed: true,
		},
		"metadata": schema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"attachments": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"file_id": schema.StringAttribute{
						Computed: true,
					},
					"tools": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of messages for a specific OpenAI thread.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"thread_id": schema.StringAttribute{
				Description: "The ID of the thread to list messages for.",
				Required:    true,
			},
			"messages": schema.ListNestedAttribute{
				Description: "List of messages.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: messageNestedAttributes,
				},
			},
		},
	}
}

func (d *MessagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MessagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MessagesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	threadID := data.ThreadID.ValueString()
	apiClient := d.client.OpenAIClient
	var allMessages []MessageResultModel
	cursor := ""

	for {
		queryParams := url.Values{}
		queryParams.Set("limit", "100")
		if cursor != "" {
			queryParams.Set("after", cursor)
		}

		path := fmt.Sprintf("threads/%s/messages?%s", threadID, queryParams.Encode())

		respBody, err := apiClient.DoRequest(http.MethodGet, path, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Messages",
				fmt.Sprintf("Could not list messages for thread %s: %s", threadID, err.Error()),
			)
			return
		}

		// Map simple response struct for list since types_thread_message.go might not have list response?
		// Checking types_thread_message.go... it has MessageResponse but not explicitly ListMessagesResponse.
		// I will define inline or use generic structure.

		var listResp struct {
			Object  string            `json:"object"`
			Data    []MessageResponse `json:"data"`
			FirstID string            `json:"first_id"`
			LastID  string            `json:"last_id"`
			HasMore bool              `json:"has_more"`
		}

		if err := json.Unmarshal(respBody, &listResp); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Messages Response",
				fmt.Sprintf("Could not parse response: %s", err.Error()),
			)
			return
		}

		for _, msgResp := range listResp.Data {
			msgModel := MessageResultModel{
				ID:          types.StringValue(msgResp.ID),
				Object:      types.StringValue(msgResp.Object),
				CreatedAt:   types.Int64Value(int64(msgResp.CreatedAt)),
				ThreadID:    types.StringValue(msgResp.ThreadID),
				Role:        types.StringValue(msgResp.Role),
				AssistantID: types.StringValue(msgResp.AssistantID),
				RunID:       types.StringValue(msgResp.RunID),
			}

			// Map Content
			if len(msgResp.Content) > 0 {
				var content []MessageContentModel
				for _, c := range msgResp.Content {
					contentModel := MessageContentModel{
						Type: types.StringValue(c.Type),
					}
					if c.Text != nil {
						textModel := &MessageContentTextModel{
							Value: types.StringValue(c.Text.Value),
						}
						// Map Annotations
						if len(c.Text.Annotations) > 0 {
							var annotations []types.String
							for _, ann := range c.Text.Annotations {
								annBytes, _ := json.Marshal(ann)
								annotations = append(annotations, types.StringValue(string(annBytes)))
							}
							textModel.Annotations = annotations
						}
						contentModel.Text = textModel
					}
					content = append(content, contentModel)
				}
				msgModel.Content = content
			} else {
				msgModel.Content = nil
			}

			// Map Metadata
			if len(msgResp.Metadata) > 0 {
				metadataVals := make(map[string]attr.Value)
				for k, v := range msgResp.Metadata {
					metadataVals[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
				msgModel.Metadata, _ = types.MapValue(types.StringType, metadataVals)
			} else {
				msgModel.Metadata = types.MapNull(types.StringType)
			}

			// Map Attachments
			if len(msgResp.Attachments) > 0 {
				var attachments []MessageAttachmentModel
				for _, att := range msgResp.Attachments {
					attModel := MessageAttachmentModel{
						FileID: types.StringValue(att.FileID),
					}
					if len(att.Tools) > 0 {
						var tools []MessageToolModel
						for _, t := range att.Tools {
							tools = append(tools, MessageToolModel{Type: types.StringValue(t.Type)})
						}
						attModel.Tools = tools
					}
					attachments = append(attachments, attModel)
				}
				msgModel.Attachments = attachments
			} else {
				msgModel.Attachments = nil
			}

			allMessages = append(allMessages, msgModel)
		}

		if !listResp.HasMore {
			break
		}
		cursor = listResp.LastID
	}

	data.ID = types.StringValue(threadID) // Use thread ID as ID for the data source
	data.Messages = allMessages

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
