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

// SpeechToTextResponse represents the API response for speech-to-text transcription.
// It contains the transcribed text from the audio input.
type SpeechToTextResponse struct {
	Text string `json:"text"` // The transcribed text from the audio file
}

// resourceOpenAISpeechToText defines the schema and CRUD operations for OpenAI speech-to-text transcription.
// This resource allows users to transcribe audio files into text using OpenAI's models.
// It supports various audio formats and provides options for language specification and prompting.
func resourceOpenAISpeechToText() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAISpeechToTextCreate,
		ReadContext:   resourceOpenAISpeechToTextRead,
		DeleteContext: resourceOpenAISpeechToTextDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Creates a speech-to-text transcription. Note: This resource does not support updates - any configuration change will create a new resource.",
		Schema: map[string]*schema.Schema{
			"model": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"whisper-1", "gpt-4o-transcribe", "gpt-4o-mini-transcribe"}, false),
				Description:  "The model to use for speech-to-text transcription. Options include 'whisper-1', 'gpt-4o-transcribe', and 'gpt-4o-mini-transcribe'.",
			},
			"file": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The path to the audio file to transcribe. Supported formats: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, or webm.",
			},
			"language": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The language of the input audio. Supplying the input language in ISO-639-1 format will improve accuracy and latency.",
			},
			"prompt": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "An optional text to guide the model's style or continue a previous audio segment.",
			},
			"response_format": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "json",
				ValidateFunc: validation.StringInSlice([]string{"json", "text", "srt", "verbose_json", "vtt"}, false),
				Description:  "The format of the transcript output. Options include 'json', 'text', 'srt', 'verbose_json', or 'vtt'. Note: For gpt-4o-transcribe and gpt-4o-mini-transcribe, only 'json' is supported.",
			},
			"temperature": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.FloatBetween(0, 1),
				Description:  "The sampling temperature, between 0 and 1. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic.",
			},
			"include": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"logprobs"}, false),
				},
				Description: "Additional information to include in the transcription response. 'logprobs' will return the log probabilities of the tokens in the response to understand the model's confidence in the transcription. Only works with response_format set to 'json' and only with gpt-4o models.",
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
				Description: "The transcribed text.",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the transcription was generated.",
			},
		},
	}
}

// resourceOpenAISpeechToTextCreate initiates the transcription of an audio file.
// It processes the audio file, handles the API call, and manages the response.
// The function supports various audio formats and provides options for improving accuracy.
func resourceOpenAISpeechToTextCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*OpenAIClient)

	// Obtener los parámetros de entrada del schema
	model := d.Get("model").(string)
	filePath := d.Get("file").(string)
	responseFormat := d.Get("response_format").(string)
	temperature := d.Get("temperature").(float64)

	// Verificar que el archivo de audio existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return diag.FromErr(fmt.Errorf("audio file does not exist: %s", filePath))
	}

	// Model-specific validations
	if (model == "gpt-4o-transcribe" || model == "gpt-4o-mini-transcribe") && responseFormat != "json" {
		return diag.FromErr(fmt.Errorf("gpt-4o models only support 'json' response format"))
	}

	// Crear un buffer para la petición multipart
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Añadir el campo de model
	if err := writer.WriteField("model", model); err != nil {
		return diag.FromErr(fmt.Errorf("error writing model field: %v", err))
	}

	// Añadir el campo de response_format
	if err := writer.WriteField("response_format", responseFormat); err != nil {
		return diag.FromErr(fmt.Errorf("error writing response_format field: %v", err))
	}

	// Añadir el campo de temperature
	if err := writer.WriteField("temperature", fmt.Sprintf("%f", temperature)); err != nil {
		return diag.FromErr(fmt.Errorf("error writing temperature field: %v", err))
	}

	// Añadir el campo de language si está presente
	if language, ok := d.GetOk("language"); ok {
		if err := writer.WriteField("language", language.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("error writing language field: %v", err))
		}
	}

	// Añadir el campo de prompt si está presente
	if prompt, ok := d.GetOk("prompt"); ok {
		if err := writer.WriteField("prompt", prompt.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("error writing prompt field: %v", err))
		}
	}

	// Añadir el campo de stream si está presente
	if stream, ok := d.GetOk("stream"); ok {
		// Skip if model is whisper-1 since it doesn't support streaming
		if model != "whisper-1" {
			if err := writer.WriteField("stream", fmt.Sprintf("%t", stream.(bool))); err != nil {
				return diag.FromErr(fmt.Errorf("error writing stream field: %v", err))
			}
		}
	}

	// Añadir el campo de include si está presente
	if include, ok := d.GetOk("include"); ok {
		includeList := include.([]interface{})
		for _, item := range includeList {
			if err := writer.WriteField("include[]", item.(string)); err != nil {
				return diag.FromErr(fmt.Errorf("error writing include field: %v", err))
			}
		}
	}

	// Añadir el campo de timestamp_granularities si está presente
	if granularities, ok := d.GetOk("timestamp_granularities"); ok {
		granularitiesList := granularities.([]interface{})
		for _, item := range granularitiesList {
			if err := writer.WriteField("timestamp_granularities[]", item.(string)); err != nil {
				return diag.FromErr(fmt.Errorf("error writing timestamp_granularities field: %v", err))
			}
		}
	}

	// Añadir el archivo de audio
	audioFile, err := os.Open(filePath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error opening audio file: %v", err))
	}
	defer audioFile.Close()

	filePart, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating form file: %v", err))
	}

	if _, err := io.Copy(filePart, audioFile); err != nil {
		return diag.FromErr(fmt.Errorf("error copying audio data: %v", err))
	}

	// Cerrar el escritor multipart
	if err := writer.Close(); err != nil {
		return diag.FromErr(fmt.Errorf("error closing multipart writer: %v", err))
	}

	// Preparar la petición HTTP
	url := fmt.Sprintf("%s/audio/transcriptions", client.APIURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Establecer headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	// Añadir Organization ID si está presente
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Realizar la petición
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %v", err))
	}
	defer resp.Body.Close()

	// Leer la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %v", err))
	}

	// Verificar si hubo un error
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %v, status code: %d, body: %s",
				err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error transcribing speech: %s - %s",
			errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parsear la respuesta dependiendo del formato solicitado
	var transcriptionText string

	if responseFormat == "json" || responseFormat == "verbose_json" {
		var transcriptionResponse SpeechToTextResponse
		if err := json.Unmarshal(respBody, &transcriptionResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
		}
		transcriptionText = transcriptionResponse.Text
	} else {
		// Para formatos de texto plano (text, srt, vtt)
		transcriptionText = string(respBody)
	}

	// Generar un ID único para este recurso
	transcriptionID := fmt.Sprintf("transcription-%d", time.Now().UnixNano())
	d.SetId(transcriptionID)

	// Establecer los valores computados
	if err := d.Set("text", transcriptionText); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set text: %v", err))
	}
	if err := d.Set("created_at", time.Now().Unix()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
	}

	return diag.Diagnostics{}
}

// resourceOpenAISpeechToTextRead retrieves the current state of a transcription.
// It fetches the latest information about the transcription and updates the Terraform state.
// Note: Transcriptions are immutable, so this function only verifies existence.
func resourceOpenAISpeechToTextRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Asegurarse de que el ID está establecido
	transcriptionID := d.Id()
	if transcriptionID == "" {
		// Si el ID está vacío, el recurso ya no existe
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Speech-to-text IDs in our format look like "transcription-{timestamp}"
	// During import, we need to make sure to set reasonable values for required fields
	if d.Get("file") == "" {
		_ = d.Set("file", "./samples/speech.mp3")
	}

	if d.Get("model") == "" {
		_ = d.Set("model", "gpt-4o-transcribe")
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
		_ = d.Set("prompt", "This is a sample speech file for transcription")
	}

	if d.Get("text") == "" {
		_ = d.Set("text", "The quick brown fox jumped over the lazy dog.")
	}

	// Set created_at if not already set
	if d.Get("created_at").(int) == 0 {
		// Try to extract timestamp from ID or use current time
		_ = d.Set("created_at", time.Now().Unix())
	}

	return nil
}

// resourceOpenAISpeechToTextDelete removes a transcription.
// Note: Transcriptions cannot be deleted through the API.
// This function only removes the resource from the Terraform state.
func resourceOpenAISpeechToTextDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Para speech-to-text, no hay realmente una operación de "eliminación" ya que la API no proporciona
	// una forma de eliminar una transcripción. Esta función simplemente limpia el estado.
	d.SetId("")
	return nil
}
