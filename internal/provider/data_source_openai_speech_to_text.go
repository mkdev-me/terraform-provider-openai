package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// dataSourceOpenAISpeechToText provides a data source to retrieve information about a speech-to-text transcription.
// Since transcriptions in OpenAI are not retrievable after creation, this data source is primarily for documentation and import.
func dataSourceOpenAISpeechToText() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAISpeechToTextRead,
		Schema: map[string]*schema.Schema{
			"transcription_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the speech-to-text transcription.",
			},
			"model": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The model used for transcription. Options include 'whisper-1', 'gpt-4o-transcribe', and 'gpt-4o-mini-transcribe'.",
				ValidateFunc: validation.StringInSlice([]string{"whisper-1", "gpt-4o-transcribe", "gpt-4o-mini-transcribe"}, false),
			},
			"text": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The transcribed text.",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The timestamp when the transcription was generated.",
			},
		},
	}
}

// dataSourceOpenAISpeechToTextRead reads information about an existing OpenAI speech-to-text transcription.
// Since OpenAI does not provide an API to retrieve transcriptions after they're created,
// this function simply validates the ID and returns a placeholder response.
func dataSourceOpenAISpeechToTextRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	transcriptionID := d.Get("transcription_id").(string)

	// Set the ID to the transcription ID
	d.SetId(transcriptionID)

	// No actual fetching from API since OpenAI doesn't provide this capability
	// The purpose of this data source is primarily for documentation and import

	return diags
}
