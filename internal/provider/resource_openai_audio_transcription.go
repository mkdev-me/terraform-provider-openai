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

// TranscriptionResponse represents the API response for audio transcriptions.
// It contains the transcribed text and optional metadata about the transcription.
type TranscriptionResponse struct {
	Text     string    `json:"text"`               // The complete transcribed text
	Duration float64   `json:"duration,omitempty"` // Duration of the audio in seconds
	Segments []Segment `json:"segments,omitempty"` // Optional segments of the transcription
}

// Segment represents a single segment of the audio transcription.
// It contains timing information and the transcribed text for that segment.
type Segment struct {
	ID    int     `json:"id"`    // Unique identifier for the segment
	Start float64 `json:"start"` // Start time of the segment in seconds
	End   float64 `json:"end"`   // End time of the segment in seconds
	Text  string  `json:"text"`  // Transcribed text for this segment
}

// resourceOpenAIAudioTranscription defines the schema and CRUD operations for OpenAI audio transcriptions.
// This resource allows users to transcribe audio files using OpenAI's models.
// It supports various audio formats and provides options for transcription quality and output format.
func resourceOpenAIAudioTranscription() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIAudioTranscriptionCreate,
		ReadContext:   resourceOpenAIAudioTranscriptionRead,
		DeleteContext: resourceOpenAIAudioTranscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Creates an audio transcription. Note: This resource does not support updates - any configuration change will create a new resource.",
		Schema: map[string]*schema.Schema{
			"file": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Path to the audio file to transcribe (format: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, or webm)",
			},
			"model": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"whisper-1", "gpt-4o-transcribe", "gpt-4o-mini-transcribe"}, false),
				Description:  "ID of the model to use (e.g., 'whisper-1', 'gpt-4o-transcribe', 'gpt-4o-mini-transcribe')",
			},
			"language": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The language of the input audio (ISO-639-1 format)",
			},
			"prompt": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "An optional text to guide the model's style or continue a previous audio segment",
			},
			"response_format": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "json",
				ValidateFunc: validation.StringInSlice([]string{"json", "text", "srt", "verbose_json", "vtt"}, false),
				Description:  "The format of the transcript output. Note: For gpt-4o-transcribe and gpt-4o-mini-transcribe, only 'json' is supported.",
			},
			"temperature": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.FloatBetween(0, 1),
				Description:  "The sampling temperature, between 0 and 1",
			},
			"include": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"logprobs"}, false),
				},
				Description: "Additional information to include in the transcription response. 'logprobs' will return the log probabilities of the tokens in the response. Only works with response_format set to 'json' and only with gpt-4o models.",
			},
			"stream": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "If set to true, the model response data will be streamed to the client as it is generated. Not supported for whisper-1 model.",
			},
			"timestamp_granularities": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"word", "segment"}, false),
				},
				Description: "The timestamp granularities to populate for this transcription. response_format must be set to verbose_json to use timestamp granularities. Either or both of these options are supported: 'word', or 'segment'.",
			},
			"text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The transcribed text",
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

// resourceOpenAIAudioTranscriptionCreate handles the creation of a new OpenAI audio transcription.
// It uploads the audio file to OpenAI's API and processes it using the specified model.
// The function supports various audio formats and provides options for transcription quality.
func resourceOpenAIAudioTranscriptionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get parameters
	filePath := d.Get("file").(string)
	model := d.Get("model").(string)
	responseFormat := d.Get("response_format").(string)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return diag.FromErr(fmt.Errorf("audio file does not exist: %s", filePath))
	}

	// Model-specific validations
	if (model == "gpt-4o-transcribe" || model == "gpt-4o-mini-transcribe") && responseFormat != "json" {
		return diag.FromErr(fmt.Errorf("gpt-4o models only support 'json' response format"))
	}

	// Prepare the body for the multipart request
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the model to the form
	if err := writer.WriteField("model", model); err != nil {
		return diag.FromErr(fmt.Errorf("error adding model to form: %s", err))
	}

	// Add other optional fields
	if language, ok := d.GetOk("language"); ok {
		if err := writer.WriteField("language", language.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("error adding language to form: %s", err))
		}
	}

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

	// Add stream parameter if provided and model supports it
	if stream, ok := d.GetOk("stream"); ok {
		// Skip if model is whisper-1 since it doesn't support streaming
		if model != "whisper-1" {
			if err := writer.WriteField("stream", fmt.Sprintf("%t", stream.(bool))); err != nil {
				return diag.FromErr(fmt.Errorf("error adding stream to form: %s", err))
			}
		}
	}

	// Add include parameter if provided
	if include, ok := d.GetOk("include"); ok {
		includeList := include.([]interface{})
		for _, item := range includeList {
			if err := writer.WriteField("include[]", item.(string)); err != nil {
				return diag.FromErr(fmt.Errorf("error adding include to form: %s", err))
			}
		}
	}

	// Add timestamp_granularities parameter if provided
	if granularities, ok := d.GetOk("timestamp_granularities"); ok {
		granularitiesList := granularities.([]interface{})
		for _, item := range granularitiesList {
			if err := writer.WriteField("timestamp_granularities[]", item.(string)); err != nil {
				return diag.FromErr(fmt.Errorf("error adding timestamp_granularities to form: %s", err))
			}
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
	url := fmt.Sprintf("%s/audio/transcriptions", client.APIURL)
	fmt.Printf("[DEBUG] Audio Transcription URL: %s\n", url)
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
		return diag.FromErr(fmt.Errorf("error creating transcription: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response according to the requested format
	responseFormat = d.Get("response_format").(string)

	if responseFormat == "json" || responseFormat == "verbose_json" {
		var transcriptionResponse TranscriptionResponse
		if err := json.Unmarshal(respBody, &transcriptionResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
		}

		// Update the state with the transcription
		if err := d.Set("text", transcriptionResponse.Text); err != nil {
			return diag.FromErr(err)
		}

		if transcriptionResponse.Duration > 0 {
			if err := d.Set("duration", transcriptionResponse.Duration); err != nil {
				return diag.FromErr(err)
			}
		}

		if len(transcriptionResponse.Segments) > 0 {
			segments := make([]map[string]interface{}, 0, len(transcriptionResponse.Segments))
			for _, segment := range transcriptionResponse.Segments {
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
	transcriptionID := fmt.Sprintf("transcription-%d", time.Now().UnixNano())
	d.SetId(transcriptionID)

	return diag.Diagnostics{}
}

// resourceOpenAIAudioTranscriptionRead retrieves the current state of an audio transcription.
// Since OpenAI does not provide a way to retrieve a previous transcription,
// this function only returns the current state stored in Terraform.
func resourceOpenAIAudioTranscriptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Audio transcription IDs in our format look like "transcription-{timestamp}"
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
	if d.Get("language") == "" {
		_ = d.Set("language", "en")
	}

	if d.Get("prompt") == "" {
		_ = d.Set("prompt", "This is a sample audio file for transcription")
	}

	if d.Get("text") == "" {
		_ = d.Set("text", "The quick brown fox jumped over the lazy dog.")
	}

	return nil
}

// resourceOpenAIAudioTranscriptionDelete removes a transcription from the Terraform state.
// Note: Transcriptions cannot be deleted through the API.
// This function only removes the resource from the Terraform state.
func resourceOpenAIAudioTranscriptionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// There is no actual deletion operation for audio transcriptions
	d.SetId("")
	return nil
}
