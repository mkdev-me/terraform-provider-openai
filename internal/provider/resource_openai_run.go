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

var _ resource.Resource = &RunResource{}
var _ resource.ResourceWithImportState = &RunResource{}

type RunResource struct {
	client *OpenAIClient
}

func NewRunResource() resource.Resource {
	return &RunResource{}
}

func (r *RunResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_run"
}

type RunResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	ThreadID     types.String   `tfsdk:"thread_id"`
	AssistantID  types.String   `tfsdk:"assistant_id"`
	Model        types.String   `tfsdk:"model"`
	Instructions types.String   `tfsdk:"instructions"`
	Tools        []RunToolModel `tfsdk:"tools"`
	Metadata     types.Map      `tfsdk:"metadata"`
	// Create params
	Temperature         types.Float64            `tfsdk:"temperature"`
	TopP                types.Float64            `tfsdk:"top_p"`
	MaxCompletionTokens types.Int64              `tfsdk:"max_completion_tokens"`
	MaxPromptTokens     types.Int64              `tfsdk:"max_prompt_tokens"`
	TruncationStrategy  *TruncationStrategyModel `tfsdk:"truncation_strategy"`
	ResponseFormat      *ResponseFormatModel     `tfsdk:"response_format"`
	Stream              types.Bool               `tfsdk:"stream"`
	// Computed props
	Object      types.String   `tfsdk:"object"`
	Status      types.String   `tfsdk:"status"`
	CreatedAt   types.Int64    `tfsdk:"created_at"`
	StartedAt   types.Int64    `tfsdk:"started_at"`
	CompletedAt types.Int64    `tfsdk:"completed_at"`
	FileIDs     []types.String `tfsdk:"file_ids"`
	Usage       *RunUsageModel `tfsdk:"usage"`
	Steps       []RunStepModel `tfsdk:"steps"`
}

type RunToolModel struct {
	Type     types.String          `tfsdk:"type"`
	Function *RunToolFunctionModel `tfsdk:"function"`
}

type RunToolFunctionModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Parameters  types.String `tfsdk:"parameters"`
}

type TruncationStrategyModel struct {
	Type          types.String `tfsdk:"type"`
	LastNMessages types.Int64  `tfsdk:"last_n_messages"`
}

type ResponseFormatModel struct {
	Type types.String `tfsdk:"type"`
	// JSONSchema handling omitted for simplicity as per types_run.go simplified handling
}

type RunUsageModel struct {
	PromptTokens     types.Int64 `tfsdk:"prompt_tokens"`
	CompletionTokens types.Int64 `tfsdk:"completion_tokens"`
	TotalTokens      types.Int64 `tfsdk:"total_tokens"`
}

type RunStepModel struct {
	ID        types.String `tfsdk:"id"`
	Object    types.String `tfsdk:"object"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
	Type      types.String `tfsdk:"type"`
	Status    types.String `tfsdk:"status"`
	Details   types.String `tfsdk:"details"` // JSON string
}

func (r *RunResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI assistant run.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the run.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thread_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the thread to run.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assistant_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the assistant to run.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model": schema.StringAttribute{
				Optional:            true,
				Computed:            true, // API defaults to assistant's model
				MarkdownDescription: "The model to use for the run.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instructions": schema.StringAttribute{
				Optional:            true,
				Computed:            true, // API defaults
				MarkdownDescription: "Instructions for the run.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tools": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The type of tool (code_interpreter, retrieval, function).",
						},
						"function": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"name":        schema.StringAttribute{Required: true},
								"description": schema.StringAttribute{Optional: true},
								"parameters":  schema.StringAttribute{Optional: true},
							},
						},
					},
				},
				MarkdownDescription: "Override tools for the run.",
			},
			"metadata": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Metadata.",
			},
			"temperature": schema.Float64Attribute{
				Optional:      true,
				PlanModifiers: []planmodifier.Float64{
					// RequiresReplace usually for runs as parameters are immutable
					// But Framework doesn't store this state?
					// We should likely make them RequiresReplace.
				},
			},
			"top_p": schema.Float64Attribute{
				Optional: true,
			},
			"max_completion_tokens": schema.Int64Attribute{
				Optional: true,
			},
			"max_prompt_tokens": schema.Int64Attribute{
				Optional: true,
			},
			"stream": schema.BoolAttribute{
				Optional: true,
			},
			"truncation_strategy": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type":            schema.StringAttribute{Required: true},
					"last_n_messages": schema.Int64Attribute{Optional: true},
				},
			},
			"response_format": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{Required: true},
				},
			},
			// Computed fields from response
			"object":       schema.StringAttribute{Computed: true},
			"status":       schema.StringAttribute{Computed: true},
			"created_at":   schema.Int64Attribute{Computed: true},
			"started_at":   schema.Int64Attribute{Computed: true},
			"completed_at": schema.Int64Attribute{Computed: true},
			"file_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"usage": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"prompt_tokens":     schema.Int64Attribute{Computed: true},
					"completion_tokens": schema.Int64Attribute{Computed: true},
					"total_tokens":      schema.Int64Attribute{Computed: true},
				},
			},
			"steps": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"object":     schema.StringAttribute{Computed: true},
						"created_at": schema.Int64Attribute{Computed: true},
						"type":       schema.StringAttribute{Computed: true},
						"status":     schema.StringAttribute{Computed: true},
						"details":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (r *RunResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RunResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RunResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := RunCreateRequest{
		AssistantID:  data.AssistantID.ValueString(),
		Model:        data.Model.ValueString(),
		Instructions: data.Instructions.ValueString(),
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

	if len(data.Tools) > 0 {
		tools := make([]map[string]interface{}, 0, len(data.Tools))
		for _, t := range data.Tools {
			tool := map[string]interface{}{
				"type": t.Type.ValueString(),
			}
			if t.Function != nil {
				fn := map[string]interface{}{
					"name": t.Function.Name.ValueString(),
				}
				if !t.Function.Description.IsNull() {
					fn["description"] = t.Function.Description.ValueString()
				}
				if !t.Function.Parameters.IsNull() {
					// Assuming parameters string is JSON
					fn["parameters"] = json.RawMessage(t.Function.Parameters.ValueString())
				}
				tool["function"] = fn
			}
			tools = append(tools, tool)
		}
		createRequest.Tools = tools
	}

	if !data.Temperature.IsNull() {
		t := data.Temperature.ValueFloat64()
		createRequest.Temperature = &t
	}
	if !data.TopP.IsNull() {
		t := data.TopP.ValueFloat64()
		createRequest.TopP = &t
	}
	if !data.MaxCompletionTokens.IsNull() {
		t := int(data.MaxCompletionTokens.ValueInt64())
		createRequest.MaxCompletionTokens = &t
	}
	if !data.MaxPromptTokens.IsNull() {
		t := int(data.MaxPromptTokens.ValueInt64())
		createRequest.MaxPromptTokens = &t
	}
	if !data.Stream.IsNull() {
		s := data.Stream.ValueBool()
		createRequest.Stream = &s
	}

	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing request", err.Error())
		return
	}

	url := fmt.Sprintf("%s/threads/%s/runs", r.client.OpenAIClient.APIURL, data.ThreadID.ValueString())
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

	var runResponse RunResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &runResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	// Populate state
	data.ID = types.StringValue(runResponse.ID)
	data.Object = types.StringValue(runResponse.Object)
	data.CreatedAt = types.Int64Value(runResponse.CreatedAt)
	data.Status = types.StringValue(runResponse.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RunResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RunResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/threads/%s/runs/%s", r.client.OpenAIClient.APIURL, data.ThreadID.ValueString(), data.ID.ValueString())
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

	var runResponse RunResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &runResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.Status = types.StringValue(runResponse.Status)
	data.CreatedAt = types.Int64Value(runResponse.CreatedAt)
	if runResponse.StartedAt != nil {
		data.StartedAt = types.Int64Value(*runResponse.StartedAt)
	}
	if runResponse.CompletedAt != nil {
		data.CompletedAt = types.Int64Value(*runResponse.CompletedAt)
	}

	// FileIDs
	fileIds := make([]types.String, len(runResponse.FileIDs))
	for i, id := range runResponse.FileIDs {
		fileIds[i] = types.StringValue(id)
	}
	data.FileIDs = fileIds

	// Usage
	if runResponse.Usage != nil {
		data.Usage = &RunUsageModel{
			PromptTokens:     types.Int64Value(int64(runResponse.Usage.PromptTokens)),
			CompletionTokens: types.Int64Value(int64(runResponse.Usage.CompletionTokens)),
			TotalTokens:      types.Int64Value(int64(runResponse.Usage.TotalTokens)),
		}
	}

	// Steps (optional listing if completed?)
	// This is expensive on every read, but SDKv2 ran it.
	// We can conditionally do it or just do it.
	if runResponse.Status == "completed" {
		steps := r.fetchSteps(ctx, data.ThreadID.ValueString(), data.ID.ValueString())
		if steps != nil {
			data.Steps = steps
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RunResource) fetchSteps(ctx context.Context, threadId, runId string) []RunStepModel {
	url := fmt.Sprintf("%s/threads/%s/runs/%s/steps", r.client.OpenAIClient.APIURL, threadId, runId)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	if r.client.OpenAIClient.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	var listResp ListRunStepsResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &listResp)

	steps := make([]RunStepModel, 0, len(listResp.Data))
	for _, s := range listResp.Data {
		details := ""
		if len(s.Details) > 0 {
			b, _ := json.Marshal(s.Details)
			details = string(b)
		}
		steps = append(steps, RunStepModel{
			ID:        types.StringValue(s.ID),
			Object:    types.StringValue(s.Object),
			CreatedAt: types.Int64Value(s.CreatedAt),
			Type:      types.StringValue(s.Type),
			Status:    types.StringValue(s.Status),
			Details:   types.StringValue(details),
		})
	}
	return steps
}

func (r *RunResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Runs are generally immutable except metadata.
	// However, Terraform logic might imply replacement if params change.
	// Our plan modifiers handle most replacements.
	// If we land here, it means in-place update.
	var data RunResourceModel
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

	reqBody, _ := json.Marshal(updateRequest)
	url := fmt.Sprintf("%s/threads/%s/runs/%s", r.client.OpenAIClient.APIURL, data.ThreadID.ValueString(), data.ID.ValueString())
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

func (r *RunResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RunResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Cancelling run if not completed?
	// SDKv2 logic: POST /cancel
	url := fmt.Sprintf("%s/threads/%s/runs/%s/cancel", r.client.OpenAIClient.APIURL, data.ThreadID.ValueString(), data.ID.ValueString())
	apiReq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return
	} // Ignore error on cleanup

	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
	apiReq.Header.Set("OpenAI-Beta", "assistants=v2")
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	// Attempt cancel. If fails, it might be completed already.
	http.DefaultClient.Do(apiReq)

	// We don't verify cancellation success as "Delete" just means stop tracking/stop running.
}

func (r *RunResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: thread_id:run_id
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: thread_id:run_id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("thread_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
