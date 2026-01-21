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

var _ resource.Resource = &BatchResource{}
var _ resource.ResourceWithImportState = &BatchResource{}

type BatchResource struct {
	client *OpenAIClient
}

func NewBatchResource() resource.Resource {
	return &BatchResource{}
}

func (r *BatchResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_batch"
}

type BatchResourceModel struct {
	ID               types.String             `tfsdk:"id"`
	InputFileID      types.String             `tfsdk:"input_file_id"`
	Endpoint         types.String             `tfsdk:"endpoint"`
	CompletionWindow types.String             `tfsdk:"completion_window"`
	Metadata         types.Map                `tfsdk:"metadata"`
	Status           types.String             `tfsdk:"status"`
	OutputFileID     types.String             `tfsdk:"output_file_id"`
	ErrorFileID      types.String             `tfsdk:"error_file_id"`
	CreatedAt        types.Int64              `tfsdk:"created_at"`
	InProgressAt     types.Int64              `tfsdk:"in_progress_at"`
	ExpiresAt        types.Int64              `tfsdk:"expires_at"`
	FinalizingAt     types.Int64              `tfsdk:"finalizing_at"`
	CompletedAt      types.Int64              `tfsdk:"completed_at"`
	FailedAt         types.Int64              `tfsdk:"failed_at"`
	ExpiredAt        types.Int64              `tfsdk:"expired_at"`
	CancellingAt     types.Int64              `tfsdk:"cancelling_at"`
	CancelledAt      types.Int64              `tfsdk:"cancelled_at"`
	RequestCounts    *BatchRequestCountsModel `tfsdk:"request_counts"`
	Errors           *BatchErrorsModel        `tfsdk:"errors"`
	// Legacy mapping: "error" string field? Legacy provider had "error" mapped to ErrorFileID.
	// We can keep it if we want backward compatibility or cleaner schema.
	// Legacy: "error": TypeString -> "Information about the error that occurred during processing, if any" (mapped to batchResponse.ErrorFileID)
	Error types.String `tfsdk:"error"`
}

type BatchRequestCountsModel struct {
	Total     types.Int64 `tfsdk:"total"`
	Completed types.Int64 `tfsdk:"completed"`
	Failed    types.Int64 `tfsdk:"failed"`
}

type BatchErrorsModel struct {
	Object types.String      `tfsdk:"object"`
	Data   []BatchErrorModel `tfsdk:"data"`
}

type BatchErrorModel struct {
	Code    types.String `tfsdk:"code"`
	Message types.String `tfsdk:"message"`
	Param   types.String `tfsdk:"param"`
	Line    types.Int64  `tfsdk:"line"`
}

func (r *BatchResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI Batch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the batch.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"input_file_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the input file.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"endpoint": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The endpoint to use for the batch request (e.g., '/v1/chat/completions').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"completion_window": schema.StringAttribute{
				Optional: true,
				Computed: true,
				// Default: "24h" - but let's just make it computed/optional and let API default handle it or user specify.
				// Legacy had default "24h".
				// Framework defaults are usually handled by PlanModifiers or just passing nil to API if optional.
				// If we want to enforce default in State, we can use a plan modifier.
				MarkdownDescription: "The time frame within which the batch should be processed. Currently only '24h' is supported.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metadata": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Metadata.",
				PlanModifiers:       []planmodifier.Map{
					// Legacy was ForceNew? Yes: ForceNew: true.
					// So we need RequiresReplace.
					// Wait, legacy map plan modifier?
					// There isn't a simple map plan modifier for RequiresReplace in standard map plan modifiers?
					// Actually, planmodifier.Map doesn't exist?
					// schema.MapAttribute uses `PlanModifiers: []planmodifier.Map`
					// But standard library usually doesn't have RequiresReplace for Map?
					// Usually it's cleaner to implement a custom one or check what's available.
					// Using `mapplanmodifier.RequiresReplace()` (need import).
				},
			},
			// Computed fields
			"status":         schema.StringAttribute{Computed: true},
			"output_file_id": schema.StringAttribute{Computed: true},
			"error_file_id":  schema.StringAttribute{Computed: true},
			"created_at":     schema.Int64Attribute{Computed: true},
			"in_progress_at": schema.Int64Attribute{Computed: true},
			"expires_at":     schema.Int64Attribute{Computed: true},
			"finalizing_at":  schema.Int64Attribute{Computed: true},
			"completed_at":   schema.Int64Attribute{Computed: true},
			"failed_at":      schema.Int64Attribute{Computed: true},
			"expired_at":     schema.Int64Attribute{Computed: true},
			"cancelling_at":  schema.Int64Attribute{Computed: true},
			"cancelled_at":   schema.Int64Attribute{Computed: true},
			"error": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the error file (legacy field naming).",
			},

			"request_counts": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"total":     schema.Int64Attribute{Computed: true},
					"completed": schema.Int64Attribute{Computed: true},
					"failed":    schema.Int64Attribute{Computed: true},
				},
			},

			"errors": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"object": schema.StringAttribute{Computed: true},
					"data": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"code":    schema.StringAttribute{Computed: true},
								"message": schema.StringAttribute{Computed: true},
								"param":   schema.StringAttribute{Computed: true},
								"line":    schema.Int64Attribute{Computed: true},
							},
						},
					},
				},
			},
		},
	}
}

func (r *BatchResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BatchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BatchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := data.Endpoint.ValueString()
	if !strings.HasPrefix(endpoint, "/v1") {
		endpoint = "/v1" + endpoint
	}

	createRequest := BatchCreateRequest{
		InputFileID:      data.InputFileID.ValueString(),
		Endpoint:         endpoint,
		CompletionWindow: "24h", // Default
	}

	if !data.CompletionWindow.IsNull() {
		createRequest.CompletionWindow = data.CompletionWindow.ValueString()
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

	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing request", err.Error())
		return
	}

	url := fmt.Sprintf("%s/batches", r.client.OpenAIClient.APIURL)
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

	var batchResp BatchResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &batchResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(batchResp.ID)
	data.CreatedAt = types.Int64Value(batchResp.CreatedAt)
	data.Status = types.StringValue(batchResp.Status)
	data.ExpiresAt = types.Int64Value(batchResp.ExpiresAt)
	data.CompletionWindow = types.StringValue(batchResp.CompletionWindow)

	// Normalize endpoint for state (remove /v1)
	ep := batchResp.Endpoint
	if strings.HasPrefix(ep, "/v1") {
		ep = strings.TrimPrefix(ep, "/v1")
	}
	data.Endpoint = types.StringValue(ep)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BatchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BatchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/batches/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	apiReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
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
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s", apiResp.Status))
		return
	}

	var batchResp BatchResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &batchResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.InputFileID = types.StringValue(batchResp.InputFileID)
	data.Status = types.StringValue(batchResp.Status)
	data.CreatedAt = types.Int64Value(batchResp.CreatedAt)
	data.ExpiresAt = types.Int64Value(batchResp.ExpiresAt)
	// Optional timestamps
	if batchResp.InProgressAt != nil {
		data.InProgressAt = types.Int64Value(*batchResp.InProgressAt)
	}
	if batchResp.FinalizingAt != nil {
		data.FinalizingAt = types.Int64Value(*batchResp.FinalizingAt)
	}
	if batchResp.CompletedAt != nil {
		data.CompletedAt = types.Int64Value(*batchResp.CompletedAt)
	}
	if batchResp.FailedAt != nil {
		data.FailedAt = types.Int64Value(*batchResp.FailedAt)
	}
	if batchResp.ExpiredAt != nil {
		data.ExpiredAt = types.Int64Value(*batchResp.ExpiredAt)
	}
	if batchResp.CancellingAt != nil {
		data.CancellingAt = types.Int64Value(*batchResp.CancellingAt)
	}
	if batchResp.CancelledAt != nil {
		data.CancelledAt = types.Int64Value(*batchResp.CancelledAt)
	}

	if batchResp.OutputFileID != "" {
		data.OutputFileID = types.StringValue(batchResp.OutputFileID)
	}
	if batchResp.ErrorFileID != "" {
		data.ErrorFileID = types.StringValue(batchResp.ErrorFileID)
		data.Error = types.StringValue(batchResp.ErrorFileID) // Legacy map
	}

	ep := batchResp.Endpoint
	if strings.HasPrefix(ep, "/v1") {
		ep = strings.TrimPrefix(ep, "/v1")
	}
	data.Endpoint = types.StringValue(ep)

	// Metadata
	if len(batchResp.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range batchResp.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		data.Metadata, _ = types.MapValueFrom(ctx, types.StringType, metadata)
	}

	// Request Counts
	if batchResp.RequestCounts != nil {
		data.RequestCounts = &BatchRequestCountsModel{
			Total:     types.Int64Value(int64(batchResp.RequestCounts.Total)),
			Completed: types.Int64Value(int64(batchResp.RequestCounts.Completed)),
			Failed:    types.Int64Value(int64(batchResp.RequestCounts.Failed)),
		}
	}

	// Errors
	if batchResp.Errors != nil {
		errorsData := []BatchErrorModel{}
		for _, e := range batchResp.Errors.Data {
			m := BatchErrorModel{
				Code:    types.StringValue(e.Code),
				Message: types.StringValue(e.Message),
				Param:   types.StringValue(e.Param),
			}
			if e.Line != nil {
				m.Line = types.Int64Value(int64(*e.Line))
			}
			errorsData = append(errorsData, m)
		}

		data.Errors = &BatchErrorsModel{
			Object: types.StringValue(batchResp.Errors.Object),
			Data:   errorsData,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BatchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Immutable
}

func (r *BatchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BatchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/batches/%s/cancel", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	apiReq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return
	}

	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	http.DefaultClient.Do(apiReq)
}

func (r *BatchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
