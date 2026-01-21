package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies interface
var _ datasource.DataSource = &TextToSpeechDataSource{}
var _ datasource.DataSource = &TextToSpeechsDataSource{}

// --- Text To Speech ---

func NewTextToSpeechDataSource() datasource.DataSource {
	return &TextToSpeechDataSource{}
}

type TextToSpeechDataSource struct{}

type TextToSpeechDataSourceModel struct {
	FilePath      types.String `tfsdk:"file_path"`
	Exists        types.Bool   `tfsdk:"exists"`
	FileSizeBytes types.Int64  `tfsdk:"file_size_bytes"`
	LastModified  types.Int64  `tfsdk:"last_modified"`
	ID            types.String `tfsdk:"id"`
}

func (d *TextToSpeechDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_text_to_speech"
}

func (d *TextToSpeechDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve OpenAI text-to-speech details. " +
			"This data source verifies the existence of a previously generated speech file.",
		Attributes: map[string]schema.Attribute{
			"file_path": schema.StringAttribute{
				Description: "The path to the text-to-speech audio file to verify.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"exists": schema.BoolAttribute{
				Description: "Whether the speech file exists.",
				Computed:    true,
			},
			"file_size_bytes": schema.Int64Attribute{
				Description: "The size of the speech file in bytes.",
				Computed:    true,
			},
			"last_modified": schema.Int64Attribute{
				Description: "The timestamp when the speech file was last modified.",
				Computed:    true,
			},
		},
	}
}

func (d *TextToSpeechDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TextToSpeechDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	filePath := data.FilePath.ValueString()

	// Check if the file exists and get stats
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, but return valid response with exists=false
			data.ID = types.StringValue(fmt.Sprintf("tts_file_%d", time.Now().Unix()))
			data.Exists = types.BoolValue(false)
			data.FileSizeBytes = types.Int64Value(0)
			data.LastModified = types.Int64Value(0)
		} else {
			resp.Diagnostics.AddError("Error accessing file", fmt.Sprintf("Error accessing file %s: %v", filePath, err))
			return
		}
	} else {
		// File exists
		data.ID = types.StringValue(fmt.Sprintf("tts_file_%d", time.Now().Unix()))
		data.Exists = types.BoolValue(true)
		data.FileSizeBytes = types.Int64Value(fileInfo.Size())
		data.LastModified = types.Int64Value(fileInfo.ModTime().Unix())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Text To Speechs (Plural) ---

func NewTextToSpeechsDataSource() datasource.DataSource {
	return &TextToSpeechsDataSource{}
}

type TextToSpeechsDataSource struct{}

func (d *TextToSpeechsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_text_to_speechs"
}

func (d *TextToSpeechsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieving all text-to-speech conversions is not supported by the OpenAI API.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the project.",
				Optional:    true,
			},
			"model": schema.StringAttribute{
				Description: "Filter by model.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("tts-1", "tts-1-hd", "tts-1-1106"),
				},
			},
			"voice": schema.StringAttribute{
				Description: "Filter by voice.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("alloy", "echo", "fable", "onyx", "nova", "shimmer"),
				},
			},
			"text_to_speechs": schema.ListAttribute{
				Description: "List of conversions (Empty as not supported).",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *TextToSpeechsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp.Diagnostics.AddError(
		"Operation Not Supported",
		"listing all text-to-speech conversions is not supported by the OpenAI API.",
	)
}
