package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// AudioTranscriptionsResponse represents the API response for listing OpenAI audio transcriptions
type AudioTranscriptionsResponse struct {
	Object  string                   `json:"object"`
	Data    []AudioTranscriptionData `json:"data"`
	HasMore bool                     `json:"has_more"`
}

// AudioTranscriptionData represents a single transcription in the list response
type AudioTranscriptionData struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Status    string `json:"status"`
	Model     string `json:"model"`
	Text      string `json:"text"`
	Duration  int    `json:"duration"`
	Language  string `json:"language"`
}

// dataSourceOpenAIAudioTranscriptions defines the schema and read operation for the OpenAI audio transcriptions data source.
// This data source allows retrieving information about all audio transcriptions for a specific OpenAI project.
func dataSourceOpenAIAudioTranscriptions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIAudioTranscriptionsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the project associated with the audio transcriptions. If not specified, the API key's default project will be used.",
			},
			"model": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Filter transcriptions by model. Options include 'whisper-1', 'gpt-4o-transcribe', and 'gpt-4o-mini-transcribe'.",
				ValidateFunc: validation.StringInSlice([]string{"whisper-1", "gpt-4o-transcribe", "gpt-4o-mini-transcribe"}, false),
			},
			"transcriptions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the audio transcription",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp when the transcription was created",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the transcription",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The model used for transcription",
						},
						"text": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The transcribed text",
						},
						"duration": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The duration of the audio in seconds",
						},
						"language": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The detected or specified language of the audio",
						},
					},
				},
			},
		},
	}
}

// dataSourceOpenAIAudioTranscriptionsRead fetches information about all audio transcriptions for a specific project from OpenAI.
func dataSourceOpenAIAudioTranscriptionsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Return an error indicating this operation is not supported
	return diag.FromErr(fmt.Errorf("listing all audio transcriptions is not supported by the OpenAI API: " +
		"The OpenAI API currently doesn't provide an endpoint to list all transcriptions. " +
		"You can only retrieve individual transcriptions using the openai_audio_transcription data source with a specific transcription_id."))
}
