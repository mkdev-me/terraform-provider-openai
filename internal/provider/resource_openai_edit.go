package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EditResource{}
var _ resource.ResourceWithImportState = &EditResource{}

type EditResource struct {
	client *OpenAIClient
}

func NewEditResource() resource.Resource {
	return &EditResource{}
}

func (r *EditResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edit"
}

type EditResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	Model       types.String  `tfsdk:"model"`
	Input       types.String  `tfsdk:"input"`
	Instruction types.String  `tfsdk:"instruction"`
	Temperature types.Float64 `tfsdk:"temperature"`
	TopP        types.Float64 `tfsdk:"top_p"`
	N           types.Int64   `tfsdk:"n"`

	// Computed
	Object    types.String `tfsdk:"object"`
	Created   types.Int64  `tfsdk:"created"`
	Text      types.String `tfsdk:"text"` // Convenience for first choice text
	EditID    types.String `tfsdk:"edit_id"`
	ModelUsed types.String `tfsdk:"model_used"`
	// We won't map "Choices" and "Usage" fully unless necessary, but let's keep it simple like sdkv2
	// SDKv2 mapped "choices" as list of map.
}

func (r *EditResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The edit resource allows you to edit text using OpenAI's models.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the edit",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model": schema.StringAttribute{
				Description: "ID of the model to use for the edit",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input": schema.StringAttribute{
				Description: "The input text to edit",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instruction": schema.StringAttribute{
				Description: "The instruction that tells the model how to edit the input",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				Description: "How many edits to generate",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"object": schema.StringAttribute{
				Description: "The object type (always 'edit')",
				Computed:    true,
			},
			"created": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) of when the edit was created",
				Computed:    true,
			},
			"text": schema.StringAttribute{
				Description: "The edited text (from the first choice)",
				Computed:    true,
			},
			"edit_id": schema.StringAttribute{
				Description: "The ID of the edit",
				Computed:    true,
			},
			"model_used": schema.StringAttribute{
				Description: "The model used for the edit",
				Computed:    true,
			},
		},
	}
}

func (r *EditResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EditResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EditResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := EditRequest{
		Model:       data.Model.ValueString(),
		Instruction: data.Instruction.ValueString(),
	}

	if !data.Input.IsNull() {
		request.Input = data.Input.ValueString()
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

	path := "edits"
	reqBody, err := json.Marshal(request)
	if err != nil {
		resp.Diagnostics.AddError("Error marshalling request", err.Error())
		return
	}

	respBody, err := r.client.DoRequest("POST", path, reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Error creating edit", err.Error())
		return
	}

	var editResp EditResponse
	if err := json.Unmarshal(respBody, &editResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(editResp.ID)
	data.EditID = types.StringValue(editResp.ID)
	data.Object = types.StringValue(editResp.Object)
	data.Created = types.Int64Value(int64(editResp.Created))
	data.ModelUsed = types.StringValue(editResp.Model)

	if len(editResp.Choices) > 0 {
		data.Text = types.StringValue(editResp.Choices[0].Text)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EditResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Edits are immutable and not retrievable by ID.
}

func (r *EditResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Operation not supported", "OpenAI Edits are immutable")
}

func (r *EditResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op
}

func (r *EditResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Not Supported", "Import is not supported for edits")
}
