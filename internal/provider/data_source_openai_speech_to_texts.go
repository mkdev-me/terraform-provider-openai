package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SpeechToTextsResponse represents the API response for listing OpenAI speech-to-text outputs
type SpeechToTextsResponse struct {
	Object  string             `json:"object"`
	Data    []SpeechToTextData `json:"data"`
	HasMore bool               `json:"has_more"`
}

// SpeechToTextData represents a single speech-to-text result in the list response
type SpeechToTextData struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Status    string `json:"status"`
	Model     string `json:"model"`
	Text      string `json:"text"`
	Duration  int    `json:"duration"`
}

// dataSourceOpenAISpeechToTexts defines the schema and read operation for the OpenAI speech-to-texts data source.
// This data source allows retrieving information about all speech-to-text conversions for a specific OpenAI project.
func dataSourceOpenAISpeechToTexts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAISpeechToTextsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the project associated with the speech-to-text conversions. If not specified, the API key's default project will be used.",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Project-specific API key to use for authentication. If not provided, the provider's default API key will be used.",
			},
			"model": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Filter by model. Options include 'whisper-1', 'gpt-4o-transcribe', and 'gpt-4o-mini-transcribe'.",
				ValidateFunc: validation.StringInSlice([]string{"whisper-1", "gpt-4o-transcribe", "gpt-4o-mini-transcribe"}, false),
			},
			"speech_to_texts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the speech-to-text conversion",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp when the conversion was created",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the conversion",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The model used for speech-to-text conversion",
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
					},
				},
			},
		},
	}
}

// dataSourceOpenAISpeechToTextsRead fetches information about all speech-to-text conversions for a specific project from OpenAI.
func dataSourceOpenAISpeechToTextsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Return an error indicating this operation is not supported
	return diag.FromErr(fmt.Errorf("listing all speech-to-text conversions is not supported by the OpenAI API: " +
		"The OpenAI API currently doesn't provide an endpoint to list all speech-to-text conversions. " +
		"You can only retrieve individual speech-to-text conversions using the openai_speech_to_text data source with a specific transcription_id."))
}
