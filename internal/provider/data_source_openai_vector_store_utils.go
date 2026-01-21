package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies interface
var _ datasource.DataSource = &VectorStoreFileDataSource{}
var _ datasource.DataSource = &VectorStoreFileBatchDataSource{}
var _ datasource.DataSource = &VectorStoreFileContentDataSource{}
var _ datasource.DataSource = &VectorStoreFilesDataSource{}
var _ datasource.DataSource = &VectorStoreFileBatchFilesDataSource{}

// --- Vector Store File ---

func NewVectorStoreFileDataSource() datasource.DataSource {
	return &VectorStoreFileDataSource{}
}

type VectorStoreFileDataSource struct {
	client *OpenAIClient
}

type VectorStoreFileDataSourceModel struct {
	VectorStoreID types.String `tfsdk:"vector_store_id"`
	FileID        types.String `tfsdk:"file_id"`
	ID            types.String `tfsdk:"id"`
	Object        types.String `tfsdk:"object"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	Status        types.String `tfsdk:"status"`
	Attributes    types.List   `tfsdk:"attributes"`
}

func (d *VectorStoreFileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_store_file"
}

func (d *VectorStoreFileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving details of a file in a vector store.",
		Attributes: map[string]schema.Attribute{
			"vector_store_id": schema.StringAttribute{Required: true},
			"file_id":         schema.StringAttribute{Required: true},
			"id":              schema.StringAttribute{Computed: true},
			"object":          schema.StringAttribute{Computed: true},
			"created_at":      schema.Int64Attribute{Computed: true},
			"status":          schema.StringAttribute{Computed: true},
			"attributes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"size":     schema.Int64Attribute{Computed: true},
						"filename": schema.StringAttribute{Computed: true},
						"purpose":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *VectorStoreFileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VectorStoreFileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VectorStoreFileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("/v1/vector_stores/%s/files/%s", data.VectorStoreID.ValueString(), data.FileID.ValueString())
	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading vector store file", err.Error())
		return
	}

	var file map[string]interface{}
	if err := json.Unmarshal(respBody, &file); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	if id, ok := file["id"].(string); ok {
		data.ID = types.StringValue(id)
	}
	if v, ok := file["object"].(string); ok {
		data.Object = types.StringValue(v)
	}
	if v, ok := file["created_at"].(float64); ok {
		data.CreatedAt = types.Int64Value(int64(v))
	}
	if v, ok := file["status"].(string); ok {
		data.Status = types.StringValue(v)
	}

	// Attributes (replicate legacy behavior)
	attrList := []attr.Value{}
	if attributes, ok := file["attributes"].(map[string]interface{}); ok {
		attrMap := map[string]attr.Value{
			"size":     types.Int64Null(),
			"filename": types.StringNull(),
			"purpose":  types.StringNull(),
		}
		if v, ok := attributes["size"].(float64); ok {
			attrMap["size"] = types.Int64Value(int64(v))
		}
		if v, ok := attributes["filename"].(string); ok {
			attrMap["filename"] = types.StringValue(v)
		}
		if v, ok := attributes["purpose"].(string); ok {
			attrMap["purpose"] = types.StringValue(v)
		}

		attrType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"size":     types.Int64Type,
				"filename": types.StringType,
				"purpose":  types.StringType,
			},
		}
		obj, _ := types.ObjectValue(attrType.AttrTypes, attrMap)
		attrList = append(attrList, obj)
	}

	attrType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":     types.Int64Type,
			"filename": types.StringType,
			"purpose":  types.StringType,
		},
	}
	data.Attributes, _ = types.ListValue(attrType, attrList)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Vector Store File Batch ---

func NewVectorStoreFileBatchDataSource() datasource.DataSource {
	return &VectorStoreFileBatchDataSource{}
}

type VectorStoreFileBatchDataSource struct {
	client *OpenAIClient
}

type VectorStoreFileBatchDataSourceModel struct {
	VectorStoreID types.String `tfsdk:"vector_store_id"`
	BatchID       types.String `tfsdk:"batch_id"`
	ID            types.String `tfsdk:"id"`
	Object        types.String `tfsdk:"object"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	Status        types.String `tfsdk:"status"`
	FileIDs       types.List   `tfsdk:"file_ids"`
	BatchType     types.String `tfsdk:"batch_type"` // Likely computed not in schema explicitly in legacy? check schema.
	Purpose       types.String `tfsdk:"purpose"`
}

func (d *VectorStoreFileBatchDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_store_file_batch"
}

func (d *VectorStoreFileBatchDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving details of a vector store file batch.",
		Attributes: map[string]schema.Attribute{
			"vector_store_id": schema.StringAttribute{Required: true},
			"batch_id":        schema.StringAttribute{Required: true},
			"id":              schema.StringAttribute{Computed: true},
			"object":          schema.StringAttribute{Computed: true},
			"created_at":      schema.Int64Attribute{Computed: true},
			"status":          schema.StringAttribute{Computed: true},
			"file_ids":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"batch_type":      schema.StringAttribute{Computed: true},
			"purpose":         schema.StringAttribute{Computed: true},
		},
	}
}

func (d *VectorStoreFileBatchDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VectorStoreFileBatchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VectorStoreFileBatchDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("/v1/vector_stores/%s/file_batches/%s", data.VectorStoreID.ValueString(), data.BatchID.ValueString())
	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading vector store file batch", err.Error())
		return
	}

	var batch map[string]interface{}
	if err := json.Unmarshal(respBody, &batch); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	if v, ok := batch["id"].(string); ok {
		data.ID = types.StringValue(v)
	}
	if v, ok := batch["object"].(string); ok {
		data.Object = types.StringValue(v)
	}
	if v, ok := batch["status"].(string); ok {
		data.Status = types.StringValue(v)
	}
	if v, ok := batch["created_at"].(float64); ok {
		data.CreatedAt = types.Int64Value(int64(v))
	}

	// Check for batch_type and purpose which legacy code reads but might not exist
	if v, ok := batch["batch_type"].(string); ok {
		data.BatchType = types.StringValue(v)
	}
	if v, ok := batch["purpose"].(string); ok {
		data.Purpose = types.StringValue(v)
	}

	// File IDs
	if fileIDs, ok := batch["file_ids"].([]interface{}); ok {
		ids := []attr.Value{}
		for _, id := range fileIDs {
			if strID, ok := id.(string); ok {
				ids = append(ids, types.StringValue(strID))
			}
		}
		data.FileIDs, _ = types.ListValue(types.StringType, ids)
	} else {
		data.FileIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Vector Store File Content ---

func NewVectorStoreFileContentDataSource() datasource.DataSource {
	return &VectorStoreFileContentDataSource{}
}

type VectorStoreFileContentDataSource struct {
	client *OpenAIClient
}

type VectorStoreFileContentDataSourceModel struct {
	VectorStoreID types.String `tfsdk:"vector_store_id"`
	FileID        types.String `tfsdk:"file_id"`
	Content       types.String `tfsdk:"content"`
	ID            types.String `tfsdk:"id"` // Dummy ID
}

func (d *VectorStoreFileContentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_store_file_content"
}

func (d *VectorStoreFileContentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving content of a file in a vector store.",
		Attributes: map[string]schema.Attribute{
			"vector_store_id": schema.StringAttribute{Required: true},
			"file_id":         schema.StringAttribute{Required: true},
			"content":         schema.StringAttribute{Computed: true},
			"id":              schema.StringAttribute{Computed: true},
		},
	}
}

func (d *VectorStoreFileContentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VectorStoreFileContentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VectorStoreFileContentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("/v1/vector_stores/%s/files/%s/content", data.VectorStoreID.ValueString(), data.FileID.ValueString())
	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading vector store file content", err.Error())
		return
	}

	data.Content = types.StringValue(string(respBody))
	data.ID = types.StringValue(fmt.Sprintf("%s_%s_content", data.VectorStoreID.ValueString(), data.FileID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Vector Store Files (Plural) ---

func NewVectorStoreFilesDataSource() datasource.DataSource {
	return &VectorStoreFilesDataSource{}
}

type VectorStoreFilesDataSource struct {
	client *OpenAIClient
}

type VectorStoreFilesDataSourceModel struct {
	VectorStoreID types.String `tfsdk:"vector_store_id"`
	Limit         types.Int64  `tfsdk:"limit"`
	Order         types.String `tfsdk:"order"`
	After         types.String `tfsdk:"after"`
	Before        types.String `tfsdk:"before"`
	Filter        types.String `tfsdk:"filter"`
	HasMore       types.Bool   `tfsdk:"has_more"`
	Files         types.List   `tfsdk:"files"`
	ID            types.String `tfsdk:"id"`
}

func (d *VectorStoreFilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_store_files"
}

func (d *VectorStoreFilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for listing files in a vector store.",
		Attributes: map[string]schema.Attribute{
			"vector_store_id": schema.StringAttribute{Required: true},
			"limit": schema.Int64Attribute{
				Optional:   true,
				Validators: []validator.Int64{int64validator.Between(1, 100)},
			},
			"order": schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf("asc", "desc")},
			},
			"after":  schema.StringAttribute{Optional: true},
			"before": schema.StringAttribute{Optional: true},
			"filter": schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf("in_progress", "completed", "failed", "cancelled")},
			},
			"has_more": schema.BoolAttribute{Computed: true},
			"id":       schema.StringAttribute{Computed: true},
			"files": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"object":     schema.StringAttribute{Computed: true},
						"created_at": schema.Int64Attribute{Computed: true},
						"status":     schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *VectorStoreFilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VectorStoreFilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VectorStoreFilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("/v1/vector_stores/%s/files", data.VectorStoreID.ValueString())
	params := []string{}
	if !data.Limit.IsNull() {
		params = append(params, fmt.Sprintf("limit=%d", data.Limit.ValueInt64()))
	}
	if !data.Order.IsNull() {
		params = append(params, fmt.Sprintf("order=%s", data.Order.ValueString()))
	}
	if !data.After.IsNull() {
		params = append(params, fmt.Sprintf("after=%s", data.After.ValueString()))
	}
	if !data.Before.IsNull() {
		params = append(params, fmt.Sprintf("before=%s", data.Before.ValueString()))
	}
	if !data.Filter.IsNull() {
		params = append(params, fmt.Sprintf("filter=%s", data.Filter.ValueString()))
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing vector store files", err.Error())
		return
	}

	var listResp struct {
		Data    []map[string]interface{} `json:"data"`
		HasMore bool                     `json:"has_more"`
	}
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.HasMore = types.BoolValue(listResp.HasMore)

	filesList := []attr.Value{}
	fileType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":         types.StringType,
			"object":     types.StringType,
			"created_at": types.Int64Type,
			"status":     types.StringType,
		},
	}

	for _, item := range listResp.Data {
		attrs := map[string]attr.Value{
			"id":         types.StringNull(),
			"object":     types.StringNull(),
			"created_at": types.Int64Null(),
			"status":     types.StringNull(),
		}
		if v, ok := item["id"].(string); ok {
			attrs["id"] = types.StringValue(v)
		}
		if v, ok := item["object"].(string); ok {
			attrs["object"] = types.StringValue(v)
		}
		if v, ok := item["created_at"].(float64); ok {
			attrs["created_at"] = types.Int64Value(int64(v))
		}
		if v, ok := item["status"].(string); ok {
			attrs["status"] = types.StringValue(v)
		}

		obj, _ := types.ObjectValue(fileType.AttrTypes, attrs)
		filesList = append(filesList, obj)
	}
	data.Files, _ = types.ListValue(fileType, filesList)
	data.ID = types.StringValue(fmt.Sprintf("vector_store_files_%d", time.Now().Unix()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Vector Store File Batch Files (Plural) ---

func NewVectorStoreFileBatchFilesDataSource() datasource.DataSource {
	return &VectorStoreFileBatchFilesDataSource{}
}

type VectorStoreFileBatchFilesDataSource struct {
	client *OpenAIClient
}

type VectorStoreFileBatchFilesDataSourceModel struct {
	VectorStoreID types.String `tfsdk:"vector_store_id"`
	BatchID       types.String `tfsdk:"batch_id"`
	Limit         types.Int64  `tfsdk:"limit"`
	Order         types.String `tfsdk:"order"`
	After         types.String `tfsdk:"after"`
	Before        types.String `tfsdk:"before"`
	Filter        types.String `tfsdk:"filter"`
	HasMore       types.Bool   `tfsdk:"has_more"`
	Files         types.List   `tfsdk:"files"`
	ID            types.String `tfsdk:"id"`
}

func (d *VectorStoreFileBatchFilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vector_store_file_batch_files"
}

func (d *VectorStoreFileBatchFilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for listing files in a vector store file batch.",
		Attributes: map[string]schema.Attribute{
			"vector_store_id": schema.StringAttribute{Required: true},
			"batch_id":        schema.StringAttribute{Required: true},
			"limit": schema.Int64Attribute{
				Optional:   true,
				Validators: []validator.Int64{int64validator.Between(1, 100)},
			},
			"order": schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf("asc", "desc")},
			},
			"after":  schema.StringAttribute{Optional: true},
			"before": schema.StringAttribute{Optional: true},
			"filter": schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf("in_progress", "completed", "failed", "cancelled")},
			},
			"has_more": schema.BoolAttribute{Computed: true},
			"id":       schema.StringAttribute{Computed: true},
			"files": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"object":     schema.StringAttribute{Computed: true},
						"created_at": schema.Int64Attribute{Computed: true},
						"status":     schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *VectorStoreFileBatchFilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VectorStoreFileBatchFilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VectorStoreFileBatchFilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("/v1/vector_stores/%s/file_batches/%s/files", data.VectorStoreID.ValueString(), data.BatchID.ValueString())
	params := []string{}
	if !data.Limit.IsNull() {
		params = append(params, fmt.Sprintf("limit=%d", data.Limit.ValueInt64()))
	}
	if !data.Order.IsNull() {
		params = append(params, fmt.Sprintf("order=%s", data.Order.ValueString()))
	}
	if !data.After.IsNull() {
		params = append(params, fmt.Sprintf("after=%s", data.After.ValueString()))
	}
	if !data.Before.IsNull() {
		params = append(params, fmt.Sprintf("before=%s", data.Before.ValueString()))
	}
	if !data.Filter.IsNull() {
		params = append(params, fmt.Sprintf("filter=%s", data.Filter.ValueString()))
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	respBody, err := d.client.DoRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing vector store file batch files", err.Error())
		return
	}

	var listResp struct {
		Data    []map[string]interface{} `json:"data"`
		HasMore bool                     `json:"has_more"`
	}
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.HasMore = types.BoolValue(listResp.HasMore)

	filesList := []attr.Value{}
	fileType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":         types.StringType,
			"object":     types.StringType,
			"created_at": types.Int64Type,
			"status":     types.StringType,
		},
	}

	for _, item := range listResp.Data {
		attrs := map[string]attr.Value{
			"id":         types.StringNull(),
			"object":     types.StringNull(),
			"created_at": types.Int64Null(),
			"status":     types.StringNull(),
		}
		if v, ok := item["id"].(string); ok {
			attrs["id"] = types.StringValue(v)
		}
		if v, ok := item["object"].(string); ok {
			attrs["object"] = types.StringValue(v)
		}
		if v, ok := item["created_at"].(float64); ok {
			attrs["created_at"] = types.Int64Value(int64(v))
		}
		if v, ok := item["status"].(string); ok {
			attrs["status"] = types.StringValue(v)
		}

		obj, _ := types.ObjectValue(fileType.AttrTypes, attrs)
		filesList = append(filesList, obj)
	}
	data.Files, _ = types.ListValue(fileType, filesList)
	data.ID = types.StringValue(fmt.Sprintf("vector_store_file_batch_files_%d", time.Now().Unix()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
