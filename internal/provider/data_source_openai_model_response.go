package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies interface
var _ datasource.DataSource = &ModelResponseDataSource{}
var _ datasource.DataSource = &ModelResponsesDataSource{}
var _ datasource.DataSource = &ModelResponseInputItemsDataSource{}

// --- Model Response ---

func NewModelResponseDataSource() datasource.DataSource {
	return &ModelResponseDataSource{}
}

type ModelResponseDataSource struct {
	client *OpenAIClient
}

type ModelResponseDataSourceModel struct {
	ResponseID  types.String  `tfsdk:"response_id"`
	ID          types.String  `tfsdk:"id"`
	Include     types.List    `tfsdk:"include"`
	CreatedAt   types.Int64   `tfsdk:"created_at"`
	Status      types.String  `tfsdk:"status"`
	Model       types.String  `tfsdk:"model"`
	Temperature types.Float64 `tfsdk:"temperature"`
	TopP        types.Float64 `tfsdk:"top_p"`
	Output      types.Map     `tfsdk:"output"`
	Usage       types.Map     `tfsdk:"usage"`
	InputItems  types.List    `tfsdk:"input_items"`
}

func (d *ModelResponseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_response"
}

func (d *ModelResponseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving OpenAI model response information",
		Attributes: map[string]schema.Attribute{
			"response_id": schema.StringAttribute{Required: true},
			"id":          schema.StringAttribute{Computed: true},
			"include":     schema.ListAttribute{Optional: true, ElementType: types.StringType},
			"created_at":  schema.Int64Attribute{Computed: true},
			"status":      schema.StringAttribute{Computed: true},
			"model":       schema.StringAttribute{Computed: true},
			"temperature": schema.Float64Attribute{Computed: true},
			"top_p":       schema.Float64Attribute{Computed: true},
			"output":      schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"usage":       schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"input_items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type":    schema.StringAttribute{Computed: true},
						"id":      schema.StringAttribute{Computed: true},
						"role":    schema.StringAttribute{Computed: true},
						"content": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ModelResponseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *OpenAIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ModelResponseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ModelResponseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseID := data.ResponseID.ValueString()
	url := fmt.Sprintf("/v1/responses/%s", responseID)

	if !data.Include.IsNull() {
		params := []string{}
		include := []string{}
		data.Include.ElementsAs(ctx, &include, false)
		if len(include) > 0 {
			params = append(params, "include="+strings.Join(include, ","))
			url += "?" + strings.Join(params, "&")
		}
	}

	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error", err.Error())
		return
	}

	var modelResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &modelResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(responseID)

	if v, ok := modelResponse["created_at"].(float64); ok {
		data.CreatedAt = types.Int64Value(int64(v))
	}
	if v, ok := modelResponse["status"].(string); ok {
		data.Status = types.StringValue(v)
	}
	if v, ok := modelResponse["model"].(string); ok {
		data.Model = types.StringValue(v)
	}
	if v, ok := modelResponse["temperature"].(float64); ok {
		data.Temperature = types.Float64Value(v)
	}
	if v, ok := modelResponse["top_p"].(float64); ok {
		data.TopP = types.Float64Value(v)
	}

	// Output
	if output, ok := modelResponse["output"].([]interface{}); ok && len(output) > 0 {
		outMap := make(map[string]string)
		if first, ok := output[0].(map[string]interface{}); ok {
			if role, ok := first["role"].(string); ok {
				outMap["role"] = role
			}
			if content, ok := first["content"].([]interface{}); ok && len(content) > 0 {
				if c0, ok := content[0].(map[string]interface{}); ok {
					if text, ok := c0["text"].(string); ok {
						outMap["text"] = text
					}
				}
			}
		}
		data.Output, _ = types.MapValueFrom(ctx, types.StringType, outMap)
	}

	// Usage
	if usage, ok := modelResponse["usage"].(map[string]interface{}); ok {
		uMap := make(map[string]string)
		for k, v := range usage {
			uMap[k] = fmt.Sprintf("%v", v)
		}
		data.Usage, _ = types.MapValueFrom(ctx, types.StringType, uMap)
	}

	// Input Items (Legacy did a separate request!)
	// /v1/responses/{id}/input_items

	itemsURL := fmt.Sprintf("/v1/responses/%s/input_items", responseID)
	itemsBody, err := d.client.DoRequest("GET", itemsURL, nil)
	if err == nil {
		var itemsResp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(itemsBody, &itemsResp); err == nil {
			items := []attr.Value{}
			itemType := types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type":    types.StringType,
					"id":      types.StringType,
					"role":    types.StringType,
					"content": types.StringType,
				},
			}
			for _, item := range itemsResp.Data {
				attrs := map[string]attr.Value{
					"type":    types.StringNull(),
					"id":      types.StringNull(),
					"role":    types.StringNull(),
					"content": types.StringNull(),
				}
				if v, ok := item["type"].(string); ok {
					attrs["type"] = types.StringValue(v)
				}
				if v, ok := item["id"].(string); ok {
					attrs["id"] = types.StringValue(v)
				}
				if v, ok := item["role"].(string); ok {
					attrs["role"] = types.StringValue(v)
				}

				// Content
				if content, ok := item["content"].([]interface{}); ok && len(content) > 0 {
					if c0, ok := content[0].(map[string]interface{}); ok {
						if text, ok := c0["text"].(string); ok {
							attrs["content"] = types.StringValue(text)
						}
					}
				}
				obj, _ := types.ObjectValue(itemType.AttrTypes, attrs)
				items = append(items, obj)
			}
			data.InputItems, _ = types.ListValue(itemType, items)
		}
	} else {
		// Just warn or ignore? Legacy triggers separate request.
		// If error, maybe empty list.
		data.InputItems = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
			"type": types.StringType, "id": types.StringType, "role": types.StringType, "content": types.StringType,
		}})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Model Responses (Plural) ---

func NewModelResponsesDataSource() datasource.DataSource {
	return &ModelResponsesDataSource{}
}

type ModelResponsesDataSource struct {
	client *OpenAIClient
}

type ModelResponsesDataSourceModel struct {
	FilterByUser types.String `tfsdk:"filter_by_user"`
	Limit        types.Int64  `tfsdk:"limit"`
	Order        types.String `tfsdk:"order"`
	After        types.String `tfsdk:"after"`
	Before       types.String `tfsdk:"before"`
	Responses    types.List   `tfsdk:"responses"`
	HasMore      types.Bool   `tfsdk:"has_more"`
	ID           types.String `tfsdk:"id"`
}

func (d *ModelResponsesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_responses"
}

func (d *ModelResponsesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for listing OpenAI model responses",
		Attributes: map[string]schema.Attribute{
			"filter_by_user": schema.StringAttribute{Optional: true},
			"limit":          schema.Int64Attribute{Optional: true},
			"order":          schema.StringAttribute{Optional: true},
			"after":          schema.StringAttribute{Optional: true},
			"before":         schema.StringAttribute{Optional: true},
			"has_more":       schema.BoolAttribute{Computed: true},
			"id":             schema.StringAttribute{Computed: true},
			"responses": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"created_at": schema.Int64Attribute{Computed: true},
						"model":      schema.StringAttribute{Computed: true},
						"status":     schema.StringAttribute{Computed: true},
						"usage":      schema.MapAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

func (d *ModelResponsesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *OpenAIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ModelResponsesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ModelResponsesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := "/v1/responses"
	params := []string{}
	// Add params
	if !data.Limit.IsNull() {
		params = append(params, fmt.Sprintf("limit=%d", data.Limit.ValueInt64()))
	}
	if !data.Order.IsNull() {
		params = append(params, fmt.Sprintf("order=%s", data.Order.ValueString()))
	}
	if !data.After.IsNull() {
		params = append(params, fmt.Sprintf("after=%s", data.After.ValueString()))
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error", err.Error())
		return
	}

	var listResp struct {
		Data    []map[string]interface{} `json:"data"`
		HasMore bool                     `json:"has_more"`
	}
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		resp.Diagnostics.AddError("Error parsing", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("model_responses_%d", time.Now().Unix()))
	data.HasMore = types.BoolValue(listResp.HasMore)

	// Map responses
	resps := []attr.Value{}
	respType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":         types.StringType,
			"created_at": types.Int64Type,
			"model":      types.StringType,
			"status":     types.StringType,
			"usage":      types.MapType{ElemType: types.StringType},
		},
	}

	for _, r := range listResp.Data {
		attrs := map[string]attr.Value{
			"id":         types.StringNull(),
			"created_at": types.Int64Null(),
			"model":      types.StringNull(),
			"status":     types.StringNull(),
			"usage":      types.MapNull(types.StringType),
		}
		if v, ok := r["id"].(string); ok {
			attrs["id"] = types.StringValue(v)
		}
		if v, ok := r["created_at"].(float64); ok {
			attrs["created_at"] = types.Int64Value(int64(v))
		}
		if v, ok := r["model"].(string); ok {
			attrs["model"] = types.StringValue(v)
		}
		if v, ok := r["status"].(string); ok {
			attrs["status"] = types.StringValue(v)
		}
		if u, ok := r["usage"].(map[string]interface{}); ok {
			um := make(map[string]string)
			for k, v := range u {
				um[k] = fmt.Sprintf("%v", v)
			}
			attrs["usage"], _ = types.MapValueFrom(ctx, types.StringType, um)
		}
		obj, _ := types.ObjectValue(respType.AttrTypes, attrs)
		resps = append(resps, obj)
	}
	data.Responses, _ = types.ListValue(respType, resps)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Model Response Input Items ---

func NewModelResponseInputItemsDataSource() datasource.DataSource {
	return &ModelResponseInputItemsDataSource{}
}

type ModelResponseInputItemsDataSource struct {
	client *OpenAIClient
}

type ModelResponseInputItemsDataSourceModel struct {
	ResponseID types.String `tfsdk:"response_id"`
	After      types.String `tfsdk:"after"`
	Before     types.String `tfsdk:"before"`
	Limit      types.Int64  `tfsdk:"limit"`
	Order      types.String `tfsdk:"order"`
	Include    types.List   `tfsdk:"include"`
	InputItems types.List   `tfsdk:"input_items"`
	HasMore    types.Bool   `tfsdk:"has_more"`
	FirstID    types.String `tfsdk:"first_id"`
	LastID     types.String `tfsdk:"last_id"`
}

func (d *ModelResponseInputItemsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_response_input_items"
}

func (d *ModelResponseInputItemsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving input items for an OpenAI model response",
		Attributes: map[string]schema.Attribute{
			"response_id": schema.StringAttribute{Required: true},
			"after":       schema.StringAttribute{Optional: true},
			"before":      schema.StringAttribute{Optional: true},
			"limit":       schema.Int64Attribute{Optional: true},
			"order":       schema.StringAttribute{Optional: true},
			"include":     schema.ListAttribute{Optional: true, ElementType: types.StringType},
			"has_more":    schema.BoolAttribute{Computed: true},
			"first_id":    schema.StringAttribute{Computed: true},
			"last_id":     schema.StringAttribute{Computed: true},
			"input_items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type":    schema.StringAttribute{Computed: true},
						"id":      schema.StringAttribute{Computed: true},
						"role":    schema.StringAttribute{Computed: true},
						"content": schema.StringAttribute{Computed: true},
						"status":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ModelResponseInputItemsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *OpenAIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ModelResponseInputItemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Replicates read logic for input items list
	var data ModelResponseInputItemsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("/v1/responses/%s/input_items", data.ResponseID.ValueString())
	// Params...

	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error", err.Error())
		return
	}

	var listResp struct {
		Data    []map[string]interface{} `json:"data"`
		HasMore bool                     `json:"has_more"`
		FirstID string                   `json:"first_id"`
		LastID  string                   `json:"last_id"`
	}
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		resp.Diagnostics.AddError("Error parsing", err.Error())
		return
	}

	data.HasMore = types.BoolValue(listResp.HasMore)
	data.FirstID = types.StringValue(listResp.FirstID)
	data.LastID = types.StringValue(listResp.LastID)

	items := []attr.Value{}
	itemType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"type":    types.StringType,
		"id":      types.StringType,
		"role":    types.StringType,
		"content": types.StringType,
		"status":  types.StringType,
	}}

	for _, item := range listResp.Data {
		attrs := map[string]attr.Value{
			"type":    types.StringNull(),
			"id":      types.StringNull(),
			"role":    types.StringNull(),
			"content": types.StringNull(),
			"status":  types.StringNull(),
		}
		if v, ok := item["type"].(string); ok {
			attrs["type"] = types.StringValue(v)
		}
		if v, ok := item["id"].(string); ok {
			attrs["id"] = types.StringValue(v)
		}
		if v, ok := item["role"].(string); ok {
			attrs["role"] = types.StringValue(v)
		}
		if v, ok := item["status"].(string); ok {
			attrs["status"] = types.StringValue(v)
		}

		if content, ok := item["content"].([]interface{}); ok && len(content) > 0 {
			if c0, ok := content[0].(map[string]interface{}); ok {
				if text, ok := c0["text"].(string); ok {
					attrs["content"] = types.StringValue(text)
				}
			}
		}
		obj, _ := types.ObjectValue(itemType.AttrTypes, attrs)
		items = append(items, obj)
	}
	data.InputItems, _ = types.ListValue(itemType, items)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
