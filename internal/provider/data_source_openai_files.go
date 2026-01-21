package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FilesDataSource{}

func NewFilesDataSource() datasource.DataSource {
	return &FilesDataSource{}
}

type FilesDataSource struct {
	client *OpenAIClient
}

func (d *FilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_files"
}

type FilesDataSourceModel struct {
	ID        types.String        `tfsdk:"id"`
	Files     []FileResponseModel `tfsdk:"files"`
	Purpose   types.String        `tfsdk:"purpose"`
	ProjectID types.String        `tfsdk:"project_id"`
}

type FileResponseModel struct {
	ID        types.String `tfsdk:"id"`
	Filename  types.String `tfsdk:"filename"`
	Bytes     types.Int64  `tfsdk:"bytes"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
	Purpose   types.String `tfsdk:"purpose"`
	Object    types.String `tfsdk:"object"`
}

func (d *FilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Files data source allows you to list files.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"purpose": schema.StringAttribute{
				Description: "Filter files by purpose (e.g., 'fine-tune', 'assistants', etc.)",
				Optional:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID (for Terraform logic only)",
				Optional:    true,
			},
			"files": schema.ListNestedAttribute{
				Description: "List of files.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the file.",
							Computed:    true,
						},
						"filename": schema.StringAttribute{
							Description: "The name of the file.",
							Computed:    true,
						},
						"bytes": schema.Int64Attribute{
							Description: "The size of the file in bytes.",
							Computed:    true,
						},
						"created_at": schema.Int64Attribute{
							Description: "The timestamp when the file was created.",
							Computed:    true,
						},
						"purpose": schema.StringAttribute{
							Description: "The purpose of the file.",
							Computed:    true,
						},
						"object": schema.StringAttribute{
							Description: "The object type.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *FilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *provider.OpenAIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *FilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := "files"
	if !strings.HasSuffix(d.client.OpenAIClient.APIURL, "/v1") && !strings.Contains(d.client.OpenAIClient.APIURL, "/v1") {
		// client usually handles base path relative to APIURL.
		// If APIURL is "https://api.openai.com/v1", path "files" -> "https://api.openai.com/v1/files".
		// This logic in client needs to be trusted.
	}

	// Query params
	if !data.Purpose.IsNull() {
		url = fmt.Sprintf("files?purpose=%s", data.Purpose.ValueString())
	}

	respBody, err := d.client.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing files", err.Error())
		return
	}

	var listResp ListFilesResponse
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		resp.Diagnostics.AddError("Error parsing files response", err.Error())
		return
	}

	files := []FileResponseModel{}
	for _, f := range listResp.Data {
		files = append(files, FileResponseModel{
			ID:        types.StringValue(f.ID),
			Filename:  types.StringValue(f.Filename),
			Bytes:     types.Int64Value(f.Bytes),
			CreatedAt: types.Int64Value(f.CreatedAt),
			Purpose:   types.StringValue(f.Purpose),
			Object:    types.StringValue(f.Object),
		})
	}

	data.Files = files
	data.ID = types.StringValue("files") // Consistent ID for data source

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
