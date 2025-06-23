package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ImageGenerationResponse represents the API response for image generation.
// It contains the generated images and metadata about the generation process.
// This structure provides access to both URL and base64-encoded image data.
type ImageGenerationResponse struct {
	Created int                   `json:"created"` // Unix timestamp of image creation
	Data    []ImageGenerationData `json:"data"`    // List of generated images
}

// ImageGenerationData represents a single generated image.
// It contains the image data in either URL or base64 format, along with any
// revised prompt that was used to generate the image.
type ImageGenerationData struct {
	URL           string `json:"url,omitempty"`            // URL to the generated image
	B64JSON       string `json:"b64_json,omitempty"`       // Base64-encoded image data
	RevisedPrompt string `json:"revised_prompt,omitempty"` // Modified prompt used for generation
}

// ImageGenerationRequest represents the request payload for generating images.
// It specifies the parameters that control the image generation process,
// including model selection, prompt, and various generation options.
type ImageGenerationRequest struct {
	Model          string `json:"model,omitempty"`           // ID of the model to use
	Prompt         string `json:"prompt"`                    // Text description of desired image
	N              int    `json:"n,omitempty"`               // Number of images to generate
	Quality        string `json:"quality,omitempty"`         // Quality level of generated images
	ResponseFormat string `json:"response_format,omitempty"` // Format of the response (url or b64_json)
	Size           string `json:"size,omitempty"`            // Dimensions of generated images
	Style          string `json:"style,omitempty"`           // Style to apply to generated images
	User           string `json:"user,omitempty"`            // Optional user identifier
}

// resourceOpenAIImageGeneration defines the schema and CRUD operations for OpenAI image generation.
// This resource allows users to generate images from text descriptions using OpenAI's models.
// It provides comprehensive control over the generation process and supports various output formats.
func resourceOpenAIImageGeneration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIImageGenerationCreate,
		ReadContext:   resourceOpenAIImageGenerationRead,
		DeleteContext: resourceOpenAIImageGenerationDelete,
		Schema: map[string]*schema.Schema{
			"prompt": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A text description of the desired image(s)",
			},
			"model": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "dall-e-3",
				Description: "The model to use for image generation",
			},
			"n": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 10),
				Description:  "The number of images to generate. Must be between 1 and 10.",
			},
			"quality": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "standard",
				ValidateFunc: validation.StringInSlice([]string{"standard", "hd"}, false),
				Description:  "The quality of the image that will be generated. Can be 'standard' or 'hd'.",
			},
			"response_format": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "url",
				ValidateFunc: validation.StringInSlice([]string{"url", "b64_json"}, false),
				Description:  "The format in which the generated images are returned. Can be 'url' or 'b64_json'.",
			},
			"size": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "1024x1024",
				ValidateFunc: validation.StringInSlice([]string{"256x256", "512x512", "1024x1024", "1792x1024", "1024x1792"}, false),
				Description:  "The size of the generated images",
			},
			"style": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "vivid",
				ValidateFunc: validation.StringInSlice([]string{"vivid", "natural"}, false),
				Description:  "The style of the generated images. Can be 'vivid' or 'natural'.",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "A unique identifier representing your end-user",
			},
			"created": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the image was created",
			},
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL of the generated image (if response_format is 'url')",
						},
						"b64_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The base64-encoded JSON of the generated image (if response_format is 'b64_json')",
						},
						"revised_prompt": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The prompt that was used to generate the image, potentially modified from the original prompt",
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIImageGenerationImportState,
		},
	}
}

// resourceOpenAIImageGenerationCreate handles the creation of a new image generation request.
// It processes the generation parameters and sends them to the OpenAI API.
func resourceOpenAIImageGenerationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client using the GetOpenAIClient helper function
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %v", err))
	}

	// Get the input parameters from the schema
	prompt := d.Get("prompt").(string)
	model := d.Get("model").(string)
	n := d.Get("n").(int)
	quality := d.Get("quality").(string)
	responseFormat := d.Get("response_format").(string)
	size := d.Get("size").(string)
	style := d.Get("style").(string)

	// Prepare the request to generate images
	request := &ImageGenerationRequest{
		Model:          model,
		Prompt:         prompt,
		N:              n,
		Quality:        quality,
		ResponseFormat: responseFormat,
		Size:           size,
		Style:          style,
	}

	// Add user if present
	if user, ok := d.GetOk("user"); ok {
		request.User = user.(string)
	}

	// Prepare the HTTP request
	url := fmt.Sprintf("%s/images/generations", client.APIURL)
	reqBody, err := json.Marshal(request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing image generation request: %v", err))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	// Add Organization ID if present
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %v", err))
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %v", err))
	}

	// Check if there was an error
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %v, status code: %d, body: %s",
				err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error generating images: %s - %s",
			errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var genResponse ImageGenerationResponse
	if err := json.Unmarshal(respBody, &genResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	// Generar un ID único para este recurso
	// En generación de imágenes no se devuelve un ID específico, así que creamos uno basado en el timestamp
	imageGenID := fmt.Sprintf("img-%d", genResponse.Created)
	d.SetId(imageGenID)

	// Set the creation timestamp
	if err := d.Set("created", genResponse.Created); err != nil {
		return diag.FromErr(err)
	}

	// Process the generated image data
	if len(genResponse.Data) > 0 {
		imageData := make([]map[string]interface{}, len(genResponse.Data))

		for i, img := range genResponse.Data {
			imageResult := map[string]interface{}{}

			if img.RevisedPrompt != "" {
				imageResult["revised_prompt"] = img.RevisedPrompt
			} else {
				imageResult["revised_prompt"] = prompt // Usar el prompt original si no hay revised_prompt
			}

			if responseFormat == "url" {
				imageResult["url"] = img.URL
				imageResult["b64_json"] = ""
			} else {
				imageResult["b64_json"] = img.B64JSON
				imageResult["url"] = ""
			}

			imageData[i] = imageResult
		}

		if err := d.Set("data", imageData); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIImageGenerationRead handles the reading of an OpenAI image generation resource.
// OpenAI doesn't provide an API to retrieve image generation details by ID after creation.
func resourceOpenAIImageGenerationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// LIMITATION: OpenAI's API doesn't allow retrieving image generation details by ID.
	// This means we cannot refresh the state with current API data.

	// Verify that required fields are present
	if _, ok := d.GetOk("prompt"); !ok {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Missing required attribute",
				Detail:   "The 'prompt' attribute is required but not set. Please add it to your configuration.",
			},
		}
	}

	// Just maintain the existing state since we can't refresh from the API
	return nil
}

// resourceOpenAIImageGenerationDelete handles the deletion logic for an OpenAI image generation resource.
// Note that this is a no-op as OpenAI doesn't provide API endpoints to delete generated images.
// The resource is simply removed from Terraform state.
func resourceOpenAIImageGenerationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// No deletion API call needed as OpenAI doesn't provide a way to delete generated images
	d.SetId("")
	return nil
}

// resourceOpenAIImageGenerationImportState handles the import of an image generation resource.
// Note: OpenAI's API does not provide a way to retrieve details of previously generated images by ID.
// This function sets the ID and timestamp from the import, and other values must be provided in configuration.
func resourceOpenAIImageGenerationImportState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ",")
	if len(parts) < 2 {
		return nil, fmt.Errorf("must provide parameters for import: id,prompt=...,model=..., etc.")
	}

	// Set the resource ID
	d.SetId(parts[0])

	// Iterate over the additional parameters to set resource attributes
	for _, param := range parts[1:] {
		paramParts := strings.SplitN(param, "=", 2)
		if len(paramParts) != 2 {
			return nil, fmt.Errorf("invalid parameter format: %s", param)
		}
		key, value := paramParts[0], paramParts[1]

		switch key {
		case "prompt", "model", "quality", "response_format", "size", "style", "user":
			if err := d.Set(key, value); err != nil {
				return nil, fmt.Errorf("error setting %s: %v", key, err)
			}
		case "n", "created":
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid integer value for %s: %v", key, err)
			}
			if err := d.Set(key, intValue); err != nil {
				return nil, fmt.Errorf("error setting %s: %v", key, err)
			}
		default:
			return nil, fmt.Errorf("unknown parameter: %s", key)
		}
	}

	return []*schema.ResourceData{d}, nil
}
