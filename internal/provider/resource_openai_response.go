package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ResponseResource{}
var _ resource.ResourceWithConfigure = &ResponseResource{}

type ResponseResource struct {
	client *OpenAIClient
}

func NewResponseResource() resource.Resource {
	return &ResponseResource{}
}

func (r *ResponseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_response"
}

type ResponseResourceModel struct {
	Model              types.String  `tfsdk:"model"`
	Input              types.String  `tfsdk:"input"`
	ID                 types.String  `tfsdk:"id"`
	CreatedAt          types.Int64   `tfsdk:"created_at"`
	Output             types.List    `tfsdk:"output"`
	ReasoningEffort    types.String  `tfsdk:"reasoning_effort"`
	Metadata           types.Map     `tfsdk:"metadata"`
	Temperature        types.Float64 `tfsdk:"temperature"`
	TopP               types.Float64 `tfsdk:"top_p"`
	TopLogprobs        types.Int64   `tfsdk:"top_logprobs"`
	MaxOutputTokens    types.Int64   `tfsdk:"max_output_tokens"`
	MaxToolCalls       types.Int64   `tfsdk:"max_tool_calls"`
	ParallelToolCalls  types.Bool    `tfsdk:"parallel_tool_calls"`
	Truncation         types.String  `tfsdk:"truncation"`
	Tools              types.List    `tfsdk:"tools"`
	ToolChoice         types.String  `tfsdk:"tool_choice"`
	ResponseFormat     types.String  `tfsdk:"response_format"`
	Instructions       types.String  `tfsdk:"instructions"`
	PreviousResponseID types.String  `tfsdk:"previous_response_id"`
	Include            types.List    `tfsdk:"include"`
	Prompt             *PromptModel  `tfsdk:"prompt"`
	ConversationID     types.String  `tfsdk:"conversation_id"`
	Content            types.String  `tfsdk:"content"`
}

type PromptModel struct {
	ID        types.String `tfsdk:"id"`
	Version   types.String `tfsdk:"version"`
	Variables types.String `tfsdk:"variables"` // JSON string
}

type ResponseToolModel struct {
	Type     types.String              `tfsdk:"type"`
	Function ResponseFunctionToolModel `tfsdk:"function"`
}

type ResponseFunctionToolModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Parameters  types.String `tfsdk:"parameters"`
}

type ResponseOutputItem struct {
	Type    types.String `tfsdk:"type"`
	Content types.String `tfsdk:"content"`
}

func (r *ResponseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates a response using the OpenAI Responses API.",
		Attributes: map[string]schema.Attribute{
			"model": schema.StringAttribute{
				MarkdownDescription: "The model to use for the response.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input": schema.StringAttribute{
				MarkdownDescription: "The input text for the response.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"reasoning_effort": schema.StringAttribute{
				MarkdownDescription: "Constrains effort on reasoning for reasoning models. Valid values are `low`, `medium`, `high`.",
				Optional:            true,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Set of key-value pairs that can be attached to an object. This can be useful for storing additional information about the object in a structured format.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"temperature": schema.Float64Attribute{
				MarkdownDescription: "What sampling temperature to use, between 0 and 2. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic.",
				Optional:            true,
			},
			"top_p": schema.Float64Attribute{
				MarkdownDescription: "An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass.",
				Optional:            true,
			},
			"top_logprobs": schema.Int64Attribute{
				MarkdownDescription: "An integer between 0 and 20 specifying the number of most likely tokens to return at each token position.",
				Optional:            true,
			},
			"max_output_tokens": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of tokens to generate in the response.",
				Optional:            true,
			},
			"max_tool_calls": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of tool calls to make in the response.",
				Optional:            true,
			},
			"parallel_tool_calls": schema.BoolAttribute{
				MarkdownDescription: "Whether to allow parallel tool calls. Defaults to true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"truncation": schema.StringAttribute{
				MarkdownDescription: "Controls how the model truncates the context if it exceeds the maximum token limit. Valid values: `auto`, `disabled`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "disabled"),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the generated response.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "The Unix timestamp (in seconds) of when the response was created.",
				Computed:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The concatenated text content of the response. This is a convenience attribute for easy access to the generated text.",
				Computed:            true,
			},
			"output": schema.ListNestedAttribute{
				MarkdownDescription: "The generated output items.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed: true,
						},
						"content": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The content of the output item. Currently only text content is extracted.",
						},
					},
				},
			},
			"tools": schema.ListNestedAttribute{
				MarkdownDescription: "A list of tools the model may call. Currently, only functions are supported as a tool.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the tool. Common values: `function`, `web_search`, `file_search`, `computer_use`, `code_interpreter`.",
							Required:            true,
						},
						"function": schema.SingleNestedAttribute{
							MarkdownDescription: "Function definition for the tool.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "The name of the function to be called.",
									Required:            true,
								},
								"description": schema.StringAttribute{
									MarkdownDescription: "A description of what the function does, used by the model to choose when and how to call the function.",
									Optional:            true,
								},
								"parameters": schema.StringAttribute{
									MarkdownDescription: "The parameters the functions accepts, described as a JSON Schema object.",
									Required:            true,
								},
							},
						},
					},
				},
			},
			"tool_choice": schema.StringAttribute{
				MarkdownDescription: "Controls which (if any) tool is called by the model. Can be `none`, `auto`, `required`, or a specific function name.",
				Optional:            true,
			},
			"response_format": schema.StringAttribute{
				MarkdownDescription: "Specifies the format that the model must output. Compatible with `json_object`, `json_schema`, `text`.",
				Optional:            true,
			},
			"instructions": schema.StringAttribute{
				MarkdownDescription: "A system (or developer) message inserted into the model's context.",
				Optional:            true,
			},
			"previous_response_id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the previous response to the model. Use this to create multi-turn conversations.",
				Optional:            true,
			},
			"conversation_id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the conversation to initiate or continue.",
				Optional:            true,
			},
			"prompt": schema.SingleNestedAttribute{
				MarkdownDescription: "Reference to a prompt template and its variables.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "The unique ID of the prompt template to use.",
						Required:            true,
					},
					"version": schema.StringAttribute{
						MarkdownDescription: "Optional version of the prompt template.",
						Optional:            true,
					},
					"variables": schema.StringAttribute{
						MarkdownDescription: "JSON string of variables to pass to the prompt template.",
						Optional:            true,
					},
				},
			},
			"include": schema.ListAttribute{
				MarkdownDescription: "Specify additional output data to include in the model response. Currently supported values include `web_search_call.action.sources`, `code_interpreter_call.outputs`, etc.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *ResponseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ResponseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResponseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReqData := CreateResponseRequest{
		Model: data.Model.ValueString(),
		Input: data.Input.ValueString(),
		Store: true,
	}

	if !data.ReasoningEffort.IsNull() {
		apiReqData.Reasoning = &ReasoningConfig{
			Effort: data.ReasoningEffort.ValueString(),
		}
	}
	if !data.Metadata.IsNull() {
		meta := make(map[string]interface{})
		for k, v := range data.Metadata.Elements() {
			meta[k] = v.(types.String).ValueString()
		}
		apiReqData.Metadata = meta
	}
	if !data.Temperature.IsNull() {
		v := data.Temperature.ValueFloat64()
		apiReqData.Temperature = &v
	}
	if !data.TopP.IsNull() {
		v := data.TopP.ValueFloat64()
		apiReqData.TopP = &v
	}
	if !data.TopLogprobs.IsNull() {
		v := data.TopLogprobs.ValueInt64()
		apiReqData.TopLogprobs = &v
	}
	if !data.MaxOutputTokens.IsNull() {
		v := data.MaxOutputTokens.ValueInt64()
		apiReqData.MaxOutputTokens = &v
	}
	if !data.MaxToolCalls.IsNull() {
		v := data.MaxToolCalls.ValueInt64()
		apiReqData.MaxToolCalls = &v
	}
	if !data.ParallelToolCalls.IsNull() {
		v := data.ParallelToolCalls.ValueBool()
		apiReqData.ParallelToolCalls = &v
	}
	if !data.Truncation.IsNull() {
		v := data.Truncation.ValueString()
		apiReqData.Truncation = &v
	}

	if !data.Tools.IsNull() {
		var tools []ResponseToolModel
		data.Tools.ElementsAs(ctx, &tools, false)
		toolsReq := []ToolConfig{}
		for _, t := range tools {
			var fReq *FunctionConfig
			// Check if Function Name is set to determine if function block exists
			if !t.Function.Name.IsNull() {
				fReq = &FunctionConfig{
					Name:        t.Function.Name.ValueString(),
					Description: t.Function.Description.ValueString(),
					Parameters:  json.RawMessage(t.Function.Parameters.ValueString()),
				}
			}
			toolsReq = append(toolsReq, ToolConfig{
				Type:     t.Type.ValueString(),
				Function: fReq,
			})
		}
		apiReqData.Tools = toolsReq
	}

	if !data.ToolChoice.IsNull() {
		// Simple string implementation for now (auto, none, required, or function name)
		// For more complex structure (e.g. specific function object), we'd need more logic.
		// Assuming user passes string.
		apiReqData.ToolChoice = data.ToolChoice.ValueString()
	}

	if !data.ResponseFormat.IsNull() {
		// Maps to `text: { format: ... }` in API
		rfVal := data.ResponseFormat.ValueString()
		// If input is simple string (not JSON), wrap it in {"type": "..."}
		if !strings.HasPrefix(strings.TrimSpace(rfVal), "{") {
			apiReqData.Text = &TextConfig{Format: map[string]interface{}{"type": rfVal}}
		} else {
			// If it is JSON (e.g. for json_schema), pass as RawMessage ?
			// Wait, Format is interface{}. RawMessage needs to be unmarshalled or passed as object.
			// Better: Unmarshal to map or RawMessage.
			// For simplicity: If it parses as JSON, use it. Else wrap it.
			var js map[string]interface{}
			if err := json.Unmarshal([]byte(rfVal), &js); err == nil {
				apiReqData.Text = &TextConfig{Format: js}
			} else {
				// Fallback to type wrapper
				apiReqData.Text = &TextConfig{Format: map[string]interface{}{"type": rfVal}}
			}
		}
	}

	if !data.Instructions.IsNull() {
		apiReqData.Instructions = data.Instructions.ValueStringPointer()
	}

	if !data.PreviousResponseID.IsNull() {
		apiReqData.PreviousResponseID = data.PreviousResponseID.ValueStringPointer()
	}

	if !data.ConversationID.IsNull() {
		apiReqData.Conversation = data.ConversationID.ValueStringPointer()
	}

	if data.Prompt != nil {
		pConf := PromptConfig{
			ID: data.Prompt.ID.ValueString(),
		}
		if !data.Prompt.Version.IsNull() {
			v := data.Prompt.Version.ValueString()
			pConf.Version = &v
		}
		if !data.Prompt.Variables.IsNull() {
			pConf.Variables = json.RawMessage(data.Prompt.Variables.ValueString())
		}
		apiReqData.Prompt = &pConf
	}

	if !data.Include.IsNull() {
		var inc []string
		data.Include.ElementsAs(ctx, &inc, false)
		apiReqData.Include = inc
	}

	reqBody, err := json.Marshal(apiReqData)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing request", err.Error())
		return
	}

	url := fmt.Sprintf("%s/responses", r.client.OpenAIClient.APIURL)
	apiReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
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

	var apiRespData ResponseResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &apiRespData); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(apiRespData.ID)
	data.CreatedAt = types.Int64Value(apiRespData.CreatedAt)

	outputItems := r.mapAPIOutputToModel(apiRespData.Output)
	outputList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type":    types.StringType,
			"content": types.StringType,
		},
	}, outputItems)
	resp.Diagnostics.Append(diags...)
	data.Output = outputList

	// Populate convenience 'content' field
	var allContent string
	for _, item := range outputItems {
		allContent += item.Content.ValueString()
	}
	data.Content = types.StringValue(allContent)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResponseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResponseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/responses/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	apiReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
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
		respBodyBytes, _ := io.ReadAll(apiResp.Body)
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	var apiRespData ResponseResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &apiRespData); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.CreatedAt = types.Int64Value(apiRespData.CreatedAt)

	outputItems := r.mapAPIOutputToModel(apiRespData.Output)
	outputList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type":    types.StringType,
			"content": types.StringType,
		},
	}, outputItems)
	resp.Diagnostics.Append(diags...)
	data.Output = outputList

	// Populate convenience 'content' field
	var allContent string
	for _, item := range outputItems {
		allContent += item.Content.ValueString()
	}
	data.Content = types.StringValue(allContent)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResponseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Replacing is enforced by planmodifiers, so this shouldn't be called for model/input changes.
	resp.Diagnostics.AddError("Update not supported", "The openai_response resource does not support updates.")
}

func (r *ResponseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: API deletion not strictly required or supported.
	// Removing from state is sufficient.
	resp.State.RemoveResource(ctx)
}

// API structs (reused from previous attempt but kept here for self-containment)
type CreateResponseRequest struct {
	Model              string                 `json:"model"`
	Input              string                 `json:"input"`
	Store              bool                   `json:"store"`
	Reasoning          *ReasoningConfig       `json:"reasoning,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	Temperature        *float64               `json:"temperature,omitempty"`
	TopP               *float64               `json:"top_p,omitempty"`
	TopLogprobs        *int64                 `json:"top_logprobs,omitempty"`
	MaxOutputTokens    *int64                 `json:"max_output_tokens,omitempty"`
	MaxToolCalls       *int64                 `json:"max_tool_calls,omitempty"`
	ParallelToolCalls  *bool                  `json:"parallel_tool_calls,omitempty"`
	Truncation         *string                `json:"truncation,omitempty"`
	Tools              []ToolConfig           `json:"tools,omitempty"`
	ToolChoice         interface{}            `json:"tool_choice,omitempty"`
	Text               *TextConfig            `json:"text,omitempty"`
	Instructions       *string                `json:"instructions,omitempty"`
	PreviousResponseID *string                `json:"previous_response_id,omitempty"`
	Include            []string               `json:"include,omitempty"`
	Prompt             *PromptConfig          `json:"prompt,omitempty"`
	Conversation       *string                `json:"conversation,omitempty"` // ID only
}

type TextConfig struct {
	Format interface{} `json:"format,omitempty"`
}

type PromptConfig struct {
	ID        string          `json:"id"`
	Version   *string         `json:"version,omitempty"`
	Variables json.RawMessage `json:"variables,omitempty"`
}

type ToolConfig struct {
	Type     string          `json:"type"`
	Function *FunctionConfig `json:"function,omitempty"`
}

type FunctionConfig struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

type ReasoningConfig struct {
	Effort string `json:"effort,omitempty"`
}

type ResponseResponse struct {
	ID        string          `json:"id"`
	CreatedAt int64           `json:"created_at"`
	Output    []APIOutputItem `json:"output"`
}

type APIOutputItem struct {
	Type    string            `json:"type"`
	Content interface{}       `json:"content"`
	Message *APIOutputMessage `json:"message,omitempty"`
}

type APIOutputMessage struct {
	Content interface{} `json:"content"`
}

type ResponseOutputModel struct {
	Type    types.String `tfsdk:"type"`
	Content types.String `tfsdk:"content"`
}

func (r *ResponseResource) mapAPIOutputToModel(items []APIOutputItem) []ResponseOutputModel {
	var models []ResponseOutputModel
	for _, item := range items {
		contentStr := ""
		switch v := item.Content.(type) {
		case nil:
			// Do nothing
		case string:
			contentStr = v
		case []interface{}:
			// Handle list of content parts, e.g. [{"type":"text", "text":"..."}]
			for _, part := range v {
				if partMap, ok := part.(map[string]interface{}); ok {
					if typeVal, ok := partMap["type"].(string); ok && (typeVal == "text" || typeVal == "output_text") {
						if textVal, ok := partMap["text"].(string); ok {
							contentStr += textVal
						}
					}
				}
			}
		case map[string]interface{}:
			// Handle single object if applicable, though usually string or list
			if typeVal, ok := v["type"].(string); ok && typeVal == "text" {
				if textVal, ok := v["text"].(string); ok {
					contentStr = textVal
				}
			}
		default:
			// Fallback: JSON stringify or empty
			if b, err := json.Marshal(v); err == nil {
				contentStr = string(b)
			}
		}

		if item.Message != nil && item.Message.Content != nil {
			// If message content exists, it might override or be the primary content
			switch v := item.Message.Content.(type) {
			case string:
				contentStr = v
			case []interface{}:
				for _, part := range v {
					if partMap, ok := part.(map[string]interface{}); ok {
						if typeVal, ok := partMap["type"].(string); ok && typeVal == "text" {
							if textVal, ok := partMap["text"].(string); ok {
								contentStr += textVal
							}
						}
					}
				}
			default:
				if b, err := json.Marshal(v); err == nil {
					contentStr = string(b)
				}
			}
		}

		models = append(models, ResponseOutputModel{
			Type:    types.StringValue(item.Type),
			Content: types.StringValue(contentStr),
		})
	}
	return models
}
