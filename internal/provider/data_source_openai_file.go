package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FileDataSource{}

func NewFileDataSource() datasource.DataSource {
	return &FileDataSource{}
}

type FileDataSource struct {
	client *OpenAIClient
}

func (d *FileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

type FileDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	FileID    types.String `tfsdk:"file_id"` // Legacy SDKv2 input argument
	Filename  types.String `tfsdk:"filename"`
	Bytes     types.Int64  `tfsdk:"bytes"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
	Purpose   types.String `tfsdk:"purpose"`
	ProjectID types.String `tfsdk:"project_id"`
}

func (d *FileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "File data source allows you to retrieve details of a specific file.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the file.",
				Computed:    true,
			},
			"file_id": schema.StringAttribute{
				Description: "The ID of the file to retrieve.",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID (for Terraform logic only)",
				Optional:    true,
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
		},
	}
}

func (d *FileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileID := data.FileID.ValueString()
	if fileID == "" {
		resp.Diagnostics.AddError("Missing file_id", "The file_id attribute is required.")
		return
	}

	// path is simply "files/{id}"
	path := fmt.Sprintf("files/%s", fileID)

	respBody, err := d.client.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving file", err.Error())
		return
	}

	var fileResp FileResponse
	if err := json.Unmarshal(respBody, &fileResp); err != nil {
		resp.Diagnostics.AddError("Error parsing file response", err.Error())
		return
	}

	data.ID = types.StringValue(fileResp.ID)
	data.Filename = types.StringValue(fileResp.Filename)
	data.Bytes = types.Int64Value(fileResp.Bytes)
	data.CreatedAt = types.Int64Value(fileResp.CreatedAt)
	data.Purpose = types.StringValue(fileResp.Purpose)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
