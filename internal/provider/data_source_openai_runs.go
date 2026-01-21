package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RunsDataSource{}

func NewRunsDataSource() datasource.DataSource {
	return &RunsDataSource{}
}

type RunsDataSource struct {
	client *OpenAIClient
}

type RunsDataSourceModel struct {
	ThreadID types.String     `tfsdk:"thread_id"`
	Result   []RunResultModel `tfsdk:"runs"`
	ID       types.String     `tfsdk:"id"` // Dummy ID
}

// RunResultModel mirrors RunDataSourceModel but for use in a list
type RunResultModel struct {
	ID                  types.String             `tfsdk:"id"`
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

func (d *RunsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runs"
}

func (d *RunsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Reusing attributes from RunDataSource for the nested list

	runNestedAttributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"thread_id": schema.StringAttribute{
			Computed: true,
		},
		"object": schema.StringAttribute{
			Computed: true,
		},
		"created_at": schema.Int64Attribute{
			Computed: true,
		},
		"assistant_id": schema.StringAttribute{
			Computed: true,
		},
		"status": schema.StringAttribute{
			Computed: true,
		},
		"started_at": schema.Int64Attribute{
			Computed: true,
		},
		"completed_at": schema.Int64Attribute{
			Computed: true,
		},
		"model": schema.StringAttribute{
			Computed: true,
		},
		"instructions": schema.StringAttribute{
			Computed: true,
		},
		"metadata": schema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"usage": schema.SingleNestedAttribute{
			Computed: true,
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
	}

	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of runs for a specific OpenAI thread.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"thread_id": schema.StringAttribute{
				Description: "The ID of the thread to list runs for.",
				Required:    true,
			},
			"runs": schema.ListNestedAttribute{
				Description: "List of runs.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: runNestedAttributes,
				},
			},
		},
	}
}

func (d *RunsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RunsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RunsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	threadID := data.ThreadID.ValueString()
	apiClient := d.client.OpenAIClient
	var allRuns []RunResultModel
	cursor := ""

	for {
		queryParams := url.Values{}
		queryParams.Set("limit", "100")
		if cursor != "" {
			queryParams.Set("after", cursor)
		}

		path := fmt.Sprintf("threads/%s/runs?%s", threadID, queryParams.Encode())

		respBody, err := apiClient.DoRequest(http.MethodGet, path, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Runs",
				fmt.Sprintf("Could not list runs for thread %s: %s", threadID, err.Error()),
			)
			return
		}

		var listResp ListRunsResponse
		if err := json.Unmarshal(respBody, &listResp); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Runs Response",
				fmt.Sprintf("Could not parse response: %s", err.Error()),
			)
			return
		}

		for _, runResponse := range listResp.Data {
			runModel := RunResultModel{
				ID:           types.StringValue(runResponse.ID),
				ThreadID:     types.StringValue(runResponse.ThreadID),
				Object:       types.StringValue(runResponse.Object),
				CreatedAt:    types.Int64Value(runResponse.CreatedAt),
				AssistantID:  types.StringValue(runResponse.AssistantID),
				Status:       types.StringValue(runResponse.Status),
				Model:        types.StringValue(runResponse.Model),
				Instructions: types.StringValue(runResponse.Instructions),
			}

			if runResponse.StartedAt != nil {
				runModel.StartedAt = types.Int64Value(*runResponse.StartedAt)
			} else {
				runModel.StartedAt = types.Int64Null()
			}

			if runResponse.CompletedAt != nil {
				runModel.CompletedAt = types.Int64Value(*runResponse.CompletedAt)
			} else {
				runModel.CompletedAt = types.Int64Null()
			}

			if runResponse.ExpiresAt != nil {
				runModel.ExpiresAt = types.Int64Value(*runResponse.ExpiresAt)
			} else {
				runModel.ExpiresAt = types.Int64Null()
			}

			if runResponse.FailedAt != nil {
				runModel.FailedAt = types.Int64Value(*runResponse.FailedAt)
			} else {
				runModel.FailedAt = types.Int64Null()
			}

			if runResponse.CancelledAt != nil {
				runModel.CancelledAt = types.Int64Value(*runResponse.CancelledAt)
			} else {
				runModel.CancelledAt = types.Int64Null()
			}

			// Map Metadata
			if len(runResponse.Metadata) > 0 {
				metadataVals := make(map[string]attr.Value)
				for k, v := range runResponse.Metadata {
					metadataVals[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
				runModel.Metadata, _ = types.MapValue(types.StringType, metadataVals)
			} else {
				runModel.Metadata = types.MapNull(types.StringType)
			}

			// Map Usage
			if runResponse.Usage != nil {
				runModel.Usage = &RunDataSourceUsageModel{
					PromptTokens:     types.Int64Value(int64(runResponse.Usage.PromptTokens)),
					CompletionTokens: types.Int64Value(int64(runResponse.Usage.CompletionTokens)),
					TotalTokens:      types.Int64Value(int64(runResponse.Usage.TotalTokens)),
				}
			} else {
				runModel.Usage = nil
			}

			if runResponse.Temperature != nil {
				runModel.Temperature = types.Float64Value(*runResponse.Temperature)
			} else {
				runModel.Temperature = types.Float64Null()
			}

			if runResponse.TopP != nil {
				runModel.TopP = types.Float64Value(*runResponse.TopP)
			} else {
				runModel.TopP = types.Float64Null()
			}

			if runResponse.MaxPromptTokens != nil {
				runModel.MaxPromptTokens = types.Int64Value(int64(*runResponse.MaxPromptTokens))
			} else {
				runModel.MaxPromptTokens = types.Int64Null()
			}

			if runResponse.MaxCompletionTokens != nil {
				runModel.MaxCompletionTokens = types.Int64Value(int64(*runResponse.MaxCompletionTokens))
			} else {
				runModel.MaxCompletionTokens = types.Int64Null()
			}

			if runResponse.ResponseFormat != nil {
				rfStr := fmt.Sprintf("%v", runResponse.ResponseFormat)
				runModel.ResponseFormat = types.StringValue(rfStr)
			} else {
				runModel.ResponseFormat = types.StringNull()
			}

			allRuns = append(allRuns, runModel)
		}

		if !listResp.HasMore {
			break
		}
		cursor = listResp.LastID
	}

	data.ID = types.StringValue(threadID)
	data.Result = allRuns

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
