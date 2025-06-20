package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// TextToSpeechsResponse represents the API response for listing OpenAI text-to-speech outputs
type TextToSpeechsResponse struct {
	Object  string             `json:"object"`
	Data    []TextToSpeechData `json:"data"`
	HasMore bool               `json:"has_more"`
}

// TextToSpeechData represents a single text-to-speech result in the list response
type TextToSpeechData struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Status    string `json:"status"`
	Model     string `json:"model"`
	Voice     string `json:"voice"`
	Input     string `json:"input"`
	Duration  int    `json:"duration"`
}

// dataSourceOpenAITextToSpeechs defines the schema and read operation for the OpenAI text-to-speech data source.
// This data source allows retrieving information about all text-to-speech conversions for a specific OpenAI project.
func dataSourceOpenAITextToSpeechs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAITextToSpeechsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the project associated with the text-to-speech conversions. If not specified, the API key's default project will be used.",
			},
			"model": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Filter by model. Options include 'tts-1', 'tts-1-hd', and 'tts-1-1106'.",
				ValidateFunc: validation.StringInSlice([]string{"tts-1", "tts-1-hd", "tts-1-1106"}, false),
			},
			"voice": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Filter by voice. Options include 'alloy', 'echo', 'fable', 'onyx', 'nova', and 'shimmer'.",
				ValidateFunc: validation.StringInSlice([]string{"alloy", "echo", "fable", "onyx", "nova", "shimmer"}, false),
			},
			"text_to_speechs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the text-to-speech conversion",
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
							Description: "The model used for text-to-speech conversion",
						},
						"voice": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The voice used for text-to-speech conversion",
						},
						"input": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The input text that was converted to speech",
						},
						"duration": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The duration of the generated audio in seconds",
						},
					},
				},
			},
		},
	}
}

// dataSourceOpenAITextToSpeechsRead fetches information about all text-to-speech conversions for a specific project from OpenAI.
func dataSourceOpenAITextToSpeechsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Return an error indicating this operation is not supported
	return diag.FromErr(fmt.Errorf("listing all text-to-speech conversions is not supported by the OpenAI API: " +
		"The OpenAI API currently doesn't provide an endpoint to list all text-to-speech conversions. " +
		"You can only retrieve individual text-to-speech conversions using the openai_text_to_speech data source with a specific file_path."))
}
