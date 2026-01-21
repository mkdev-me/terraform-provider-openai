package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ImageGenerationResource{}
var _ resource.ResourceWithImportState = &ImageGenerationResource{}

type ImageGenerationResource struct {
	client *OpenAIClient
}

func NewImageGenerationResource() resource.Resource {
	return &ImageGenerationResource{}
}

func (r *ImageGenerationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_generation"
}

type ImageGenerationResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Prompt         types.String `tfsdk:"prompt"`
	Model          types.String `tfsdk:"model"`
	N              types.Int64  `tfsdk:"n"`
	Quality        types.String `tfsdk:"quality"`
	ResponseFormat types.String `tfsdk:"response_format"`
	Size           types.String `tfsdk:"size"`
	Style          types.String `tfsdk:"style"`
	User           types.String `tfsdk:"user"`

	Created types.Int64 `tfsdk:"created"`
	Data    types.List  `tfsdk:"data"` // List of Objects
}

func (r *ImageGenerationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates images from text.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"prompt": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"n": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					planmodifier.Int64(nil), // RequiresReplace? Yes ideally.
				},
			},
			"quality": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"response_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"style": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"data": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"url":            types.StringType,
						"b64_json":       types.StringType,
						"revised_prompt": types.StringType,
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ImageGenerationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	rc_client, ok := req.ProviderData.(*OpenAIClient)
	if ok {
		r.client = rc_client
	}
}

func (r *ImageGenerationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ImageGenerationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqStruct := ImageGenerationRequest{
		Prompt: data.Prompt.ValueString(),
	}

	if !data.Model.IsNull() {
		reqStruct.Model = data.Model.ValueString()
	}
	if !data.N.IsNull() {
		reqStruct.N = int(data.N.ValueInt64())
	}
	if !data.Quality.IsNull() {
		reqStruct.Quality = data.Quality.ValueString()
	}
	if !data.ResponseFormat.IsNull() {
		reqStruct.ResponseFormat = data.ResponseFormat.ValueString()
	}
	if !data.Size.IsNull() {
		reqStruct.Size = data.Size.ValueString()
	}
	if !data.Style.IsNull() {
		reqStruct.Style = data.Style.ValueString()
	}
	if !data.User.IsNull() {
		reqStruct.User = data.User.ValueString()
	}

	reqBody, _ := json.Marshal(reqStruct)
	url := fmt.Sprintf("%s/images/generations", r.client.OpenAIClient.APIURL)

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

	if apiResp.StatusCode != http.StatusOK {
		respBodyBytes, _ := io.ReadAll(apiResp.Body)
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	var imgResp ImageGenerationResponseFramework
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &imgResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.Created = types.Int64Value(imgResp.Created)

	if len(imgResp.Data) > 0 {
		objs := []attr.Value{}
		for _, d := range imgResp.Data {
			obj, _ := types.ObjectValue(
				map[string]attr.Type{
					"url":            types.StringType,
					"b64_json":       types.StringType,
					"revised_prompt": types.StringType,
				},
				map[string]attr.Value{
					"url":            types.StringValue(d.URL),
					"b64_json":       types.StringValue(d.B64JSON),
					"revised_prompt": types.StringValue(d.RevisedPrompt),
				},
			)
			objs = append(objs, obj)
		}
		listVal, _ := types.ListValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"url":            types.StringType,
				"b64_json":       types.StringType,
				"revised_prompt": types.StringType,
			},
		}, objs)
		data.Data = listVal
	} else {
		data.Data = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
			"url":            types.StringType,
			"b64_json":       types.StringType,
			"revised_prompt": types.StringType,
		}})
	}

	data.ID = types.StringValue(fmt.Sprintf("img-%d", imgResp.Created))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ImageGenerationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// No-op read
	var data ImageGenerationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ImageGenerationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *ImageGenerationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *ImageGenerationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
