package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure implementation satisfies interface
var _ datasource.DataSource = &BatchDataSource{}
var _ datasource.DataSource = &BatchesDataSource{}

func NewBatchDataSource() datasource.DataSource {
	return &BatchDataSource{}
}

type BatchDataSource struct {
	client *OpenAIClient
}

type BatchDataSourceModel struct {
	BatchID          types.String `tfsdk:"batch_id"`
	ProjectID        types.String `tfsdk:"project_id"`
	ID               types.String `tfsdk:"id"`
	InputFileID      types.String `tfsdk:"input_file_id"`
	Endpoint         types.String `tfsdk:"endpoint"`
	CompletionWindow types.String `tfsdk:"completion_window"`
	OutputFileID     types.String `tfsdk:"output_file_id"`
	ErrorFileID      types.String `tfsdk:"error_file_id"`
	Status           types.String `tfsdk:"status"`
	CreatedAt        types.Int64  `tfsdk:"created_at"`
	InProgressAt     types.Int64  `tfsdk:"in_progress_at"`
	ExpiresAt        types.Int64  `tfsdk:"expires_at"`
	CompletedAt      types.Int64  `tfsdk:"completed_at"`
	RequestCounts    types.Map    `tfsdk:"request_counts"` // Map<Int>
	Error            types.String `tfsdk:"error"`
	Metadata         types.Map    `tfsdk:"metadata"`
}

func (d *BatchDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_batch"
}

func (d *BatchDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI batch job.",
		Attributes: map[string]schema.Attribute{
			"batch_id": schema.StringAttribute{
				Description: "The ID of the batch job to retrieve",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project associated with the batch job. If not specified, the API key's default project will be used.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the batch job",
				Computed:    true,
			},
			"input_file_id": schema.StringAttribute{
				Description: "The ID of the input file used for the batch",
				Computed:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "The endpoint used for the batch request (e.g., '/v1/chat/completions')",
				Computed:    true,
			},
			"completion_window": schema.StringAttribute{
				Description: "The time window specified for batch completion",
				Computed:    true,
			},
			"output_file_id": schema.StringAttribute{
				Description: "The ID of the output file (if available)",
				Computed:    true,
			},
			"error_file_id": schema.StringAttribute{
				Description: "The ID of the error file (if available)",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the batch job",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The timestamp when the batch job was created",
				Computed:    true,
			},
			"in_progress_at": schema.Int64Attribute{
				Description: "The timestamp when the batch job began processing",
				Computed:    true,
			},
			"expires_at": schema.Int64Attribute{
				Description: "The timestamp when the batch job expires",
				Computed:    true,
			},
			"completed_at": schema.Int64Attribute{
				Description: "The timestamp when the batch job completed",
				Computed:    true,
			},
			"request_counts": schema.MapAttribute{
				Description: "Statistics about request processing",
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"error": schema.StringAttribute{
				Description: "Information about errors that occurred during processing (JSON string)",
				Computed:    true,
			},
			"metadata": schema.MapAttribute{
				Description: "Custom metadata attached to the batch job",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *BatchDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *BatchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BatchDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchID := data.BatchID.ValueString()

	url := fmt.Sprintf("%s/batches/%s", d.client.OpenAIClient.APIURL, batchID)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+d.client.OpenAIClient.APIKey)
	if d.client.OpenAIClient.OrganizationID != "" {
		httpReq.Header.Set("OpenAI-Organization", d.client.OpenAIClient.OrganizationID)
	}
	if !data.ProjectID.IsNull() {
		httpReq.Header.Set("OpenAI-Project", data.ProjectID.ValueString())
	}

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s, Body: %s", httpResp.Status, string(bodyBytes)))
		return
	}

	var batchResp BatchResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&batchResp); err != nil {
		resp.Diagnostics.AddError("Error decoding response", err.Error())
		return
	}

	data.ID = types.StringValue(batchResp.ID)
	data.InputFileID = types.StringValue(batchResp.InputFileID)
	data.Endpoint = types.StringValue(batchResp.Endpoint)
	data.CompletionWindow = types.StringValue(batchResp.CompletionWindow)
	data.Status = types.StringValue(batchResp.Status)
	data.CreatedAt = types.Int64Value(batchResp.CreatedAt)
	data.ExpiresAt = types.Int64Value(batchResp.ExpiresAt)

	if batchResp.OutputFileID != "" {
		data.OutputFileID = types.StringValue(batchResp.OutputFileID)
	}
	if batchResp.ErrorFileID != "" {
		data.ErrorFileID = types.StringValue(batchResp.ErrorFileID)
	}

	if batchResp.InProgressAt != nil {
		data.InProgressAt = types.Int64Value(*batchResp.InProgressAt)
	}
	if batchResp.CompletedAt != nil {
		data.CompletedAt = types.Int64Value(*batchResp.CompletedAt)
	}

	if batchResp.RequestCounts != nil {
		rc := map[string]int64{
			"total":     int64(batchResp.RequestCounts.Total),
			"completed": int64(batchResp.RequestCounts.Completed),
			"failed":    int64(batchResp.RequestCounts.Failed),
		}
		data.RequestCounts, _ = types.MapValueFrom(ctx, types.Int64Type, rc)
	}

	if batchResp.Errors != nil {
		errorStr, err := json.Marshal(batchResp.Errors)
		if err == nil {
			data.Error = types.StringValue(string(errorStr))
		}
	}

	if len(batchResp.Metadata) > 0 {
		meta := make(map[string]string)
		for k, v := range batchResp.Metadata {
			meta[k] = fmt.Sprintf("%v", v)
		}
		data.Metadata, _ = types.MapValueFrom(ctx, types.StringType, meta)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Batches (Plural) ---

func NewBatchesDataSource() datasource.DataSource {
	return &BatchesDataSource{}
}

type BatchesDataSource struct {
	client *OpenAIClient
}

type BatchesDataSourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	Batches   types.List   `tfsdk:"batches"` // List<BatchResultModel>
	ID        types.String `tfsdk:"id"`
}

type BatchResultModel struct {
	ID               types.String `tfsdk:"id"`
	InputFileID      types.String `tfsdk:"input_file_id"`
	Endpoint         types.String `tfsdk:"endpoint"`
	CompletionWindow types.String `tfsdk:"completion_window"`
	OutputFileID     types.String `tfsdk:"output_file_id"`
	ErrorFileID      types.String `tfsdk:"error_file_id"`
	Status           types.String `tfsdk:"status"`
	CreatedAt        types.Int64  `tfsdk:"created_at"`
	InProgressAt     types.Int64  `tfsdk:"in_progress_at"`
	ExpiresAt        types.Int64  `tfsdk:"expires_at"`
	CompletedAt      types.Int64  `tfsdk:"completed_at"`
	RequestCounts    types.Map    `tfsdk:"request_counts"`
	Metadata         types.Map    `tfsdk:"metadata"`
}

// BatchesListResponse represents the API response for listing batches
type BatchesListResponse struct {
	Object  string          `json:"object"`
	Data    []BatchResponse `json:"data"`
	HasMore bool            `json:"has_more"`
}

func (d *BatchesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_batches"
}

func (d *BatchesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of batch jobs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project associated with the batch jobs. If not specified, the API key's default project will be used.",
				Optional:    true,
			},
			"batches": schema.ListNestedAttribute{
				Description: "List of batch jobs.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.StringAttribute{Computed: true},
						"input_file_id":     schema.StringAttribute{Computed: true},
						"endpoint":          schema.StringAttribute{Computed: true},
						"completion_window": schema.StringAttribute{Computed: true},
						"output_file_id":    schema.StringAttribute{Computed: true},
						"error_file_id":     schema.StringAttribute{Computed: true},
						"status":            schema.StringAttribute{Computed: true},
						"created_at":        schema.Int64Attribute{Computed: true},
						"in_progress_at":    schema.Int64Attribute{Computed: true},
						"expires_at":        schema.Int64Attribute{Computed: true},
						"completed_at":      schema.Int64Attribute{Computed: true},
						"request_counts": schema.MapAttribute{
							Computed:    true,
							ElementType: types.Int64Type,
						},
						"metadata": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *BatchesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *BatchesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BatchesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/batches", d.client.OpenAIClient.APIURL)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+d.client.OpenAIClient.APIKey)
	if d.client.OpenAIClient.OrganizationID != "" {
		httpReq.Header.Set("OpenAI-Organization", d.client.OpenAIClient.OrganizationID)
	}
	projectID := data.ProjectID.ValueString()
	if projectID != "" {
		httpReq.Header.Set("OpenAI-Project", projectID)
	}

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		// Check for permission error
		if strings.Contains(string(bodyBytes), "insufficient permissions") {
			tflog.Info(ctx, fmt.Sprintf("Permission error reading batches: %s", string(bodyBytes)))
			resp.Diagnostics.AddError("Permission Error", "You have insufficient permissions for this operation")
			return
		}
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s, Body: %s", httpResp.Status, string(bodyBytes)))
		return
	}

	var listResp BatchesListResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
		resp.Diagnostics.AddError("Error decoding response", err.Error())
		return
	}

	var batches []BatchResultModel
	for _, batch := range listResp.Data {
		b := BatchResultModel{
			ID:               types.StringValue(batch.ID),
			InputFileID:      types.StringValue(batch.InputFileID),
			Endpoint:         types.StringValue(batch.Endpoint),
			CompletionWindow: types.StringValue(batch.CompletionWindow),
			Status:           types.StringValue(batch.Status),
			CreatedAt:        types.Int64Value(batch.CreatedAt),
			ExpiresAt:        types.Int64Value(batch.ExpiresAt),
		}

		if batch.OutputFileID != "" {
			b.OutputFileID = types.StringValue(batch.OutputFileID)
		}
		if batch.ErrorFileID != "" {
			b.ErrorFileID = types.StringValue(batch.ErrorFileID)
		}
		if batch.InProgressAt != nil {
			b.InProgressAt = types.Int64Value(*batch.InProgressAt)
		}
		if batch.CompletedAt != nil {
			b.CompletedAt = types.Int64Value(*batch.CompletedAt)
		}
		if batch.RequestCounts != nil {
			rc := map[string]int64{
				"total":     int64(batch.RequestCounts.Total),
				"completed": int64(batch.RequestCounts.Completed),
				"failed":    int64(batch.RequestCounts.Failed),
			}
			b.RequestCounts, _ = types.MapValueFrom(ctx, types.Int64Type, rc)
		}
		if len(batch.Metadata) > 0 {
			meta := make(map[string]string)
			for k, v := range batch.Metadata {
				meta[k] = fmt.Sprintf("%v", v)
			}
			b.Metadata, _ = types.MapValueFrom(ctx, types.StringType, meta)
		}

		batches = append(batches, b)
	}

	data.ID = types.StringValue("batches")
	if projectID != "" {
		data.ID = types.StringValue(fmt.Sprintf("batches-%s", projectID))
	}

	data.Batches, _ = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":                types.StringType,
			"input_file_id":     types.StringType,
			"endpoint":          types.StringType,
			"completion_window": types.StringType,
			"output_file_id":    types.StringType,
			"error_file_id":     types.StringType,
			"status":            types.StringType,
			"created_at":        types.Int64Type,
			"in_progress_at":    types.Int64Type,
			"expires_at":        types.Int64Type,
			"completed_at":      types.Int64Type,
			"request_counts":    types.MapType{ElemType: types.Int64Type},
			"metadata":          types.MapType{ElemType: types.StringType},
		},
	}, batches)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
