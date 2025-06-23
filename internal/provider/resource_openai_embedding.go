package provider

import (
	"context"
	"encoding/json"
	"fmt"

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
	// Get the OpenAI client
	client := meta.(*OpenAIClient)

	// Prepare the request with all fields
	request := &EmbeddingRequest{
		Model: d.Get("model").(string),
	}

	// Process the input (can be string or array of strings)
	input := d.Get("input").(string)

	// Check if the input is a JSON array
	var inputArray []string
	if err := json.Unmarshal([]byte(input), &inputArray); err == nil && len(inputArray) > 0 {
		// Si se pudo parsear como array, usarlo así
		request.Input = inputArray
	} else {
		// De lo contrario, crear un array con un solo string
		request.Input = []string{input}
	}

	// Add user if present
	if user, ok := d.GetOk("user"); ok {
		request.User = user.(string)
	}

	// Add encoding_format if present
	if format, ok := d.GetOk("encoding_format"); ok {
		request.EncodingFormat = format.(string)
	}

	// Add dimensions if present
	if dimensions, ok := d.GetOk("dimensions"); ok {
		request.Dimensions = dimensions.(int)
	}

	// Use the client's DoRequest method
	url := "/v1/embeddings"

	// Make the API request using the client's DoRequest method
	// DoRequest will handle JSON marshaling
	respBody, err := client.DoRequest("POST", url, request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating embedding: %s", err))
	}

	// Parse the response
	var embeddingResponse EmbeddingResponse
	if err := json.Unmarshal(respBody, &embeddingResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Generar un ID único para este embedding (ya que la API no proporciona uno)
	// Usamos un hash del modelo y los primeros valores de embedding
	embeddingID := fmt.Sprintf("embd_%s", embeddingResponse.Model)
	d.SetId(embeddingID)

	// Update the state with response data
	if err := d.Set("embedding_id", embeddingID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("object", embeddingResponse.Object); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("model_used", embeddingResponse.Model); err != nil {
		return diag.FromErr(err)
	}

	// Process the embeddings
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

	// Update the usage statistics
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
