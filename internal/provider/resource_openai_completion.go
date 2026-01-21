package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CompletionResource{}
var _ resource.ResourceWithImportState = &CompletionResource{}

type CompletionResource struct {
	client *OpenAIClient
}

func NewCompletionResource() resource.Resource {
	return &CompletionResource{}
}

func (r *CompletionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_completion"
}

type CompletionResourceModel struct {
	ID               types.String  `tfsdk:"id"`
	Model            types.String  `tfsdk:"model"`
	Prompt           types.String  `tfsdk:"prompt"`
	MaxTokens        types.Int64   `tfsdk:"max_tokens"`
	Temperature      types.Float64 `tfsdk:"temperature"`
	TopP             types.Float64 `tfsdk:"top_p"`
	N                types.Int64   `tfsdk:"n"`
	Stream           types.Bool    `tfsdk:"stream"`
	Logprobs         types.Int64   `tfsdk:"logprobs"`
	Echo             types.Bool    `tfsdk:"echo"`
	Stop             types.List    `tfsdk:"stop"`
	PresencePenalty  types.Float64 `tfsdk:"presence_penalty"`
	FrequencyPenalty types.Float64 `tfsdk:"frequency_penalty"`
	BestOf           types.Int64   `tfsdk:"best_of"`
	LogitBias        types.Map     `tfsdk:"logit_bias"`
	User             types.String  `tfsdk:"user"`
	Suffix           types.String  `tfsdk:"suffix"`

	// Computed fields
	Object  types.String `tfsdk:"object"`
	Created types.Int64  `tfsdk:"created"`
	Text    types.String `tfsdk:"text"` // Helper for the first choice's text, common use case
}

func (r *CompletionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The completion resource allows you to generate text completions using OpenAI's models.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the completion",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model": schema.StringAttribute{
				Description: "The model to use for the completion",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt": schema.StringAttribute{
				Description: "The prompt to generate completions for",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_tokens": schema.Int64Attribute{
				Description: "The maximum number of tokens to generate",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"temperature": schema.Float64Attribute{
				Description: "Sampling temperature",
				Optional:    true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"top_p": schema.Float64Attribute{
				Description: "Nucleus sampling parameter",
				Optional:    true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"n": schema.Int64Attribute{
				Description: "How many completions to generate",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"stream": schema.BoolAttribute{
				Description: "Whether to stream back partial progress",
				Optional:    true,
				// PlanModifiers cannot be applied to BoolAttribute directly via helper but requires custom or generic
				// Using BoolAttribute PlanModifiers was added in newer SDK versions, check compatibility.
				// Assuming standard replacement requirement.
			},
			"logprobs": schema.Int64Attribute{
				Description: "Include the log probabilities on the logprobs most likely tokens",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"echo": schema.BoolAttribute{
				Description: "Echo back the prompt in addition to the completion",
				Optional:    true,
			},
			"stop": schema.ListAttribute{
				Description: "Up to 4 sequences where the API will stop generating further tokens",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"presence_penalty": schema.Float64Attribute{
				Description: "Positive values penalize new tokens based on whether they appear in the text so far",
				Optional:    true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"frequency_penalty": schema.Float64Attribute{
				Description: "Positive values penalize new tokens based on their existing frequency in the text so far",
				Optional:    true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"best_of": schema.Int64Attribute{
				Description: "Generates best_of completions server-side and returns the 'best' (the one with the highest log probability per token)",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"logit_bias": schema.MapAttribute{
				Description: "Modify the likelihood of specified tokens appearing in the completion",
				Optional:    true,
				ElementType: types.Float64Type,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Description: "A unique identifier representing your end-user",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"suffix": schema.StringAttribute{
				Description: "The suffix that comes after a completion of inserted text",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object": schema.StringAttribute{
				Description: "The object type (always 'text_completion')",
				Computed:    true,
			},
			"created": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) of when the completion was created",
				Computed:    true,
			},
			"text": schema.StringAttribute{
				Description: "The generated text (from the first choice)",
				Computed:    true,
			},
		},
	}
}

func (r *CompletionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CompletionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CompletionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := CompletionRequest{
		Model:  data.Model.ValueString(),
		Prompt: data.Prompt.ValueString(),
	}

	if !data.MaxTokens.IsNull() {
		request.MaxTokens = int(data.MaxTokens.ValueInt64())
	}
	if !data.Temperature.IsNull() {
		request.Temperature = data.Temperature.ValueFloat64()
	}
	if !data.TopP.IsNull() {
		request.TopP = data.TopP.ValueFloat64()
	}
	if !data.N.IsNull() {
		request.N = int(data.N.ValueInt64())
	}
	if !data.Stream.IsNull() {
		request.Stream = data.Stream.ValueBool()
	}
	if !data.Logprobs.IsNull() {
		val := int(data.Logprobs.ValueInt64())
		request.Logprobs = &val
	}
	if !data.Echo.IsNull() {
		request.Echo = data.Echo.ValueBool()
	}
	if !data.Stop.IsNull() {
		var stops []string
		resp.Diagnostics.Append(data.Stop.ElementsAs(ctx, &stops, false)...)
		request.Stop = stops
	}
	if !data.PresencePenalty.IsNull() {
		request.PresencePenalty = data.PresencePenalty.ValueFloat64()
	}
	if !data.FrequencyPenalty.IsNull() {
		request.FrequencyPenalty = data.FrequencyPenalty.ValueFloat64()
	}
	if !data.BestOf.IsNull() {
		request.BestOf = int(data.BestOf.ValueInt64())
	}
	if !data.LogitBias.IsNull() {
		var logitBias map[string]float64
		resp.Diagnostics.Append(data.LogitBias.ElementsAs(ctx, &logitBias, false)...)
		request.LogitBias = logitBias
	}
	if !data.User.IsNull() {
		request.User = data.User.ValueString()
	}
	if !data.Suffix.IsNull() {
		request.Suffix = data.Suffix.ValueString()
	}

	path := "completions"
	reqBody, err := json.Marshal(request)
	if err != nil {
		resp.Diagnostics.AddError("Error marshalling request", err.Error())
		return
	}

	respBody, err := r.client.DoRequest("POST", path, reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Error creating completion", err.Error())
		return
	}

	var completionResp CompletionResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(completionResp.ID)
	data.Object = types.StringValue(completionResp.Object)
	data.Created = types.Int64Value(int64(completionResp.Created))
	// Completion IDs are ephemeral? No, they have IDs.
	// But usually they can't be retrieved via GET /completions/{id} later.
	// We'll treat them as essentially standalone, maybe with empty Read implementation or just from state.
	// OpenAI completions are generally fire-and-forget/immutable.

	if len(completionResp.Choices) > 0 {
		data.Text = types.StringValue(completionResp.Choices[0].Text)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CompletionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// OpenAI Completions do not support retrieval by ID.
	// We just return what is in state.
}

func (r *CompletionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Operation not supported", "OpenAI Completions are immutable")
}

func (r *CompletionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op
}

func (r *CompletionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Not Supported", "Import is not supported for completions")
}
