package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ModerationResource{}
var _ resource.ResourceWithImportState = &ModerationResource{}

type ModerationResource struct {
	client *OpenAIClient
}

func NewModerationResource() resource.Resource {
	return &ModerationResource{}
}

func (r *ModerationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_moderation"
}

type ModerationResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Model types.String `tfsdk:"model"`
	Input types.String `tfsdk:"input"`

	// Computed
	Flagged        types.Bool `tfsdk:"flagged"`
	Categories     types.Map  `tfsdk:"categories"`      // map[string]bool
	CategoryScores types.Map  `tfsdk:"category_scores"` // map[string]float64
}

func (r *ModerationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The moderation resource allows you to check text usage against OpenAI's content policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the moderation",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model": schema.StringAttribute{
				Description: "The model to use for moderation",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input": schema.StringAttribute{
				Description: "The input text to moderate",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"flagged": schema.BoolAttribute{
				Description: "Whether the content was flagged",
				Computed:    true,
			},
			"categories": schema.MapAttribute{
				Description: "Map of category names to boolean values",
				Computed:    true,
				ElementType: types.BoolType,
			},
			"category_scores": schema.MapAttribute{
				Description: "Map of category names to scores",
				Computed:    true,
				ElementType: types.Float64Type,
			},
		},
	}
}

func (r *ModerationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModerationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ModerationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := ModerationRequest{
		Input: data.Input.ValueString(),
	}

	if !data.Model.IsNull() {
		request.Model = data.Model.ValueString()
	}

	path := "moderations"
	reqBody, err := json.Marshal(request)
	if err != nil {
		resp.Diagnostics.AddError("Error marshalling request", err.Error())
		return
	}

	respBody, err := r.client.DoRequest("POST", path, reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Error creating moderation", err.Error())
		return
	}

	var modResp ModerationResponse
	if err := json.Unmarshal(respBody, &modResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(modResp.ID)
	data.Model = types.StringValue(modResp.Model) // API returns model used

	if len(modResp.Results) > 0 {
		res := modResp.Results[0]
		data.Flagged = types.BoolValue(res.Flagged)

		// Convert Categories map to types.Map
		cats := make(map[string]types.Bool)
		for k, v := range res.Categories {
			cats[k] = types.BoolValue(v)
		}
		// Convert to map[string]attr.Value
		// Actually MapValueFrom expects (ctx, type, map[string]interface{}) where interface matches type

		// Since map[string]bool is returned, we can use ElementsAs or manually construct
		// Let's use simple map
		categoriesMap, diag := types.MapValueFrom(ctx, types.BoolType, res.Categories)
		resp.Diagnostics.Append(diag...)
		data.Categories = categoriesMap

		scoresMap, diag := types.MapValueFrom(ctx, types.Float64Type, res.CategoryScores)
		resp.Diagnostics.Append(diag...)
		data.CategoryScores = scoresMap
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModerationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Not retrievable.
}

func (r *ModerationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Operation not supported", "Moderations are immutable")
}

func (r *ModerationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op
}

func (r *ModerationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Not Supported", "Import is not supported for moderations")
}
