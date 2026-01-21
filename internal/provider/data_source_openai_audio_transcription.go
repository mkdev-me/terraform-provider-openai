package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies interface
var _ datasource.DataSource = &AudioTranscriptionDataSource{}
var _ datasource.DataSource = &AudioTranscriptionsDataSource{}
var _ datasource.DataSource = &AudioTranslationDataSource{}
var _ datasource.DataSource = &AudioTranslationsDataSource{}
var _ datasource.DataSource = &SpeechToTextDataSource{}
var _ datasource.DataSource = &SpeechToTextsDataSource{}

// --- Audio Transcription ---

func NewAudioTranscriptionDataSource() datasource.DataSource {
	return &AudioTranscriptionDataSource{}
}

type AudioTranscriptionDataSource struct{}

type AudioTranscriptionDataSourceModel struct {
	TranscriptionID types.String  `tfsdk:"transcription_id"`
	Model           types.String  `tfsdk:"model"`
	Text            types.String  `tfsdk:"text"`
	Duration        types.Float64 `tfsdk:"duration"`
	Segments        types.List    `tfsdk:"segments"`
	ID              types.String  `tfsdk:"id"`
}

type AudioTranscriptionSegmentModel struct {
	ID    types.Int64   `tfsdk:"id"`
	Start types.Float64 `tfsdk:"start"`
	End   types.Float64 `tfsdk:"end"`
	Text  types.String  `tfsdk:"text"`
}

func (d *AudioTranscriptionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audio_transcription"
}

func (d *AudioTranscriptionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an audio transcription. " +
			"Since transcriptions in OpenAI are not retrievable after creation, this data source is primarily for documentation and import " +
			"purposes, acting as a placeholder.",
		Attributes: map[string]schema.Attribute{
			"transcription_id": schema.StringAttribute{
				Description: "The ID of the audio transcription.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"model": schema.StringAttribute{
				Description: "The model used for audio transcription. Options include 'whisper-1'.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("whisper-1"),
				},
			},
			"text": schema.StringAttribute{
				Description: "The transcribed text from the audio.",
				Optional:    true,
			},
			"duration": schema.Float64Attribute{
				Description: "The duration of the audio file in seconds.",
				Optional:    true,
			},
			"segments": schema.ListNestedAttribute{
				Description: "The segments of the audio transcription, with timing information.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the segment.",
							Optional:    true,
						},
						"start": schema.Float64Attribute{
							Description: "The start time of the segment in seconds.",
							Optional:    true,
						},
						"end": schema.Float64Attribute{
							Description: "The end time of the segment in seconds.",
							Optional:    true,
						},
						"text": schema.StringAttribute{
							Description: "The transcribed text for this segment.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (d *AudioTranscriptionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AudioTranscriptionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Just set ID to transcription ID
	data.ID = data.TranscriptionID

	// Return data as is (preserving optional inputs)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Audio Transcriptions (Plural) ---

func NewAudioTranscriptionsDataSource() datasource.DataSource {
	return &AudioTranscriptionsDataSource{}
}

type AudioTranscriptionsDataSource struct{}

func (d *AudioTranscriptionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audio_transcriptions"
}

func (d *AudioTranscriptionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieving all audio transcriptions is not supported by the OpenAI API.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the project.",
				Optional:    true,
			},
			"model": schema.StringAttribute{
				Description: "Filter by model.",
				Optional:    true,
			},
			"transcriptions": schema.ListAttribute{
				Description: "List of transcriptions (Empty as not supported).",
				Computed:    true,
				ElementType: types.StringType, // Keeping simple as it errors anyway
			},
		},
	}
}

func (d *AudioTranscriptionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp.Diagnostics.AddError(
		"Operation Not Supported",
		"listing all audio transcriptions is not supported by the OpenAI API: "+
			"The OpenAI API currently doesn't provide an endpoint to list all transcriptions. "+
			"You can only retrieve individual transcriptions using the openai_audio_transcription data source with a specific transcription_id.",
	)
}

// --- Audio Translation ---

func NewAudioTranslationDataSource() datasource.DataSource {
	return &AudioTranslationDataSource{}
}

type AudioTranslationDataSource struct{}

type AudioTranslationDataSourceModel struct {
	TranslationID types.String  `tfsdk:"translation_id"`
	Model         types.String  `tfsdk:"model"`
	Text          types.String  `tfsdk:"text"`
	Duration      types.Float64 `tfsdk:"duration"`
	Segments      types.List    `tfsdk:"segments"`
	ID            types.String  `tfsdk:"id"`
}

func (d *AudioTranslationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audio_translation"
}

func (d *AudioTranslationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an audio translation. " +
			"Since translations in OpenAI are not retrievable after creation, this data source is primarily for documentation and import.",
		Attributes: map[string]schema.Attribute{
			"translation_id": schema.StringAttribute{
				Description: "The ID of the audio translation.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"model": schema.StringAttribute{
				Description: "The model used for audio translation. Currently only 'whisper-1' is available.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("whisper-1"),
				},
			},
			"text": schema.StringAttribute{
				Description: "The translated text from the audio.",
				Optional:    true,
			},
			"duration": schema.Float64Attribute{
				Description: "The duration of the audio file in seconds.",
				Optional:    true,
			},
			"segments": schema.ListNestedAttribute{
				Description: "The segments of the audio translation, with timing information.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the segment.",
							Optional:    true,
						},
						"start": schema.Float64Attribute{
							Description: "The start time of the segment in seconds.",
							Optional:    true,
						},
						"end": schema.Float64Attribute{
							Description: "The end time of the segment in seconds.",
							Optional:    true,
						},
						"text": schema.StringAttribute{
							Description: "The translated text for this segment.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (d *AudioTranslationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AudioTranslationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = data.TranslationID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Audio Translations (Plural) ---

func NewAudioTranslationsDataSource() datasource.DataSource {
	return &AudioTranslationsDataSource{}
}

type AudioTranslationsDataSource struct{}

func (d *AudioTranslationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audio_translations"
}

func (d *AudioTranslationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieving all audio translations is not supported by the OpenAI API.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the project.",
				Optional:    true,
			},
			"model": schema.StringAttribute{
				Description: "Filter by model.",
				Optional:    true,
			},
			"translations": schema.ListAttribute{
				Description: "List of translations (Empty as not supported).",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *AudioTranslationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp.Diagnostics.AddError(
		"Operation Not Supported",
		"listing all audio translations is not supported by the OpenAI API.",
	)
}

// --- Speech to Text ---

func NewSpeechToTextDataSource() datasource.DataSource {
	return &SpeechToTextDataSource{}
}

type SpeechToTextDataSource struct{}

type SpeechToTextDataSourceModel struct {
	TranscriptionID types.String `tfsdk:"transcription_id"`
	Model           types.String `tfsdk:"model"`
	Text            types.String `tfsdk:"text"`
	CreatedAt       types.Int64  `tfsdk:"created_at"`
	ID              types.String `tfsdk:"id"`
}

func (d *SpeechToTextDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_speech_to_text"
}

func (d *SpeechToTextDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a speech-to-text transcription. " +
			"Since transcriptions in OpenAI are not retrievable after creation, this data source is primarily for documentation and import.",
		Attributes: map[string]schema.Attribute{
			"transcription_id": schema.StringAttribute{
				Description: "The ID of the speech-to-text transcription.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"model": schema.StringAttribute{
				Description: "The model used for transcription.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("whisper-1", "gpt-4o-transcribe", "gpt-4o-mini-transcribe"),
				},
			},
			"text": schema.StringAttribute{
				Description: "The transcribed text.",
				Optional:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The timestamp when the transcription was generated.",
				Optional:    true,
			},
		},
	}
}

func (d *SpeechToTextDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SpeechToTextDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = data.TranscriptionID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- Speech to Texts (Plural) ---

func NewSpeechToTextsDataSource() datasource.DataSource {
	return &SpeechToTextsDataSource{}
}

type SpeechToTextsDataSource struct{}

func (d *SpeechToTextsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_speech_to_texts"
}

func (d *SpeechToTextsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieving all speech-to-text transcriptions is not supported by the OpenAI API.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the project.",
				Optional:    true,
			},
			"model": schema.StringAttribute{
				Description: "Filter by model.",
				Optional:    true,
			},
			"transcriptions": schema.ListAttribute{
				Description: "List of transcriptions (Empty as not supported).",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *SpeechToTextsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp.Diagnostics.AddError(
		"Operation Not Supported",
		"listing all speech-to-text transcriptions is not supported by the OpenAI API.",
	)
}
