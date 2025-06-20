package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// dataSourceOpenAIAudioTranscription provides a data source to retrieve information about an audio transcription.
// Since transcriptions in OpenAI are not retrievable after creation, this data source is primarily for documentation and import.
func dataSourceOpenAIAudioTranscription() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIAudioTranscriptionRead,
		Schema: map[string]*schema.Schema{
			"transcription_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the audio transcription.",
			},
			"model": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The model used for audio transcription. Options include 'whisper-1'.",
				ValidateFunc: validation.StringInSlice([]string{"whisper-1"}, false),
			},
			"text": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The transcribed text from the audio.",
			},
			"duration": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "The duration of the audio file in seconds.",
			},
			"segments": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The ID of the segment.",
						},
						"start": {
							Type:        schema.TypeFloat,
							Optional:    true,
							Description: "The start time of the segment in seconds.",
						},
						"end": {
							Type:        schema.TypeFloat,
							Optional:    true,
							Description: "The end time of the segment in seconds.",
						},
						"text": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The transcribed text for this segment.",
						},
					},
				},
				Description: "The segments of the audio transcription, with timing information.",
			},
		},
	}
}

// dataSourceOpenAIAudioTranscriptionRead reads information about an existing OpenAI audio transcription.
// Since OpenAI does not provide an API to retrieve transcriptions after they're created,
// this function simply validates the ID and returns a placeholder response.
func dataSourceOpenAIAudioTranscriptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	transcriptionID := d.Get("transcription_id").(string)

	// Set the ID to the transcription ID
	d.SetId(transcriptionID)

	// No actual fetching from API since OpenAI doesn't provide this capability
	// The purpose of this data source is primarily for documentation and import

	return diags
}
