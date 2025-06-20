package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOpenAIVectorStoreFileBatch() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIVectorStoreFileBatchRead,
		Schema: map[string]*schema.Schema{
			"vector_store_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the vector store.",
			},
			"batch_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the file batch to retrieve details for.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the file batch.",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type (always 'vector_store.file_batch').",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the file batch was created.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the file batch.",
			},
			"file_ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of file IDs in the batch.",
			},
			"batch_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the batch.",
			},
			"purpose": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The purpose of the file batch.",
			},
		},
	}
}

func dataSourceOpenAIVectorStoreFileBatchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)
	batchID := d.Get("batch_id").(string)

	// Prepare request URL
	url := fmt.Sprintf("/v1/vector_stores/%s/file_batches/%s", vectorStoreID, batchID)

	// Make API request
	responseBytes, err := client.DoRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading vector store file batch %s: %s", batchID, err))
	}

	// Parse response
	var batch map[string]interface{}
	if err := json.Unmarshal(responseBytes, &batch); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set ID
	if id, ok := batch["id"].(string); ok {
		d.SetId(id)
	} else {
		return diag.FromErr(fmt.Errorf("batch ID not found in response"))
	}

	// Set basic attributes
	if err := d.Set("object", batch["object"]); err != nil {
		return diag.FromErr(fmt.Errorf("error setting object: %s", err))
	}

	if err := d.Set("status", batch["status"]); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %s", err))
	}

	if err := d.Set("batch_type", batch["batch_type"]); err != nil {
		return diag.FromErr(fmt.Errorf("error setting batch_type: %s", err))
	}

	if err := d.Set("purpose", batch["purpose"]); err != nil {
		return diag.FromErr(fmt.Errorf("error setting purpose: %s", err))
	}

	// Set numeric attributes
	if createdAt, ok := batch["created_at"].(float64); ok {
		if err := d.Set("created_at", int(createdAt)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting created_at: %s", err))
		}
	}

	// Set file_ids
	if fileIDs, ok := batch["file_ids"].([]interface{}); ok {
		ids := make([]string, 0, len(fileIDs))
		for _, id := range fileIDs {
			if strID, ok := id.(string); ok {
				ids = append(ids, strID)
			}
		}
		if err := d.Set("file_ids", ids); err != nil {
			return diag.FromErr(fmt.Errorf("error setting file_ids: %s", err))
		}
	}

	return diag.Diagnostics{}
}
