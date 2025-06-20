package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// dataSourceOpenAIAudioTranslation provides a data source to retrieve information about an audio translation.
// Since translations in OpenAI are not retrievable after creation, this data source is primarily for documentation and import.
func dataSourceOpenAIAudioTranslation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIAudioTranslationRead,
		Schema: map[string]*schema.Schema{
			"translation_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the audio translation.",
			},
			"model": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The model used for audio translation. Currently only 'whisper-1' is available.",
				ValidateFunc: validation.StringInSlice([]string{"whisper-1"}, false),
			},
			"text": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The translated text from the audio.",
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
							Description: "The translated text for this segment.",
						},
					},
				},
				Description: "The segments of the audio translation, with timing information.",
			},
		},
	}
}

// dataSourceOpenAIAudioTranslationRead reads information about an existing OpenAI audio translation.
// Since OpenAI does not provide an API to retrieve translations after they're created,
// this function simply validates the ID and returns a placeholder response.
func dataSourceOpenAIAudioTranslationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	translationID := d.Get("translation_id").(string)

	// Set the ID to the translation ID
	d.SetId(translationID)

	// No actual fetching from API since OpenAI doesn't provide this capability
	// The purpose of this data source is primarily for documentation and import

	return diags
}
