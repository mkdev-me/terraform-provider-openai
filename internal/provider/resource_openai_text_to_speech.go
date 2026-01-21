package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &TextToSpeechResource{}
var _ resource.ResourceWithImportState = &TextToSpeechResource{}

type TextToSpeechResource struct {
	client *OpenAIClient
}

func NewTextToSpeechResource() resource.Resource {
	return &TextToSpeechResource{}
}

func (r *TextToSpeechResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_text_to_speech"
}

type TextToSpeechResourceModel struct {
	ID             types.String  `tfsdk:"id"`
	Model          types.String  `tfsdk:"model"`
	Input          types.String  `tfsdk:"input"`
	Voice          types.String  `tfsdk:"voice"`
	ResponseFormat types.String  `tfsdk:"response_format"`
	Speed          types.Float64 `tfsdk:"speed"`
	Instructions   types.String  `tfsdk:"instructions"`
	OutputFile     types.String  `tfsdk:"output_file"`
	CreatedAt      types.Int64   `tfsdk:"created_at"`
}

func (r *TextToSpeechResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates audio from text.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The model to use.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The text input.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"voice": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The voice to use.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"response_format": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Audio format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"speed": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"instructions": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"output_file": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path to save the output audio.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *TextToSpeechResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	rc_client, ok := req.ProviderData.(*OpenAIClient)
	if ok {
		r.client = rc_client
	}
}

func (r *TextToSpeechResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TextToSpeechResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Request
	reqStruct := TextToSpeechRequest{
		Model: data.Model.ValueString(),
		Input: data.Input.ValueString(),
		Voice: data.Voice.ValueString(),
	}

	if !data.ResponseFormat.IsNull() {
		reqStruct.ResponseFormat = data.ResponseFormat.ValueString()
	}
	if !data.Speed.IsNull() {
		reqStruct.Speed = data.Speed.ValueFloat64()
	}
	if !data.Instructions.IsNull() {
		reqStruct.Instructions = data.Instructions.ValueString()
	}

	reqBody, _ := json.Marshal(reqStruct)
	url := fmt.Sprintf("%s/audio/speech", r.client.OpenAIClient.APIURL)

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

	// Write to file
	outPath := data.OutputFile.ValueString()
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		resp.Diagnostics.AddError("Error creating dir", err.Error())
		return
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		resp.Diagnostics.AddError("Error creating file", err.Error())
		return
	}
	defer outFile.Close()

	io.Copy(outFile, apiResp.Body)

	data.ID = types.StringValue(fmt.Sprintf("speech-%d", time.Now().UnixNano()))
	data.CreatedAt = types.Int64Value(time.Now().Unix())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TextToSpeechResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TextToSpeechResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := os.Stat(data.OutputFile.ValueString()); os.IsNotExist(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TextToSpeechResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *TextToSpeechResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Optionally delete file?
	var data TextToSpeechResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	os.Remove(data.OutputFile.ValueString())
}

func (r *TextToSpeechResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
