package provider

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ModelsResponse represents the API response from the models endpoint
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// dataSourceOpenAIModels returns a schema.Resource that represents a data source for all available OpenAI models.
// This data source allows users to retrieve information about all models accessible with their API key.
func dataSourceOpenAIModels() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIModelsRead,
		Schema: map[string]*schema.Schema{
			"models": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the model",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The object type, which is always 'model'",
						},
						"created": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp for when the model was created",
						},
						"owned_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The organization that owns the model",
						},
					},
				},
			},
		},
	}
}

// dataSourceOpenAIModelsRead handles the read operation for the OpenAI models data source.
// It retrieves information about all available models from the OpenAI API and updates the Terraform state.
func dataSourceOpenAIModelsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClientWithProjectKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	respBody, reqErr := client.DoRequest(http.MethodGet, "models", nil)

	if reqErr != nil {
		return diag.Errorf("Error retrieving models: %s", reqErr)
	}

	var modelsResponse ModelsResponse
	if err := json.Unmarshal(respBody, &modelsResponse); err != nil {
		return diag.Errorf("Error parsing models response: %s", err)
	}

	// Set a unique ID for the data source
	d.SetId("openai-models")

	// Transform the models for the Terraform state
	tfModels := make([]map[string]interface{}, len(modelsResponse.Data))
	for i, model := range modelsResponse.Data {
		tfModels[i] = map[string]interface{}{
			"id":       model.ID,
			"object":   model.Object,
			"created":  model.Created,
			"owned_by": model.OwnedBy,
		}
	}

	if err := d.Set("models", tfModels); err != nil {
		return diag.Errorf("Error setting models: %s", err)
	}

	return nil
}
