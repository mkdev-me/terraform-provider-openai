package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AudioTranslationResource{}
var _ resource.ResourceWithImportState = &AudioTranslationResource{}

type AudioTranslationResource struct {
	client *OpenAIClient
}

func NewAudioTranslationResource() resource.Resource {
	return &AudioTranslationResource{}
}

func (r *AudioTranslationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audio_translation"
}

type AudioTranslationResourceModel struct {
	ID             types.String  `tfsdk:"id"`
	File           types.String  `tfsdk:"file"`
	Model          types.String  `tfsdk:"model"`
	Prompt         types.String  `tfsdk:"prompt"`
	ResponseFormat types.String  `tfsdk:"response_format"`
	Temperature    types.Float64 `tfsdk:"temperature"`

	// Computed outputs
	Text     types.String  `tfsdk:"text"`
	Duration types.Float64 `tfsdk:"duration"`
	Segments types.List    `tfsdk:"segments"` // List of Objects
}

func (r *AudioTranslationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates an audio translation. Note: This resource does not support updates.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path to the audio file to translate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the model to use (e.g., 'whisper-1').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "An optional text to guide the model.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"response_format": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The format of the transcript output (e.g. json, text).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"temperature": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The sampling temperature, between 0 and 1.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"text": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The translated text.",
			},
			"duration": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Duration of the audio in seconds.",
			},
			"segments": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":    types.Int64Type,
						"start": types.Float64Type,
						"end":   types.Float64Type,
						"text":  types.StringType,
					},
				},
				MarkdownDescription: "Segments of the translation.",
			},
		},
	}
}

func (r *AudioTranslationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AudioTranslationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AudioTranslationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare multipart request
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	writer.WriteField("model", data.Model.ValueString())
	if !data.Prompt.IsNull() {
		writer.WriteField("prompt", data.Prompt.ValueString())
	}

	responseFormat := "json"
	if !data.ResponseFormat.IsNull() {
		responseFormat = data.ResponseFormat.ValueString()
		writer.WriteField("response_format", responseFormat)
	}

	if !data.Temperature.IsNull() {
		writer.WriteField("temperature", fmt.Sprintf("%f", data.Temperature.ValueFloat64()))
	}

	// File
	filePathProp := data.File.ValueString()
	file, err := os.Open(filePathProp)
	if err != nil {
		resp.Diagnostics.AddError("Error opening file", err.Error())
		return
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePathProp))
	if err != nil {
		resp.Diagnostics.AddError("Error adding file to form", err.Error())
		return
	}
	io.Copy(part, file)
	writer.Close()

	url := fmt.Sprintf("%s/audio/translations", r.client.OpenAIClient.APIURL)
	apiReq, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiReq.Header.Set("Content-Type", writer.FormDataContentType())
	apiKey := r.client.OpenAIClient.APIKey
	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
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

	respBodyBytes, _ := io.ReadAll(apiResp.Body)

	if responseFormat == "json" || responseFormat == "verbose_json" {
		var transResp TranslationResponseFramework
		if err := json.Unmarshal(respBodyBytes, &transResp); err != nil {
			resp.Diagnostics.AddError("Error parsing response", err.Error())
			return
		}
		data.Text = types.StringValue(transResp.Text)
		data.Duration = types.Float64Value(transResp.Duration)

		// Segments
		if len(transResp.Segments) > 0 {
			segmentObjects := []attr.Value{}
			for _, s := range transResp.Segments {
				obj, _ := types.ObjectValue(
					map[string]attr.Type{
						"id":    types.Int64Type,
						"start": types.Float64Type,
						"end":   types.Float64Type,
						"text":  types.StringType,
					},
					map[string]attr.Value{
						"id":    types.Int64Value(int64(s.ID)),
						"start": types.Float64Value(s.Start),
						"end":   types.Float64Value(s.End),
						"text":  types.StringValue(s.Text),
					},
				)
				segmentObjects = append(segmentObjects, obj)
			}

			listVal, _ := types.ListValue(types.ObjectType{AttrTypes: map[string]attr.Type{
				"id":    types.Int64Type,
				"start": types.Float64Type,
				"end":   types.Float64Type,
				"text":  types.StringType,
			}}, segmentObjects)
			data.Segments = listVal
		} else {
			data.Segments = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
				"id":    types.Int64Type,
				"start": types.Float64Type,
				"end":   types.Float64Type,
				"text":  types.StringType,
			}})
		}
	} else {
		// Text/SRT/VTT
		data.Text = types.StringValue(string(respBodyBytes))
		data.Duration = types.Float64Null()
		data.Segments = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
			"id":    types.Int64Type,
			"start": types.Float64Type,
			"end":   types.Float64Type,
			"text":  types.StringType,
		}})
	}

	data.ID = types.StringValue(fmt.Sprintf("translation-%d", time.Now().UnixNano()))
	if data.ResponseFormat.IsNull() {
		data.ResponseFormat = types.StringValue(responseFormat)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AudioTranslationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AudioTranslationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := os.Stat(data.File.ValueString()); os.IsNotExist(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AudioTranslationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *AudioTranslationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *AudioTranslationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
