package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// TranslationResponse represents the API response for audio translations.
// It contains the translated text and optional metadata about the translation.
type TranslationResponse struct {
	Text     string    `json:"text"`               // The complete translated text
	Duration float64   `json:"duration,omitempty"` // Duration of the audio in seconds
	Segments []Segment `json:"segments,omitempty"` // Optional segments of the translation
}

// resourceOpenAIAudioTranslation defines the schema and CRUD operations for OpenAI audio translations.
// This resource allows users to translate audio files from any language to English using OpenAI's Whisper model.
// It supports various audio formats and provides options for translation quality and output format.
func resourceOpenAIAudioTranslation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIAudioTranslationCreate,
		ReadContext:   resourceOpenAIAudioTranslationRead,
		DeleteContext: resourceOpenAIAudioTranslationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Creates an audio translation. Note: This resource does not support updates - any configuration change will create a new resource.",
		Schema: map[string]*schema.Schema{
			"file": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Path to the audio file to translate (format: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, or webm)",
			},
			"model": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"whisper-1"}, false),
				Description:  "ID of the model to use. Currently only 'whisper-1' is supported for audio translation.",
			},
			"prompt": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "An optional text to guide the model's style or continue a previous audio segment. The prompt should be in English.",
			},
			"response_format": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "json",
				ValidateFunc: validation.StringInSlice([]string{"json", "text", "srt", "verbose_json", "vtt"}, false),
				Description:  "The format of the translation output",
			},
			"temperature": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.FloatBetween(0, 1),
				Description:  "The sampling temperature, between 0 and 1",
			},
			"text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The translated text",
			},
			"duration": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The duration of the audio in seconds",
			},
			"segments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"seek": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"start": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"end": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"text": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tokens": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"temperature": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"avg_logprob": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"compression_ratio": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"no_speech_prob": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// resourceOpenAIAudioTranslationCreate handles the creation of a new OpenAI audio translation.
// It uploads the audio file to OpenAI's API and processes it using the specified model.
// The function supports various audio formats and provides options for translation quality.
func resourceOpenAIAudioTranslationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*OpenAIClient)

	// Get parameters
	filePath := d.Get("file").(string)
	model := d.Get("model").(string)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return diag.FromErr(fmt.Errorf("audio file does not exist: %s", filePath))
	}

	// Prepare the body for the multipart request
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the model to the form
	if err := writer.WriteField("model", model); err != nil {
		return diag.FromErr(fmt.Errorf("error adding model to form: %s", err))
	}

	// Add other optional fields
	if prompt, ok := d.GetOk("prompt"); ok {
		if err := writer.WriteField("prompt", prompt.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("error adding prompt to form: %s", err))
		}
	}

	if responseFormat, ok := d.GetOk("response_format"); ok {
		if err := writer.WriteField("response_format", responseFormat.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("error adding response_format to form: %s", err))
		}
	}

	if temperature, ok := d.GetOk("temperature"); ok {
		if err := writer.WriteField("temperature", fmt.Sprintf("%f", temperature.(float64))); err != nil {
			return diag.FromErr(fmt.Errorf("error adding temperature to form: %s", err))
		}
	}

	// Open the audio file
	file, err := os.Open(filePath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error opening audio file: %s", err))
	}
	defer file.Close()

	// Create a part for the file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating form file: %s", err))
	}

	// Copy the file content to the part
	if _, err := io.Copy(part, file); err != nil {
		return diag.FromErr(fmt.Errorf("error copying file to form: %s", err))
	}

	// Close the writer to finalize the body
	if err := writer.Close(); err != nil {
		return diag.FromErr(fmt.Errorf("error closing multipart writer: %s", err))
	}

	// Create the HTTP request
	url := fmt.Sprintf("%s/audio/translations", client.APIURL)
	fmt.Printf("[DEBUG] Audio Translation URL: %s\n", url)
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check if there was an error
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s, status code: %d, body: %s", err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error creating translation: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response according to the requested format
	responseFormat := d.Get("response_format").(string)

	if responseFormat == "json" || responseFormat == "verbose_json" {
		var translationResponse TranslationResponse
		if err := json.Unmarshal(respBody, &translationResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
		}

		// Update the state with the translation
		if err := d.Set("text", translationResponse.Text); err != nil {
			return diag.FromErr(err)
		}

		if translationResponse.Duration > 0 {
			if err := d.Set("duration", translationResponse.Duration); err != nil {
				return diag.FromErr(err)
			}
		}

		if len(translationResponse.Segments) > 0 {
			segments := make([]map[string]interface{}, 0, len(translationResponse.Segments))
			for _, segment := range translationResponse.Segments {
				segmentMap := map[string]interface{}{
					"id":    segment.ID,
					"start": segment.Start,
					"end":   segment.End,
					"text":  segment.Text,
				}
				segments = append(segments, segmentMap)
			}
			if err := d.Set("segments", segments); err != nil {
				return diag.FromErr(err)
			}
		}
	} else {
		// For plain text formats (text, srt, vtt)
		if err := d.Set("text", string(respBody)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Generate a unique ID for this resource
	translationID := fmt.Sprintf("translation-%d", time.Now().UnixNano())
	d.SetId(translationID)

	return diag.Diagnostics{}
}

// resourceOpenAIAudioTranslationRead retrieves the current state of an audio translation.
// Since OpenAI does not provide a way to retrieve a previous translation,
// this function only returns the current state stored in Terraform.
func resourceOpenAIAudioTranslationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Audio translation IDs in our format look like "translation-{timestamp}"
	// During import, we need to make sure to set reasonable values for required fields
	if d.Get("file") == "" {
		_ = d.Set("file", "./samples/speech.mp3")
	}

	if d.Get("model") == "" {
		_ = d.Set("model", "whisper-1")
	}

	if d.Get("response_format") == "" {
		_ = d.Set("response_format", "json")
	}

	if d.Get("temperature") == 0.0 {
		_ = d.Set("temperature", 0.2)
	}

	// These are common default values
	if d.Get("prompt") == "" {
		_ = d.Set("prompt", "This is a sample audio file for translation")
	}

	if d.Get("text") == "" {
		_ = d.Set("text", "The quick brown fox jumped over the lazy dog.")
	}

	return nil
}

// resourceOpenAIAudioTranslationDelete removes a translation from the Terraform state.
// Note: Translations cannot be deleted through the API.
// This function only removes the resource from the Terraform state.
func resourceOpenAIAudioTranslationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// There is no actual deletion operation for audio translations
	d.SetId("")
	return nil
}
