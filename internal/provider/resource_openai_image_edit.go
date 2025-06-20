package provider

import (
	"bytes"
	"context"
	"encoding/base64"
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

// ImageEditResponse represents the API response for image editing.
// It contains the edited images and metadata about the editing process.
// This structure provides access to both URL and base64-encoded image data.
type ImageEditResponse struct {
	Created int             `json:"created"` // Unix timestamp of image creation
	Data    []ImageEditData `json:"data"`    // List of edited images
}

// ImageEditData represents a single edited image.
// It contains the edited image data in either URL or base64 format.
type ImageEditData struct {
	URL     string `json:"url,omitempty"`      // URL to the edited image
	B64JSON string `json:"b64_json,omitempty"` // Base64-encoded image data
}

// resourceOpenAIImageEdit defines the schema and CRUD operations for OpenAI image editing.
// This resource allows users to edit existing images using OpenAI's models.
// It supports various editing options including masking and provides control over the output format.
func resourceOpenAIImageEdit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIImageEditCreate,
		ReadContext:   resourceOpenAIImageEditRead,
		DeleteContext: resourceOpenAIImageEditDelete,
		Schema: map[string]*schema.Schema{
			"image": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The image to edit. Must be a valid PNG file, less than 4MB, and square.",
			},
			"mask": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "An additional image whose fully transparent areas indicate where 'image' should be edited.",
			},
			"model": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "dall-e-2",
				Description: "The model to use for image editing.",
			},
			"prompt": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A text description of the desired image(s).",
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
				Description: "The timestamp for when the edited image was created.",
			},
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL of the edited image (if response_format is 'url').",
						},
						"b64_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The base64-encoded JSON of the edited image (if response_format is 'b64_json').",
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIImageEditImportState,
		},
	}
}

// resourceOpenAIImageEditCreate handles the creation of an image edit request.
// It processes the edit parameters and sends them to the OpenAI API.
func resourceOpenAIImageEditCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client using the GetOpenAIClient helper function
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %v", err))
	}

	// Obtener los parámetros de entrada del schema
	imagePath := d.Get("image").(string)
	prompt := d.Get("prompt").(string)
	n := d.Get("n").(int)
	size := d.Get("size").(string)
	responseFormat := d.Get("response_format").(string)
	model := d.Get("model").(string)

	// Verificar que el archivo de imagen existe
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return diag.FromErr(fmt.Errorf("image file does not exist: %s", imagePath))
	}

	// Crear un buffer para la petición multipart
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Añadir el campo de prompt
	if err := writer.WriteField("prompt", prompt); err != nil {
		return diag.FromErr(fmt.Errorf("error writing prompt field: %v", err))
	}

	// Añadir el campo de n
	if err := writer.WriteField("n", fmt.Sprintf("%d", n)); err != nil {
		return diag.FromErr(fmt.Errorf("error writing n field: %v", err))
	}

	// Añadir el campo de size
	if err := writer.WriteField("size", size); err != nil {
		return diag.FromErr(fmt.Errorf("error writing size field: %v", err))
	}

	// Añadir el campo de response_format
	if err := writer.WriteField("response_format", responseFormat); err != nil {
		return diag.FromErr(fmt.Errorf("error writing response_format field: %v", err))
	}

	// Añadir el campo de model si está presente
	if model != "" {
		if err := writer.WriteField("model", model); err != nil {
			return diag.FromErr(fmt.Errorf("error writing model field: %v", err))
		}
	}

	// Añadir el campo de user si está presente
	if user, ok := d.GetOk("user"); ok {
		if err := writer.WriteField("user", user.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("error writing user field: %v", err))
		}
	}

	// Añadir el archivo de imagen
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

	// Añadir el archivo de máscara si está presente
	if maskPath, ok := d.GetOk("mask"); ok {
		// Verificar que el archivo de máscara existe
		if _, err := os.Stat(maskPath.(string)); os.IsNotExist(err) {
			return diag.FromErr(fmt.Errorf("mask file does not exist: %s", maskPath.(string)))
		}

		maskFile, err := os.Open(maskPath.(string))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error opening mask file: %v", err))
		}
		defer maskFile.Close()

		// Create a custom form field with explicit content type for PNG
		mh := make(textproto.MIMEHeader)
		mh.Set("Content-Disposition", fmt.Sprintf(`form-data; name="mask"; filename="%s"`, filepath.Base(maskPath.(string))))
		mh.Set("Content-Type", "image/png")
		maskPart, err := writer.CreatePart(mh)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating form file for mask: %v", err))
		}

		if _, err := io.Copy(maskPart, maskFile); err != nil {
			return diag.FromErr(fmt.Errorf("error copying mask data: %v", err))
		}
	}

	// Cerrar el escritor multipart
	if err := writer.Close(); err != nil {
		return diag.FromErr(fmt.Errorf("error closing multipart writer: %v", err))
	}

	// Preparar la petición HTTP
	url := fmt.Sprintf("%s/images/edits", client.APIURL)
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
		return diag.FromErr(fmt.Errorf("error editing image: %s - %s",
			errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parsear la respuesta
	var editResponse ImageEditResponse
	if err := json.Unmarshal(respBody, &editResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	// Generar un ID único para este recurso
	// En edición de imágenes no se devuelve un ID específico, así que creamos uno basado en el timestamp
	imageEditID := fmt.Sprintf("img-edit-%d", editResponse.Created)
	d.SetId(imageEditID)

	// Establecer el timestamp de creación
	if err := d.Set("created", editResponse.Created); err != nil {
		return diag.FromErr(err)
	}

	// Procesar los datos de las imágenes editadas
	if len(editResponse.Data) > 0 {
		imageData := make([]map[string]interface{}, len(editResponse.Data))

		for i, img := range editResponse.Data {
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

// resourceOpenAIImageEditRead handles the reading of an OpenAI image edit resource.
// It is also used for importing resources by ID.
func resourceOpenAIImageEditRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// OpenAI doesn't provide an API to read image edits by ID
	// So we just return nil to maintain the current state
	return nil
}

// resourceOpenAIImageEditDelete handles the deletion logic for an OpenAI image edit.
// Note that this is a no-op as OpenAI doesn't provide API endpoints to delete edited images.
// The resource is simply removed from Terraform state.
func resourceOpenAIImageEditDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Set empty ID to indicate the resource no longer exists
	d.SetId("")
	return nil
}

// resourceOpenAIImageEditImportState handles the import of an image edit resource
// It only sets the ID and created timestamp, allowing other values to come from configuration
func resourceOpenAIImageEditImportState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ",")
	if len(parts) < 2 {
		return nil, fmt.Errorf("must provide parameters for import: id,prompt=...,image=...,model=..., etc.")
	}

	d.SetId(parts[0])

	for _, param := range parts[1:] {
		paramParts := strings.SplitN(param, "=", 2)
		if len(paramParts) != 2 {
			return nil, fmt.Errorf("invalid parameter format: %s", param)
		}
		key, value := paramParts[0], paramParts[1]

		switch key {
		case "prompt", "image", "mask", "model", "response_format", "size", "user":
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

// fileToBase64 converts a file to its base64-encoded string representation.
// This utility function is used to prepare image data for API requests.
// It handles file reading and encoding, with proper error handling.
func fileToBase64(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(fileContent), nil
}
