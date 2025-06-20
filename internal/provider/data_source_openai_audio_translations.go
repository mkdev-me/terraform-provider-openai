package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// AudioTranslationsResponse represents the API response for listing OpenAI audio translations
type AudioTranslationsResponse struct {
	Object  string                 `json:"object"`
	Data    []AudioTranslationData `json:"data"`
	HasMore bool                   `json:"has_more"`
}

// AudioTranslationData represents a single translation in the list response
type AudioTranslationData struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Status    string `json:"status"`
	Model     string `json:"model"`
	Text      string `json:"text"`
	Duration  int    `json:"duration"`
}

// dataSourceOpenAIAudioTranslations defines the schema and read operation for the OpenAI audio translations data source.
// This data source allows retrieving information about all audio translations for a specific OpenAI project.
func dataSourceOpenAIAudioTranslations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIAudioTranslationsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the project associated with the audio translations. If not specified, the API key's default project will be used.",
			},
			"model": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Filter translations by model. Options include 'whisper-1'.",
				ValidateFunc: validation.StringInSlice([]string{"whisper-1"}, false),
			},
			"translations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the audio translation",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp when the translation was created",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the translation",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The model used for translation",
						},
						"text": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The translated text",
						},
						"duration": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The duration of the audio in seconds",
						},
					},
				},
			},
		},
	}
}

// dataSourceOpenAIAudioTranslationsRead fetches information about all audio translations for a specific project from OpenAI.
func dataSourceOpenAIAudioTranslationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Return an error indicating this operation is not supported
	return diag.FromErr(fmt.Errorf("listing all audio translations is not supported by the OpenAI API: " +
		"The OpenAI API currently doesn't provide an endpoint to list all translations. " +
		"You can only retrieve individual translations using the openai_audio_translation data source with a specific translation_id."))
}
