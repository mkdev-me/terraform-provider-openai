package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Model represents an OpenAI model
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// dataSourceOpenAIModel returns a schema.Resource that represents a data source for a single OpenAI model.
// This data source allows users to retrieve detailed information about a specific OpenAI model by its ID.
func dataSourceOpenAIModel() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIModelRead,
		Schema: map[string]*schema.Schema{
			"model_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the model to retrieve information for",
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
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type, which is always 'model'",
			},
		},
	}
}

// dataSourceOpenAIModelRead handles the read operation for the OpenAI model data source.
// It retrieves information about a specific model from the OpenAI API and updates the Terraform state.
func dataSourceOpenAIModelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClientWithProjectKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	modelID := d.Get("model_id").(string)
	url := fmt.Sprintf("models/%s", modelID)

	respBody, reqErr := client.DoRequest(http.MethodGet, url, nil)

	if reqErr != nil {
		return diag.Errorf("Error retrieving model: %s", reqErr)
	}

	var model Model
	if err := json.Unmarshal(respBody, &model); err != nil {
		return diag.Errorf("Error parsing model response: %s", err)
	}

	d.SetId(modelID)

	if err := d.Set("created", model.Created); err != nil {
		return diag.Errorf("Error setting created: %s", err)
	}

	if err := d.Set("owned_by", model.OwnedBy); err != nil {
		return diag.Errorf("Error setting owned_by: %s", err)
	}

	if err := d.Set("object", model.Object); err != nil {
		return diag.Errorf("Error setting object: %s", err)
	}

	return nil
}
