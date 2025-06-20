package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ModelInfoResponse represents the API response for an OpenAI model info endpoint
type ModelInfoResponse struct {
	ID         string            `json:"id"`
	Object     string            `json:"object"`
	Created    int               `json:"created"`
	OwnedBy    string            `json:"owned_by"`
	Permission []ModelPermission `json:"permission"`
}

// ModelPermission represents the permission details for a model
type ModelPermission struct {
	ID                 string      `json:"id"`
	Object             string      `json:"object"`
	Created            int         `json:"created"`
	AllowCreateEngine  bool        `json:"allow_create_engine"`
	AllowSampling      bool        `json:"allow_sampling"`
	AllowLogprobs      bool        `json:"allow_logprobs"`
	AllowSearchIndices bool        `json:"allow_search_indices"`
	AllowView          bool        `json:"allow_view"`
	AllowFineTuning    bool        `json:"allow_fine_tuning"`
	Organization       string      `json:"organization"`
	Group              interface{} `json:"group"`
	IsBlocking         bool        `json:"is_blocking"`
}

// resourceOpenAIModel defines the schema and CRUD operations for OpenAI models.
// This allows users to manage OpenAI models through Terraform.
func resourceOpenAIModel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIModelCreate,
		ReadContext:   resourceOpenAIModelRead,
		DeleteContext: resourceOpenAIModelDelete,
		Schema: map[string]*schema.Schema{
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the model (e.g., gpt-4, gpt-3.5-turbo, etc.)",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Project API key for authentication. If not provided, the provider's default API key will be used.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Always suppress the diff for the API key
					return true
				},
				// This ensures the API key never gets stored in the state file
				StateFunc: func(val interface{}) string {
					// Return empty string instead of the actual API key
					return ""
				},
			},
			"owned_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization that owns the model",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the model was created",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type (always 'model')",
			},
			"permission": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the permission",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The object type (always 'model_permission')",
						},
						"created": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp when the permission was created",
						},
						"allow_create_engine": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the user can create engines with this model",
						},
						"allow_sampling": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the user can sample from this model",
						},
						"allow_logprobs": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the user can request log probabilities from this model",
						},
						"allow_search_indices": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the user can use this model in search indices",
						},
						"allow_view": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the user can view this model",
						},
						"allow_fine_tuning": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the user can fine-tune this model",
						},
						"organization": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The organization this permission applies to",
						},
						"group": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group this permission applies to",
						},
						"is_blocking": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this permission is blocking",
						},
					},
				},
			},
		},
	}
}

// resourceOpenAIModelCreate handles the creation of a new model in OpenAI.
// Since OpenAI API doesn't allow direct model creation, this function mostly sets
// the model ID and performs a read to populate the remaining fields.
func resourceOpenAIModelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the model ID from the configuration
	modelID := d.Get("model").(string)
	if modelID == "" {
		return diag.Errorf("model ID cannot be empty")
	}

	// Set the ID to the model ID
	d.SetId(modelID)

	// Read the model to populate the state
	return resourceOpenAIModelRead(ctx, d, m)
}

// resourceOpenAIModelRead reads the model information from OpenAI API
func resourceOpenAIModelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClientWithProjectKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the model ID from state
	modelID := d.Id()
	if modelID == "" {
		return diag.Errorf("model ID is empty")
	}

	// Get custom API key if provided
	apiKey := client.APIKey
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/models/%s", client.APIURL, modelID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/models/%s", client.APIURL, modelID)
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Add headers with the appropriate API key
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// For 404 errors, we remove the resource from state
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, errorResponse.Error.Message))
	}

	// Parse the response
	var model ModelInfoResponse
	err = json.Unmarshal(responseBody, &model)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update the state
	if err := d.Set("owned_by", model.OwnedBy); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", model.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("object", model.Object); err != nil {
		return diag.FromErr(err)
	}

	// Set permissions
	permissions := make([]map[string]interface{}, 0, len(model.Permission))
	for _, p := range model.Permission {
		permMap := map[string]interface{}{
			"id":                   p.ID,
			"object":               p.Object,
			"created":              p.Created,
			"allow_create_engine":  p.AllowCreateEngine,
			"allow_sampling":       p.AllowSampling,
			"allow_logprobs":       p.AllowLogprobs,
			"allow_search_indices": p.AllowSearchIndices,
			"allow_view":           p.AllowView,
			"allow_fine_tuning":    p.AllowFineTuning,
			"organization":         p.Organization,
			"group":                p.Group,
			"is_blocking":          p.IsBlocking,
		}
		permissions = append(permissions, permMap)
	}
	if err := d.Set("permission", permissions); err != nil {
		return diag.FromErr(err)
	}

	// Explicitly set the api_key to empty in the state
	if err := d.Set("api_key", ""); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIModelDelete removes the OpenAI model.
func resourceOpenAIModelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// In reality, users can't delete OpenAI models, they can just stop using them
	// So we'll just clear the ID from the terraform state

	d.SetId("")
	return diag.Diagnostics{}
}

// createOpenAIModel is a helper function that makes the API call to create a model.
// This is currently a placeholder implementation.
func createOpenAIModel(apiKey, name string) (string, error) {
	// Implement the API call to create a model and return its ID.
	fmt.Printf("Creating model %s with API key %s\n", name, apiKey)
	return "new-model-id", nil
}
