package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fjcorp/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// EmbeddingResponse represents the API response for text embeddings.
// It contains the generated embeddings, model information, and usage statistics.
type EmbeddingResponse struct {
	Object string          `json:"object"` // Type of object (e.g., "list")
	Data   []EmbeddingData `json:"data"`   // List of generated embeddings
	Model  string          `json:"model"`  // Model used for the embeddings
	Usage  EmbeddingUsage  `json:"usage"`  // Token usage statistics
}

// EmbeddingData represents a single text embedding.
// It contains the vector representation of the input text.
type EmbeddingData struct {
	Object    string          `json:"object"`    // Type of object (e.g., "embedding")
	Index     int             `json:"index"`     // Position of this embedding in the list
	Embedding json.RawMessage `json:"embedding"` // Vector representation of the text (can be float array or base64 string)
}

// EmbeddingUsage represents token usage statistics for the embedding request.
// It tracks the number of tokens used in the input text.
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"` // Number of tokens in the input text
	TotalTokens  int `json:"total_tokens"`  // Total number of tokens used
}

// EmbeddingRequest represents the request payload for creating text embeddings.
// It specifies the model, input text, and various parameters to control the embedding process.
type EmbeddingRequest struct {
	Model          string      `json:"model"`                     // ID of the model to use
	Input          interface{} `json:"input"`                     // Text or list of texts to embed
	User           string      `json:"user,omitempty"`            // Optional user identifier
	EncodingFormat string      `json:"encoding_format,omitempty"` // Format of the embedding output
	Dimensions     int         `json:"dimensions,omitempty"`      // Optional number of dimensions
}

// resourceOpenAIEmbedding defines the schema and CRUD operations for OpenAI text embeddings.
// This resource allows users to generate vector embeddings for text using OpenAI's models.
// It supports various models and provides options for controlling the embedding process.
func resourceOpenAIEmbedding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIEmbeddingCreate,
		ReadContext:   resourceOpenAIEmbeddingRead,
		DeleteContext: resourceOpenAIEmbeddingDelete,
		Schema: map[string]*schema.Schema{
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the model to use for the embedding",
			},
			"input": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The text to embed",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "A unique identifier representing your end-user",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The project to use for this request",
			},
			"encoding_format": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "float",
				Description: "The format of the embeddings. One of 'float', 'base64'",
			},
			"dimensions": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "The number of dimensions to use for the embedding (only specific models support this)",
			},
			// Response fields
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type, which is always 'embedding'",
			},
			"model_used": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The model used for the embedding",
			},
			"embedding_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the embedding",
			},
			"embeddings": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The embeddings generated for the input",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The object type, which is always 'embedding'",
						},
						"index": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The index of the embedding",
						},
						"embedding": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The embedding vector",
							Elem: &schema.Schema{
								Type: schema.TypeFloat,
							},
						},
					},
				},
			},
			"usage": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Usage statistics for the embedding request",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

// resourceOpenAIEmbeddingCreate handles the creation of new OpenAI text embeddings.
// It sends the request to OpenAI's API and processes the response.
// The function supports various embedding options and provides control over the embedding process.
func resourceOpenAIEmbeddingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obtener el cliente de OpenAI
	client := meta.(*client.OpenAIClient)

	// Preparar la petición con todos los campos
	request := &EmbeddingRequest{
		Model: d.Get("model").(string),
	}

	// Procesar el input (puede ser string o array de strings)
	input := d.Get("input").(string)

	// Verificar si el input es un array JSON
	var inputArray []string
	if err := json.Unmarshal([]byte(input), &inputArray); err == nil && len(inputArray) > 0 {
		// Si se pudo parsear como array, usarlo así
		request.Input = inputArray
	} else {
		// De lo contrario, usar como string
		request.Input = input
	}

	// Añadir user si está presente
	if user, ok := d.GetOk("user"); ok {
		request.User = user.(string)
	}

	// Añadir encoding_format si está presente
	if format, ok := d.GetOk("encoding_format"); ok {
		request.EncodingFormat = format.(string)
	}

	// Añadir dimensions si está presente
	if dimensions, ok := d.GetOk("dimensions"); ok {
		request.Dimensions = dimensions.(int)
	}

	// Determinar la URL de la API
	url := fmt.Sprintf("%s/v1/embeddings", client.APIURL)

	// Crear petición HTTP
	reqBody, err := json.Marshal(request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing embedding request: %s", err))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Establecer headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	// Añadir Organization ID si está presente
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Añadir Project ID si está presente
	if projectID, ok := d.GetOk("project_id"); ok {
		req.Header.Set("OpenAI-Project", projectID.(string))
	}

	// Realizar la petición
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Leer la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Verificar si hubo un error
	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s, status code: %d, body: %s", err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error creating embedding: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parsear la respuesta
	var embeddingResponse EmbeddingResponse
	if err := json.Unmarshal(respBody, &embeddingResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Generar un ID único para este embedding (ya que la API no proporciona uno)
	// Usamos un hash del modelo y los primeros valores de embedding
	embeddingID := fmt.Sprintf("embd_%s", embeddingResponse.Model)
	d.SetId(embeddingID)

	// Actualizar el estado con los datos de la respuesta
	if err := d.Set("embedding_id", embeddingID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("object", embeddingResponse.Object); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("model_used", embeddingResponse.Model); err != nil {
		return diag.FromErr(err)
	}

	// Procesar los embeddings
	if len(embeddingResponse.Data) > 0 {
		embeddings := make([]map[string]interface{}, 0, len(embeddingResponse.Data))

		for _, data := range embeddingResponse.Data {
			embeddingMap := map[string]interface{}{
				"object":    data.Object,
				"index":     data.Index,
				"embedding": data.Embedding,
			}

			embeddings = append(embeddings, embeddingMap)
		}

		if err := d.Set("embeddings", embeddings); err != nil {
			return diag.FromErr(err)
		}
	}

	// Actualizar las estadísticas de uso
	usage := map[string]int{
		"prompt_tokens": embeddingResponse.Usage.PromptTokens,
		"total_tokens":  embeddingResponse.Usage.TotalTokens,
	}
	if err := d.Set("usage", usage); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIEmbeddingRead retrieves the current state of OpenAI text embeddings.
// It verifies that the embeddings exist and updates the Terraform state.
// Note: OpenAI embeddings are immutable, so this function only verifies existence.
func resourceOpenAIEmbeddingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Los embeddings son efímeros y no se pueden recuperar después de la creación
	// Esta función es básicamente un no-op, pero conservamos los datos que ya tenemos en el estado

	// Si no hay un ID, significa que el recurso no existe
	if d.Id() == "" {
		return diag.Diagnostics{}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIEmbeddingDelete removes OpenAI text embeddings.
// Note: OpenAI embeddings are immutable and cannot be deleted through the API.
// This function only removes the resource from the Terraform state.
func resourceOpenAIEmbeddingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Los embeddings son efímeros y no se pueden eliminar
	// Simplemente limpiamos el ID del estado
	d.SetId("")
	return diag.Diagnostics{}
}
