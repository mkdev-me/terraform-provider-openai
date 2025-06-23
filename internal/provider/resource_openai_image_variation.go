package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ImageVariationResponse represents the API response for image variations.
// It contains the generated variations and metadata about the variation process.
// This structure provides access to both URL and base64-encoded image data.
type ImageVariationResponse struct {
	Created int                  `json:"created"` // Unix timestamp of variation creation
	Data    []ImageVariationData `json:"data"`    // List of generated variations
}

// ImageVariationData represents a single image variation.
// It contains the variation data in either URL or base64 format.
type ImageVariationData struct {
	URL     string `json:"url,omitempty"`      // URL to the variation image
	B64JSON string `json:"b64_json,omitempty"` // Base64-encoded image data
}

// resourceOpenAIImageVariation defines the schema and CRUD operations for OpenAI image variations.
// This resource allows users to generate variations of existing images using OpenAI's models.
// It provides control over the variation process and supports various output formats.
func resourceOpenAIImageVariation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIImageVariationCreate,
		ReadContext:   resourceOpenAIImageVariationRead,
		DeleteContext: resourceOpenAIImageVariationDelete,
		Schema: map[string]*schema.Schema{
			"image": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The image to use as the basis for the variation(s). Must be a valid PNG file, less than 4MB, and square.",
			},
			"model": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "dall-e-2",
				Description: "The model to use for image variation.",
			},
			"n": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 10),
				Description:  "The number of images to generate. Must be between 1 and 10.",
			},
			"size": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "1024x1024",
				ValidateFunc: validation.StringInSlice([]string{"256x256", "512x512", "1024x1024"}, false),
				Description:  "The size of the generated images.",
			},
			"response_format": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "url",
				ValidateFunc: validation.StringInSlice([]string{"url", "b64_json"}, false),
				Description:  "The format in which the generated images are returned.",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "A unique identifier representing your end-user.",
			},
			"created": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the varied image was created.",
			},
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL of the varied image (if response_format is 'url').",
						},
						"b64_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The base64-encoded JSON of the varied image (if response_format is 'b64_json').",
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIImageVariationImportState,
		},
	}
}

// resourceOpenAIImageVariationCreate handles the creation of a new image variation request.
// It processes the parameters and sends them to the OpenAI API.
func resourceOpenAIImageVariationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client using the GetOpenAIClient helper function
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %v", err))
	}

	// Get the input parameters from the schema
	imagePath := d.Get("image").(string)
	n := d.Get("n").(int)
	size := d.Get("size").(string)
	responseFormat := d.Get("response_format").(string)
	model := d.Get("model").(string)

	// Check if the image file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return diag.FromErr(fmt.Errorf("image file does not exist: %s", imagePath))
	}

	// Create a buffer for the multipart request
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the n field
	if err := writer.WriteField("n", fmt.Sprintf("%d", n)); err != nil {
		return diag.FromErr(fmt.Errorf("error writing n field: %v", err))
	}

	// Add the size field
	if err := writer.WriteField("size", size); err != nil {
		return diag.FromErr(fmt.Errorf("error writing size field: %v", err))
	}

	// Add the response_format field
	if err := writer.WriteField("response_format", responseFormat); err != nil {
		return diag.FromErr(fmt.Errorf("error writing response_format field: %v", err))
	}

	// Add the model field if present
	if model != "" {
		if err := writer.WriteField("model", model); err != nil {
			return diag.FromErr(fmt.Errorf("error writing model field: %v", err))
		}
	}

	// Add the user field if present
	if user, ok := d.GetOk("user"); ok {
		if err := writer.WriteField("user", user.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("error writing user field: %v", err))
		}
	}

	// Add the image file
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error opening image file: %v", err))
	}
	defer imageFile.Close()

	// Create a custom form field with explicit content type for PNG
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, filepath.Base(imagePath)))
	h.Set("Content-Type", "image/png")
	imagePart, err := writer.CreatePart(h)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating form file: %v", err))
	}

	if _, err := io.Copy(imagePart, imageFile); err != nil {
		return diag.FromErr(fmt.Errorf("error copying image data: %v", err))
	}

	// Cerrar el escritor multipart
	if err := writer.Close(); err != nil {
		return diag.FromErr(fmt.Errorf("error closing multipart writer: %v", err))
	}

	// Prepare the HTTP request
	url := fmt.Sprintf("%s/images/variations", client.APIURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
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
		return diag.FromErr(fmt.Errorf("error creating image variation: %s - %s",
			errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var variationResponse ImageVariationResponse
	if err := json.Unmarshal(respBody, &variationResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	// Generar un ID único para este recurso
	// En variación de imágenes no se devuelve un ID específico, así que creamos uno basado en el timestamp
	imageVariationID := fmt.Sprintf("img-var-%d", variationResponse.Created)
	d.SetId(imageVariationID)

	// Set the creation timestamp
	if err := d.Set("created", variationResponse.Created); err != nil {
		return diag.FromErr(err)
	}

	// Process the generated image data
	if len(variationResponse.Data) > 0 {
		imageData := make([]map[string]interface{}, len(variationResponse.Data))

		for i, img := range variationResponse.Data {
			imageResult := map[string]interface{}{}

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

// resourceOpenAIImageVariationRead handles the reading of an OpenAI image variation resource.
// It is also used for importing resources by ID.
func resourceOpenAIImageVariationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// OpenAI doesn't provide an API to read image variations by ID
	// So we just return nil to maintain the current state
	return nil
}

// resourceOpenAIImageVariationDelete handles the deletion logic for an OpenAI image variation.
// Note that this is a no-op as OpenAI doesn't provide API endpoints to delete image variations.
// The resource is simply removed from Terraform state.
func resourceOpenAIImageVariationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Set empty ID to indicate the resource no longer exists
	d.SetId("")
	return nil
}

// resourceOpenAIImageVariationImportState handles the import of an image variation resource
// It only sets the ID and created timestamp, allowing other values to come from configuration
func resourceOpenAIImageVariationImportState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ",")
	if len(parts) < 2 {
		return nil, fmt.Errorf("must provide parameters for import: id,image=...,model=..., etc.")
	}

	d.SetId(parts[0])

	for _, param := range parts[1:] {
		paramParts := strings.SplitN(param, "=", 2)
		if len(paramParts) != 2 {
			return nil, fmt.Errorf("invalid parameter format: %s", param)
		}
		key, value := paramParts[0], paramParts[1]

		switch key {
		case "image", "model", "response_format", "size", "user":
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
