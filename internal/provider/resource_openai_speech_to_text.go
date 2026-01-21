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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SpeechToTextResource{}
var _ resource.ResourceWithImportState = &SpeechToTextResource{}

type SpeechToTextResource struct {
	client *OpenAIClient
}

func NewSpeechToTextResource() resource.Resource {
	return &SpeechToTextResource{}
}

func (r *SpeechToTextResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_speech_to_text"
}

type SpeechToTextResourceModel struct {
	ID                     types.String  `tfsdk:"id"`
	File                   types.String  `tfsdk:"file"`
	Model                  types.String  `tfsdk:"model"`
	Language               types.String  `tfsdk:"language"`
	Prompt                 types.String  `tfsdk:"prompt"`
	ResponseFormat         types.String  `tfsdk:"response_format"`
	Temperature            types.Float64 `tfsdk:"temperature"`
	Include                types.List    `tfsdk:"include"`
	Stream                 types.Bool    `tfsdk:"stream"`
	TimestampGranularities types.List    `tfsdk:"timestamp_granularities"`

	// Computed outputs
	Text types.String `tfsdk:"text"`
}

func (r *SpeechToTextResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a speech to text transcription.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"language": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt": schema.StringAttribute{
				Optional: true,
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
			"temperature": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"include": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"stream": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"timestamp_granularities": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"text": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *SpeechToTextResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	rc_client, ok := req.ProviderData.(*OpenAIClient)
	if ok {
		r.client = rc_client
	}
}

func (r *SpeechToTextResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SpeechToTextResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare multipart request
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	writer.WriteField("model", data.Model.ValueString())
	if !data.Language.IsNull() {
		writer.WriteField("language", data.Language.ValueString())
	}
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

	if !data.Stream.IsNull() && data.Stream.ValueBool() {
		writer.WriteField("stream", "true")
	}

	if !data.Include.IsNull() {
		var includes []string
		data.Include.ElementsAs(ctx, &includes, false)
		for _, inc := range includes {
			writer.WriteField("include[]", inc)
		}
	}

	if !data.TimestampGranularities.IsNull() {
		var granules []string
		data.TimestampGranularities.ElementsAs(ctx, &granules, false)
		for _, g := range granules {
			writer.WriteField("timestamp_granularities[]", g)
		}
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

	url := fmt.Sprintf("%s/audio/transcriptions", r.client.OpenAIClient.APIURL)
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

	// Simplified handling - just extract text
	// The legacy resource maps to SpeechToTextResponse { Text string }

	if responseFormat == "json" || responseFormat == "verbose_json" {
		var transResp struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(respBodyBytes, &transResp); err != nil {
			resp.Diagnostics.AddError("Error parsing response", err.Error())
			return
		}
		data.Text = types.StringValue(transResp.Text)
	} else {
		data.Text = types.StringValue(string(respBodyBytes))
	}

	data.ID = types.StringValue(fmt.Sprintf("speech-to-text-%d", time.Now().UnixNano()))
	if data.ResponseFormat.IsNull() {
		data.ResponseFormat = types.StringValue(responseFormat)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SpeechToTextResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SpeechToTextResourceModel
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

func (r *SpeechToTextResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *SpeechToTextResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *SpeechToTextResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
