package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ThreadRunDataSource{}

func NewThreadRunDataSource() datasource.DataSource {
	return &ThreadRunDataSource{}
}

type ThreadRunDataSource struct {
	client *OpenAIClient
}

// ThreadRunDataSourceModel mirrors RunDataSourceModel, effectively an alias for Run but specific to Thread
type ThreadRunDataSourceModel struct {
	ID                  types.String             `tfsdk:"id"`
	RunID               types.String             `tfsdk:"run_id"`
	ThreadID            types.String             `tfsdk:"thread_id"`
	Object              types.String             `tfsdk:"object"`
	CreatedAt           types.Int64              `tfsdk:"created_at"`
	AssistantID         types.String             `tfsdk:"assistant_id"`
	Status              types.String             `tfsdk:"status"`
	StartedAt           types.Int64              `tfsdk:"started_at"`
	CompletedAt         types.Int64              `tfsdk:"completed_at"`
	Model               types.String             `tfsdk:"model"`
	Instructions        types.String             `tfsdk:"instructions"`
	Metadata            types.Map                `tfsdk:"metadata"`
	Usage               *RunDataSourceUsageModel `tfsdk:"usage"`
	Temperature         types.Float64            `tfsdk:"temperature"`
	TopP                types.Float64            `tfsdk:"top_p"`
	MaxPromptTokens     types.Int64              `tfsdk:"max_prompt_tokens"`
	MaxCompletionTokens types.Int64              `tfsdk:"max_completion_tokens"`
	ExpiresAt           types.Int64              `tfsdk:"expires_at"`
	FailedAt            types.Int64              `tfsdk:"failed_at"`
	CancelledAt         types.Int64              `tfsdk:"cancelled_at"`
	ResponseFormat      types.String             `tfsdk:"response_format"`
}

func (d *ThreadRunDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thread_run"
}

func (d *ThreadRunDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific OpenAI run within a thread.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"run_id": schema.StringAttribute{
				Description: "The ID of the run.",
				Required:    true,
			},
			"thread_id": schema.StringAttribute{
				Description: "The ID of the thread the run belongs to.",
				Required:    true,
			},
			"object": schema.StringAttribute{
				Description: "The object type, which is always 'thread.run'.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the run was created.",
				Computed:    true,
			},
			"assistant_id": schema.StringAttribute{
				Description: "The ID of the assistant used for execution of this run.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the run.",
				Computed:    true,
			},
			"started_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the run was started.",
				Computed:    true,
			},
			"completed_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) for when the run was completed.",
				Computed:    true,
			},
			"model": schema.StringAttribute{
				Description: "The model that the assistant used for this run.",
				Computed:    true,
			},
			"instructions": schema.StringAttribute{
				Description: "The instructions that the assistant used for this run.",
				Computed:    true,
			},
			"metadata": schema.MapAttribute{
				Description: "Set of key-value pairs that can be attached to an object.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"usage": schema.SingleNestedAttribute{
				Description: "Usage statistics related to the run.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"prompt_tokens": schema.Int64Attribute{
						Computed: true,
					},
					"completion_tokens": schema.Int64Attribute{
						Computed: true,
					},
					"total_tokens": schema.Int64Attribute{
						Computed: true,
					},
				},
			},
			"temperature": schema.Float64Attribute{
				Computed: true,
			},
			"top_p": schema.Float64Attribute{
				Computed: true,
			},
			"max_prompt_tokens": schema.Int64Attribute{
				Computed: true,
			},
			"max_completion_tokens": schema.Int64Attribute{
				Computed: true,
			},
			"expires_at": schema.Int64Attribute{
				Computed: true,
			},
			"failed_at": schema.Int64Attribute{
				Computed: true,
			},
			"cancelled_at": schema.Int64Attribute{
				Computed: true,
			},
			"response_format": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ThreadRunDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ThreadRunDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ThreadRunDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	threadID := data.ThreadID.ValueString()
	runID := data.RunID.ValueString()
	path := fmt.Sprintf("threads/%s/runs/%s", threadID, runID)

	apiClient := d.client.OpenAIClient

	respBody, err := apiClient.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Run",
			fmt.Sprintf("Could not read run with ID %s in thread %s: %s", runID, threadID, err.Error()),
		)
		return
	}

	var runResponse RunResponse
	if err := json.Unmarshal(respBody, &runResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Run Response",
			fmt.Sprintf("Could not parse response for run %s: %s", runID, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(runResponse.ID)
	data.RunID = types.StringValue(runResponse.ID)
	data.ThreadID = types.StringValue(runResponse.ThreadID)
	data.Object = types.StringValue(runResponse.Object)
	data.CreatedAt = types.Int64Value(runResponse.CreatedAt)
	data.AssistantID = types.StringValue(runResponse.AssistantID)
	data.Status = types.StringValue(runResponse.Status)
	data.Model = types.StringValue(runResponse.Model)
	data.Instructions = types.StringValue(runResponse.Instructions)

	if runResponse.StartedAt != nil {
		data.StartedAt = types.Int64Value(*runResponse.StartedAt)
	} else {
		data.StartedAt = types.Int64Null()
	}

	if runResponse.CompletedAt != nil {
		data.CompletedAt = types.Int64Value(*runResponse.CompletedAt)
	} else {
		data.CompletedAt = types.Int64Null()
	}

	if runResponse.ExpiresAt != nil {
		data.ExpiresAt = types.Int64Value(*runResponse.ExpiresAt)
	} else {
		data.ExpiresAt = types.Int64Null()
	}

	if runResponse.FailedAt != nil {
		data.FailedAt = types.Int64Value(*runResponse.FailedAt)
	} else {
		data.FailedAt = types.Int64Null()
	}

	if runResponse.CancelledAt != nil {
		data.CancelledAt = types.Int64Value(*runResponse.CancelledAt)
	} else {
		data.CancelledAt = types.Int64Null()
	}

	// Map Metadata
	if len(runResponse.Metadata) > 0 {
		metadataVals := make(map[string]attr.Value)
		for k, v := range runResponse.Metadata {
			metadataVals[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		data.Metadata, _ = types.MapValue(types.StringType, metadataVals)
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	// Map Usage
	if runResponse.Usage != nil {
		data.Usage = &RunDataSourceUsageModel{
			PromptTokens:     types.Int64Value(int64(runResponse.Usage.PromptTokens)),
			CompletionTokens: types.Int64Value(int64(runResponse.Usage.CompletionTokens)),
			TotalTokens:      types.Int64Value(int64(runResponse.Usage.TotalTokens)),
		}
	} else {
		data.Usage = nil
	}

	if runResponse.Temperature != nil {
		data.Temperature = types.Float64Value(*runResponse.Temperature)
	} else {
		data.Temperature = types.Float64Null()
	}

	if runResponse.TopP != nil {
		data.TopP = types.Float64Value(*runResponse.TopP)
	} else {
		data.TopP = types.Float64Null()
	}

	if runResponse.MaxPromptTokens != nil {
		data.MaxPromptTokens = types.Int64Value(int64(*runResponse.MaxPromptTokens))
	} else {
		data.MaxPromptTokens = types.Int64Null()
	}

	if runResponse.MaxCompletionTokens != nil {
		data.MaxCompletionTokens = types.Int64Value(int64(*runResponse.MaxCompletionTokens))
	} else {
		data.MaxCompletionTokens = types.Int64Null()
	}

	if runResponse.ResponseFormat != nil {
		rfStr := fmt.Sprintf("%v", runResponse.ResponseFormat)
		data.ResponseFormat = types.StringValue(rfStr)
	} else {
		data.ResponseFormat = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
