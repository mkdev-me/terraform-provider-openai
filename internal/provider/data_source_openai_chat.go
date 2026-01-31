package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies interface
var _ datasource.DataSource = &ChatCompletionDataSource{}
var _ datasource.DataSource = &ChatCompletionsDataSource{}
var _ datasource.DataSource = &ChatCompletionMessagesDataSource{}

// --- Chat Completion (Singular) ---

func NewChatCompletionDataSource() datasource.DataSource {
	return &ChatCompletionDataSource{}
}

type ChatCompletionDataSource struct {
	client *OpenAIClient
}

type ChatCompletionDataSourceModel struct {
	CompletionID types.String `tfsdk:"completion_id"`
	ID           types.String `tfsdk:"id"`
	Created      types.Int64  `tfsdk:"created"`
	Object       types.String `tfsdk:"object"`
	Model        types.String `tfsdk:"model"`
	Choices      types.List   `tfsdk:"choices"`
	Usage        types.Map    `tfsdk:"usage"`
}

func (d *ChatCompletionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chat_completion"
}

func (d *ChatCompletionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific chat completion by ID.",
		Attributes: map[string]schema.Attribute{
			"completion_id": schema.StringAttribute{
				Description: "The ID of the chat completion to retrieve (format: chatcmpl-xxx)",
				Required:    true,
			},
			"id":      schema.StringAttribute{Computed: true},
			"created": schema.Int64Attribute{Computed: true},
			"object":  schema.StringAttribute{Computed: true},
			"model":   schema.StringAttribute{Computed: true},
			"usage": schema.MapAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"choices": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index":         schema.Int64Attribute{Computed: true},
						"finish_reason": schema.StringAttribute{Computed: true},
						"message": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"role":    schema.StringAttribute{Computed: true},
									"content": schema.StringAttribute{Computed: true},
									"name":    schema.StringAttribute{Computed: true},
									"function_call": schema.ListNestedAttribute{
										Computed: true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name":      schema.StringAttribute{Computed: true},
												"arguments": schema.StringAttribute{Computed: true},
											},
										},
									},
									"tool_calls": schema.ListNestedAttribute{
										Computed: true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"id":   schema.StringAttribute{Computed: true},
												"type": schema.StringAttribute{Computed: true},
												"function": schema.ListNestedAttribute{
													Computed: true,
													NestedObject: schema.NestedAttributeObject{
														Attributes: map[string]schema.Attribute{
															"name":      schema.StringAttribute{Computed: true},
															"arguments": schema.StringAttribute{Computed: true},
														},
													},
												},
											},
										},
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

func (d *ChatCompletionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *OpenAIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ChatCompletionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ChatCompletionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	completionID := data.CompletionID.ValueString()
	url := fmt.Sprintf("/v1/chat/completions/%s", completionID)
	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Legacy behavior: warn and return ID
			resp.Diagnostics.AddWarning("Chat completion not found", fmt.Sprintf("Chat completion with ID '%s' not found.", completionID))
			data.ID = data.CompletionID
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
		resp.Diagnostics.AddError("Error retrieving chat completion", err.Error())
		return
	}

	var completion ChatCompletionResponse // From types_chat.go? Assuming yes or need define.
	// If types_chat.go is in package provider, it is available.
	if err := json.Unmarshal(respBody, &completion); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(completion.ID)
	data.Created = types.Int64Value(int64(completion.Created))
	data.Object = types.StringValue(completion.Object)
	data.Model = types.StringValue(completion.Model)

	// Usage
	usage := map[string]int64{
		"prompt_tokens":     int64(completion.Usage.PromptTokens),
		"completion_tokens": int64(completion.Usage.CompletionTokens),
		"total_tokens":      int64(completion.Usage.TotalTokens),
	}
	data.Usage, _ = types.MapValueFrom(ctx, types.Int64Type, usage)

	// We need to map completion.Choices to the nested structure.
	// Since mapping nested Objects with ListNestedAttribute requires correct struct or types.List.
	// Let's use `attr.Value` approach for choices list.

	choicesList := []attr.Value{}
	choicesElemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"index":         types.Int64Type,
			"finish_reason": types.StringType,
			"message": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"role":    types.StringType,
						"content": types.StringType,
						"name":    types.StringType,
						"function_call": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":      types.StringType,
									"arguments": types.StringType,
								},
							},
						},
						"tool_calls": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"id":   types.StringType,
									"type": types.StringType,
									"function": types.ListType{
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"name":      types.StringType,
												"arguments": types.StringType,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Re-marshalling into local struct to capture ToolCalls if standard type misses them.
	// (Assuming standard type misses unnecessary fields or handled correctly).
	// For simplicity, reusing completion struct if it works, but adding local struct is safer as per previous thought.

	type LocalToolCallFunction struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}
	type LocalToolCall struct {
		ID       string                `json:"id"`
		Type     string                `json:"type"`
		Function LocalToolCallFunction `json:"function"`
	}
	type LocalMessage struct {
		Role         string `json:"role"`
		Content      string `json:"content"`
		Name         string `json:"name,omitempty"`
		FunctionCall *struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function_call,omitempty"`
		ToolCalls []LocalToolCall `json:"tool_calls,omitempty"`
	}
	type LocalChoice struct {
		Index        int          `json:"index"`
		Message      LocalMessage `json:"message"`
		FinishReason string       `json:"finish_reason"`
	}
	type LocalResponse struct {
		ID      string        `json:"id"`
		Created int           `json:"created"`
		Object  string        `json:"object"`
		Model   string        `json:"model"`
		Choices []LocalChoice `json:"choices"`
		Usage   struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	var localComp LocalResponse
	if err := json.Unmarshal(respBody, &localComp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	// Now map localComp to state using loop above but with localComp
	data.ID = types.StringValue(localComp.ID)
	data.Created = types.Int64Value(int64(localComp.Created))
	data.Object = types.StringValue(localComp.Object)
	data.Model = types.StringValue(localComp.Model)

	usage = map[string]int64{
		"prompt_tokens":     int64(localComp.Usage.PromptTokens),
		"completion_tokens": int64(localComp.Usage.CompletionTokens),
		"total_tokens":      int64(localComp.Usage.TotalTokens),
	}
	data.Usage, _ = types.MapValueFrom(ctx, types.Int64Type, usage)

	for _, choice := range localComp.Choices {
		msgAttrs := map[string]attr.Value{
			"role":          types.StringValue(choice.Message.Role),
			"content":       types.StringValue(choice.Message.Content),
			"name":          types.StringNull(),
			"function_call": types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}}),
			"tool_calls":    types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType, "type": types.StringType, "function": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}}}}}),
		}

		if choice.Message.Name != "" {
			msgAttrs["name"] = types.StringValue(choice.Message.Name)
		}

		if choice.Message.FunctionCall != nil {
			fcObj, _ := types.ObjectValue(map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}, map[string]attr.Value{
				"name":      types.StringValue(choice.Message.FunctionCall.Name),
				"arguments": types.StringValue(choice.Message.FunctionCall.Arguments),
			})
			msgAttrs["function_call"], _ = types.ListValue(types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}}, []attr.Value{fcObj})
		}

		if len(choice.Message.ToolCalls) > 0 {
			tcList := []attr.Value{}
			tcType := types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}}
			tcListType := types.ListType{ElemType: tcType}

			toolCallObjType := types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":       types.StringType,
					"type":     types.StringType,
					"function": tcListType,
				},
			}

			for _, tc := range choice.Message.ToolCalls {
				funcObj, _ := types.ObjectValue(map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}, map[string]attr.Value{
					"name":      types.StringValue(tc.Function.Name),
					"arguments": types.StringValue(tc.Function.Arguments),
				})
				funcList, _ := types.ListValue(tcType, []attr.Value{funcObj})

				tcObj, _ := types.ObjectValue(toolCallObjType.AttrTypes, map[string]attr.Value{
					"id":       types.StringValue(tc.ID),
					"type":     types.StringValue(tc.Type),
					"function": funcList,
				})
				tcList = append(tcList, tcObj)
			}
			msgAttrs["tool_calls"], _ = types.ListValue(toolCallObjType, tcList)
		}

		msgObj, _ := types.ObjectValue(choicesElemType.AttrTypes["message"].(types.ListType).ElemType.(types.ObjectType).AttrTypes, msgAttrs)
		msgList, _ := types.ListValue(choicesElemType.AttrTypes["message"].(types.ListType).ElemType, []attr.Value{msgObj})

		choiceObj, _ := types.ObjectValue(choicesElemType.AttrTypes, map[string]attr.Value{
			"index":         types.Int64Value(int64(choice.Index)),
			"finish_reason": types.StringValue(choice.FinishReason),
			"message":       msgList,
		})
		choicesList = append(choicesList, choiceObj)
	}

	data.Choices, _ = types.ListValue(choicesElemType, choicesList)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Chat Completions (Plural) ---

func NewChatCompletionsDataSource() datasource.DataSource {
	return &ChatCompletionsDataSource{}
}

type ChatCompletionsDataSource struct {
	client *OpenAIClient
}

type ChatCompletionsDataSourceModel struct {
	After           types.String `tfsdk:"after"`
	Before          types.String `tfsdk:"before"`
	Limit           types.Int64  `tfsdk:"limit"`
	Order           types.String `tfsdk:"order"`
	Model           types.String `tfsdk:"model"`
	Metadata        types.Map    `tfsdk:"metadata"`
	ID              types.String `tfsdk:"id"`
	HasMore         types.Bool   `tfsdk:"has_more"`
	ChatCompletions types.List   `tfsdk:"chat_completions"`
}

func (d *ChatCompletionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chat_completions"
}

func (d *ChatCompletionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of chat completions.",
		Attributes: map[string]schema.Attribute{
			"after":  schema.StringAttribute{Optional: true},
			"before": schema.StringAttribute{Optional: true},
			"limit": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 100),
				},
			},
			"order": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf("asc", "desc"),
				},
			},
			"model":    schema.StringAttribute{Optional: true},
			"metadata": schema.MapAttribute{Optional: true, ElementType: types.StringType},
			"id":       schema.StringAttribute{Computed: true},
			"has_more": schema.BoolAttribute{Computed: true},
			"chat_completions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"object":  schema.StringAttribute{Computed: true},
						"created": schema.Int64Attribute{Computed: true},
						"model":   schema.StringAttribute{Computed: true},
						"usage":   schema.MapAttribute{Computed: true, ElementType: types.Int64Type},
						"choices": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"index":         schema.Int64Attribute{Computed: true},
									"finish_reason": schema.StringAttribute{Computed: true},
									"message": schema.ListNestedAttribute{
										Computed: true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"role":    schema.StringAttribute{Computed: true},
												"content": schema.StringAttribute{Computed: true},
												"name":    schema.StringAttribute{Computed: true},
												"function_call": schema.ListNestedAttribute{
													Computed: true,
													NestedObject: schema.NestedAttributeObject{
														Attributes: map[string]schema.Attribute{
															"name":      schema.StringAttribute{Computed: true},
															"arguments": schema.StringAttribute{Computed: true},
														},
													},
												},
											},
										},
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

func (d *ChatCompletionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *OpenAIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ChatCompletionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ChatCompletionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := "/v1/chat/completions"
	params := []string{}

	if !data.Limit.IsNull() {
		params = append(params, fmt.Sprintf("limit=%d", data.Limit.ValueInt64()))
	}
	if !data.Order.IsNull() {
		params = append(params, fmt.Sprintf("order=%s", data.Order.ValueString()))
	}
	if !data.After.IsNull() {
		params = append(params, fmt.Sprintf("after=%s", data.After.ValueString()))
	}
	if !data.Before.IsNull() {
		params = append(params, fmt.Sprintf("before=%s", data.Before.ValueString()))
	}
	// Metadata filter... (Legacy Map)
	if !data.Metadata.IsNull() {
		m := make(map[string]string)
		data.Metadata.ElementsAs(ctx, &m, false)
		for k, v := range m {
			params = append(params, fmt.Sprintf("metadata[%s]=%s", k, v))
		}
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		// Warn and return empty
		resp.Diagnostics.AddWarning("Error retrieving chat completions", err.Error())
		data.ID = types.StringValue(fmt.Sprintf("chat_completions_%d", time.Now().Unix()))
		data.HasMore = types.BoolValue(false)

		// Construct empty list with correct type per fixed lint
		listElemType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":      types.StringType,
				"object":  types.StringType,
				"created": types.Int64Type,
				"model":   types.StringType,
				"usage":   types.MapType{ElemType: types.Int64Type},
				"choices": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"index":         types.Int64Type,
							"finish_reason": types.StringType,
							"message": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"role":    types.StringType,
										"content": types.StringType,
										"name":    types.StringType,
										"function_call": types.ListType{
											ElemType: types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"name":      types.StringType,
													"arguments": types.StringType,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		data.ChatCompletions, _ = types.ListValue(listElemType, []attr.Value{})

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Unmarshal logic similar to singular but iterating list
	var listResp struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int    `json:"created"`
			Model   string `json:"model"`
			Choices []struct {
				Index        int    `json:"index"`
				FinishReason string `json:"finish_reason"`
				Message      struct {
					Role         string `json:"role"`
					Content      string `json:"content"`
					Name         string `json:"name"`
					FunctionCall *struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function_call"`
				} `json:"message"`
			} `json:"choices"`
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		} `json:"data"`
		HasMore bool `json:"has_more"`
	}

	if err := json.Unmarshal(respBody, &listResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("chat_completions_%d", time.Now().Unix()))
	data.HasMore = types.BoolValue(listResp.HasMore)

	comps := []attr.Value{}
	// Get Types from Schema
	// (Assumption: Schema correct)
	compType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":      types.StringType,
			"object":  types.StringType,
			"created": types.Int64Type,
			"model":   types.StringType,
			"usage":   types.MapType{ElemType: types.Int64Type},
			"choices": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"index":         types.Int64Type,
				"finish_reason": types.StringType,
				"message": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"role":    types.StringType,
					"content": types.StringType,
					"name":    types.StringType,
					"function_call": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
						"name":      types.StringType,
						"arguments": types.StringType,
					}}},
				}}},
			}}},
		},
	}

	for _, item := range listResp.Data {
		// Usage
		u := map[string]int64{
			"prompt_tokens":     int64(item.Usage.PromptTokens),
			"completion_tokens": int64(item.Usage.CompletionTokens),
			"total_tokens":      int64(item.Usage.TotalTokens),
		}
		usageMap, _ := types.MapValueFrom(ctx, types.Int64Type, u)

		// Choices
		chList := []attr.Value{}
		chType := compType.AttrTypes["choices"].(types.ListType).ElemType.(types.ObjectType)

		for _, ch := range item.Choices {
			msgAttrs := map[string]attr.Value{
				"role":          types.StringValue(ch.Message.Role),
				"content":       types.StringValue(ch.Message.Content),
				"name":          types.StringNull(),
				"function_call": types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}}),
			}
			if ch.Message.Name != "" {
				msgAttrs["name"] = types.StringValue(ch.Message.Name)
			}
			if ch.Message.FunctionCall != nil {
				fcObj, _ := types.ObjectValue(map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}, map[string]attr.Value{
					"name":      types.StringValue(ch.Message.FunctionCall.Name),
					"arguments": types.StringValue(ch.Message.FunctionCall.Arguments),
				})
				msgAttrs["function_call"], _ = types.ListValue(types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}}, []attr.Value{fcObj})
			}

			msgObj, _ := types.ObjectValue(chType.AttrTypes["message"].(types.ListType).ElemType.(types.ObjectType).AttrTypes, msgAttrs)
			msgList, _ := types.ListValue(chType.AttrTypes["message"].(types.ListType).ElemType, []attr.Value{msgObj})

			chObj, _ := types.ObjectValue(chType.AttrTypes, map[string]attr.Value{
				"index":         types.Int64Value(int64(ch.Index)),
				"finish_reason": types.StringValue(ch.FinishReason),
				"message":       msgList,
			})
			chList = append(chList, chObj)
		}
		choicesVal, _ := types.ListValue(chType, chList)

		compObj, _ := types.ObjectValue(compType.AttrTypes, map[string]attr.Value{
			"id":      types.StringValue(item.ID),
			"object":  types.StringValue(item.Object),
			"created": types.Int64Value(int64(item.Created)),
			"model":   types.StringValue(item.Model),
			"usage":   usageMap,
			"choices": choicesVal,
		})
		comps = append(comps, compObj)
	}

	data.ChatCompletions, _ = types.ListValue(compType, comps)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Chat Completion Messages ---

func NewChatCompletionMessagesDataSource() datasource.DataSource {
	return &ChatCompletionMessagesDataSource{}
}

type ChatCompletionMessagesDataSource struct {
	client *OpenAIClient
}

type ChatCompletionMessagesDataSourceModel struct {
	CompletionID types.String `tfsdk:"completion_id"`
	After        types.String `tfsdk:"after"`
	Limit        types.Int64  `tfsdk:"limit"`
	Order        types.String `tfsdk:"order"`
	Messages     types.List   `tfsdk:"messages"`
	HasMore      types.Bool   `tfsdk:"has_more"`
	FirstID      types.String `tfsdk:"first_id"`
	LastID       types.String `tfsdk:"last_id"`
	ID           types.String `tfsdk:"id"`
}

func (d *ChatCompletionMessagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chat_completion_messages"
}

func (d *ChatCompletionMessagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve messages from a chat completion.",
		Attributes: map[string]schema.Attribute{
			"completion_id": schema.StringAttribute{Required: true},
			"after":         schema.StringAttribute{Optional: true},
			"limit": schema.Int64Attribute{
				Optional:   true,
				Validators: []validator.Int64{int64validator.Between(1, 100)},
			},
			"order":    schema.StringAttribute{Optional: true},
			"has_more": schema.BoolAttribute{Computed: true},
			"first_id": schema.StringAttribute{Computed: true},
			"last_id":  schema.StringAttribute{Computed: true},
			"id":       schema.StringAttribute{Computed: true},
			"messages": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role":    schema.StringAttribute{Computed: true},
						"content": schema.StringAttribute{Computed: true},
						"name":    schema.StringAttribute{Computed: true},
						"function_call": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name":      schema.StringAttribute{Computed: true},
									"arguments": schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *ChatCompletionMessagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *OpenAIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ChatCompletionMessagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ChatCompletionMessagesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	completionID := data.CompletionID.ValueString()
	url := fmt.Sprintf("/v1/chat/completions/%s/messages", completionID)
	params := []string{}
	if !data.Limit.IsNull() {
		params = append(params, fmt.Sprintf("limit=%d", data.Limit.ValueInt64()))
	}
	if !data.Order.IsNull() {
		params = append(params, fmt.Sprintf("order=%s", data.Order.ValueString()))
	}
	if !data.After.IsNull() {
		params = append(params, fmt.Sprintf("after=%s", data.After.ValueString()))
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.Diagnostics.AddWarning("Not Found", "Chat completion messages not found")
			data.ID = types.StringValue(fmt.Sprintf("%s-messages", completionID))
			data.HasMore = types.BoolValue(false)
			data.Messages = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{}})

			// Empty list fixed
			msgElemType := types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"role":    types.StringType,
					"content": types.StringType,
					"name":    types.StringType,
					"function_call": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":      types.StringType,
								"arguments": types.StringType,
							},
						},
					},
				},
			}
			data.Messages, _ = types.ListValue(msgElemType, []attr.Value{})

			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
		resp.Diagnostics.AddError("Error", err.Error())
		return
	}

	var listResp struct {
		Data []struct {
			Role         string `json:"role"`
			Content      string `json:"content"`
			Name         string `json:"name"`
			FunctionCall *struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function_call"`
		} `json:"data"`
		HasMore bool   `json:"has_more"`
		FirstID string `json:"first_id"`
		LastID  string `json:"last_id"`
	}

	if err := json.Unmarshal(respBody, &listResp); err != nil {
		resp.Diagnostics.AddError("Error parsing", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s-messages", completionID))
	data.HasMore = types.BoolValue(listResp.HasMore)
	data.FirstID = types.StringValue(listResp.FirstID)
	data.LastID = types.StringValue(listResp.LastID)

	msgs := []attr.Value{}
	msgType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"role":          types.StringType,
			"content":       types.StringType,
			"name":          types.StringType,
			"function_call": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}}},
		},
	}

	for _, m := range listResp.Data {
		attrs := map[string]attr.Value{
			"role":          types.StringValue(m.Role),
			"content":       types.StringValue(m.Content),
			"name":          types.StringNull(),
			"function_call": types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}}),
		}
		if m.Name != "" {
			attrs["name"] = types.StringValue(m.Name)
		}
		if m.FunctionCall != nil {
			fcObj, _ := types.ObjectValue(map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}, map[string]attr.Value{
				"name":      types.StringValue(m.FunctionCall.Name),
				"arguments": types.StringValue(m.FunctionCall.Arguments),
			})
			attrs["function_call"], _ = types.ListValue(types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "arguments": types.StringType}}, []attr.Value{fcObj})
		}

		obj, _ := types.ObjectValue(msgType.AttrTypes, attrs)
		msgs = append(msgs, obj)
	}

	data.Messages, _ = types.ListValue(msgType, msgs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
