package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ModelResponseResource{}
var _ resource.ResourceWithImportState = &ModelResponseResource{}

type ModelResponseResource struct {
	client *OpenAIClient
}

func NewModelResponseResource() resource.Resource {
	return &ModelResponseResource{}
}

func (r *ModelResponseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_response"
}

type ModelResponseResourceModel struct {
	ID               types.String  `tfsdk:"id"`
	Model            types.String  `tfsdk:"model"`
	Input            types.String  `tfsdk:"input"`
	MaxOutputTokens  types.Int64   `tfsdk:"max_output_tokens"`
	Temperature      types.Float64 `tfsdk:"temperature"`
	TopP             types.Float64 `tfsdk:"top_p"`
	TopK             types.Int64   `tfsdk:"top_k"`
	Include          types.List    `tfsdk:"include"`
	Instructions     types.String  `tfsdk:"instructions"`
	StopSequences    types.List    `tfsdk:"stop_sequences"`
	FrequencyPenalty types.Float64 `tfsdk:"frequency_penalty"`
	PresencePenalty  types.Float64 `tfsdk:"presence_penalty"`
	User             types.String  `tfsdk:"user"`

	// Computed
	Content types.String `tfsdk:"content"`
}

func (r *ModelResponseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The model_response resource allows you to generate text completions from OpenAI models.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model": schema.StringAttribute{
				Description: "ID of the model to use.",
				Optional:    true,
				Computed:    true,
			},
			"input": schema.StringAttribute{
				Description: "The input text to the model.",
				Optional:    true,
				Computed:    true,
			},
			"max_output_tokens": schema.Int64Attribute{
				Optional: true,
			},
			"temperature": schema.Float64Attribute{
				Optional: true,
			},
			"top_p": schema.Float64Attribute{
				Optional: true,
				Computed: true,
			},
			"top_k": schema.Int64Attribute{
				Optional: true,
			},
			"include": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"instructions": schema.StringAttribute{
				Optional: true,
			},
			"stop_sequences": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"frequency_penalty": schema.Float64Attribute{
				Optional: true,
			},
			"presence_penalty": schema.Float64Attribute{
				Optional: true,
			},
			"user": schema.StringAttribute{
				Optional: true,
			},
			"content": schema.StringAttribute{
				Computed:    true,
				Description: "The generated content.",
			},
		},
	}
}

func (r *ModelResponseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModelResponseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ModelResponseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Logic to call Chat Completion or Completion API depending on model?
	// Or maybe it strictly uses Chat Completion as it has "content" output.
	// "input" usually maps to prompt or messages content.

	// We'll assume Chat Completion for modern models.

	// TODO: fully implement based on SDKv2 logic which I need to check.

	resp.Diagnostics.AddError("Implementation Pending", "Need to verify API logic")
}

func (r *ModelResponseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *ModelResponseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *ModelResponseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *ModelResponseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
