package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// TextToSpeechRequest represents the request payload for text-to-speech conversion.
// It specifies the parameters that control the speech synthesis process,
// including model selection, input text, voice, and various generation options.
type TextToSpeechRequest struct {
	Model          string  `json:"model"`                     // ID of the model to use (e.g., "tts-1", "tts-1-hd", "gpt-4o-mini-tts")
	Input          string  `json:"input"`                     // Text to convert to speech
	Voice          string  `json:"voice"`                     // Voice to use for synthesis
	ResponseFormat string  `json:"response_format,omitempty"` // Format of the audio output
	Speed          float64 `json:"speed,omitempty"`           // Speed of speech (0.25 to 4.0)
	Instructions   string  `json:"instructions,omitempty"`    // Instructions to guide the voice (gpt-4o-mini-tts only)
}

// resourceOpenAITextToSpeech defines the schema and CRUD operations for OpenAI text-to-speech conversion.
// This resource allows users to convert text into natural-sounding speech using OpenAI's models.
// It provides comprehensive control over the synthesis process and supports various output formats.
func resourceOpenAITextToSpeech() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAITextToSpeechCreate,
		ReadContext:   resourceOpenAITextToSpeechRead,
		DeleteContext: resourceOpenAITextToSpeechDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"model": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"tts-1", "tts-1-hd", "gpt-4o-mini-tts"}, false),
				Description:  "The model to use for text-to-speech conversion. Options include 'tts-1', 'tts-1-hd', and 'gpt-4o-mini-tts'.",
			},
			"input": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The text to convert to speech. Max length is 4096 characters.",
			},
			"voice": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"alloy", "echo", "fable", "onyx", "nova", "shimmer", "ash", "ballad", "coral", "sage", "verse"}, false),
				Description:  "The voice to use for speech. Options include 'alloy', 'echo', 'fable', 'onyx', 'nova', 'shimmer', 'ash', 'ballad', 'coral', 'sage', and 'verse'.",
			},
			"response_format": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "mp3",
				ValidateFunc: validation.StringInSlice([]string{"mp3", "opus", "aac", "flac", "wav", "pcm"}, false),
				Description:  "The format of the audio output. Options include 'mp3', 'opus', 'aac', 'flac', 'wav', and 'pcm'.",
			},
			"speed": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				Default:      1.0,
				ValidateFunc: validation.FloatBetween(0.25, 4.0),
				Description:  "The speed of the audio output. Must be between 0.25 and 4.0. Default is 1.0.",
			},
			"instructions": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Control the voice of your generated audio with additional instructions. Only works with 'gpt-4o-mini-tts' model.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The path where the audio file will be saved.",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the speech was generated.",
			},
		},
	}
}

// resourceOpenAITextToSpeechCreate initiates the text-to-speech conversion.
// It processes the synthesis request, handles the API call, and manages the response.
// The function supports various voices and provides control over speech characteristics.
func resourceOpenAITextToSpeechCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client from the provider meta
	openaiClient, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %v", err))
	}

	// Debug information
	fmt.Printf("[DEBUG] API URL: %s\n", openaiClient.APIURL)
	fmt.Printf("[DEBUG] Using API Key: %s...\n", openaiClient.APIKey[:10])
	fmt.Printf("[DEBUG] Organization ID: %s\n", openaiClient.OrganizationID)

	// Obtener los parámetros de entrada del schema
	model := d.Get("model").(string)
	input := d.Get("input").(string)
	voice := d.Get("voice").(string)
	responseFormat := d.Get("response_format").(string)
	speed := d.Get("speed").(float64)
	outputFilePath := d.Get("output_file").(string)

	// Debug request details
	fmt.Printf("[DEBUG] Creating text-to-speech with model=%s, voice=%s, format=%s\n",
		model, voice, responseFormat)

	// Preparar la petición para convertir texto a voz
	request := &TextToSpeechRequest{
		Model:          model,
		Input:          input,
		Voice:          voice,
		ResponseFormat: responseFormat,
		Speed:          speed,
	}

	// Add instructions if provided and using gpt-4o-mini-tts model
	if instructions, ok := d.GetOk("instructions"); ok {
		if model == "gpt-4o-mini-tts" {
			request.Instructions = instructions.(string)
		} else {
			return diag.FromErr(fmt.Errorf("instructions parameter only works with gpt-4o-mini-tts model"))
		}
	}

	// Preparar la petición HTTP
	// Use the base API URL without adding "/v1" since it's already included
	url := fmt.Sprintf("%s/audio/speech", openaiClient.APIURL)
	fmt.Printf("[DEBUG] Making request to URL: %s\n", url)

	reqBody, err := json.Marshal(request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing text-to-speech request: %v", err))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Establecer headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openaiClient.APIKey)

	// Añadir Organization ID si está presente
	if openaiClient.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", openaiClient.OrganizationID)
	}

	// Realizar la petición
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %v", err))
	}
	defer resp.Body.Close()

	// Verificar si hubo un error
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %v, status code: %d, body: %s",
				err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error generating speech: %s - %s",
			errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Crear el directorio de salida si no existe
	outputDir := filepath.Dir(outputFilePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return diag.FromErr(fmt.Errorf("error creating output directory: %v", err))
	}

	// Crear el archivo de salida
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating output file: %v", err))
	}
	defer outputFile.Close()

	// Copiar el contenido de la respuesta al archivo
	if _, err := io.Copy(outputFile, resp.Body); err != nil {
		return diag.FromErr(fmt.Errorf("error writing audio data to file: %v", err))
	}

	// Generar un ID único para este recurso
	speechID := fmt.Sprintf("speech-%d", time.Now().UnixNano())
	d.SetId(speechID)

	// Establecer los valores computados
	if err := d.Set("created_at", time.Now().Unix()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
	}

	return diag.Diagnostics{}
}

// resourceOpenAITextToSpeechRead retrieves the current state of a speech synthesis.
// It fetches the latest information about the synthesis and updates the Terraform state.
// Note: Synthesized speech is immutable, so this function only verifies existence.
func resourceOpenAITextToSpeechRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	outputFilePath := d.Get("output_file").(string)

	// During import, we need to make sure to set reasonable values for required fields
	if outputFilePath == "" {
		outputFilePath = "./output/speech.mp3"
		_ = d.Set("output_file", outputFilePath)
	}

	// Check if the file exists
	fileExists := true
	if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
		// File doesn't exist, but we'll continue with the import
		fileExists = false
	}

	// In case of import, we need to set default values for required fields
	if d.Get("model") == "" {
		_ = d.Set("model", "tts-1")
	}

	if d.Get("input") == "" {
		_ = d.Set("input", "Hello, this is a sample text for speech synthesis.")
	}

	if d.Get("voice") == "" {
		_ = d.Set("voice", "alloy")
	}

	if d.Get("response_format") == "" {
		_ = d.Set("response_format", "mp3")
	}

	if d.Get("speed") == 0.0 {
		_ = d.Set("speed", 1.0)
	}

	// Set created_at if not already set
	if d.Get("created_at").(int) == 0 {
		// Try to extract timestamp from ID or use current time
		_ = d.Set("created_at", time.Now().Unix())
	}

	// If the file doesn't exist and we're not importing, mark the resource as gone
	if !fileExists && !d.IsNewResource() {
		d.SetId("")
	}

	return nil
}

// resourceOpenAITextToSpeechDelete removes synthesized speech.
// Note: Synthesized speech cannot be deleted through the API.
// This function only removes the resource from the Terraform state.
func resourceOpenAITextToSpeechDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Para text-to-speech, la operación de "eliminación" simplemente limpia el estado
	// Opcionalmente, podríamos eliminar el archivo de audio generado

	outputFilePath := d.Get("output_file").(string)

	// Eliminar el archivo de audio si existe
	if _, err := os.Stat(outputFilePath); err == nil {
		if err := os.Remove(outputFilePath); err != nil {
			return diag.FromErr(fmt.Errorf("error removing audio file: %v", err))
		}
	}

	d.SetId("")
	return nil
}
