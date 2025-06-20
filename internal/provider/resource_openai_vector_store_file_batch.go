package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOpenAIVectorStoreFileBatch() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIVectorStoreFileBatchCreate,
		ReadContext:   resourceOpenAIVectorStoreFileBatchRead,
		UpdateContext: resourceOpenAIVectorStoreFileBatchUpdate,
		DeleteContext: resourceOpenAIVectorStoreFileBatchDelete,
		Importer: &schema.ResourceImporter{
			StateContext: importVectorStoreFileBatchState,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the vector store file batch operation.",
			},
			"vector_store_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the vector store to add the files to.",
			},
			"file_ids": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of file IDs to add to the vector store.",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Set of key-value pairs that can be attached to an object. Values can be strings, booleans, or numbers.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"chunking_strategy": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ForceNew:    true,
				Description: "The chunking strategy used to chunk the files.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of chunking strategy (auto, fixed, or semantic).",
						},
						"size": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The size in characters for fixed chunking strategy.",
						},
						"max_tokens": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The maximum tokens per chunk for semantic chunking strategy.",
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the files were added to the vector store.",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type (always 'vector_store.file.batch').",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the file batch operation.",
			},
		},
	}
}

// importVectorStoreFileBatchState handles the import of a vector store file batch resource
// The ID is expected to be in the format "{vector_store_id}/{batch_id}"
func importVectorStoreFileBatchState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid ID format. Expected 'vector_store_id/batch_id', got: %s", d.Id())
	}

	vectorStoreID, batchID := parts[0], parts[1]

	// Set the component IDs
	if err := d.Set("vector_store_id", vectorStoreID); err != nil {
		return nil, err
	}

	// Set the ID to the batch_id
	d.SetId(batchID)

	return []*schema.ResourceData{d}, nil
}

func resourceOpenAIVectorStoreFileBatchCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)

	// Convert file_ids to []string
	var fileIDs []string
	if v, ok := d.GetOk("file_ids"); ok {
		for _, id := range v.([]interface{}) {
			fileIDs = append(fileIDs, id.(string))
		}
	}

	// Validate that file_ids is not empty
	if len(fileIDs) == 0 {
		return diag.Errorf("Error: file_ids cannot be empty. You must specify at least one file ID")
	}

	// Extract attributes
	var attributes map[string]interface{}
	if v, ok := d.GetOk("attributes"); ok {
		attributes = make(map[string]interface{})
		for k, attr := range v.(map[string]interface{}) {
			attributes[k] = attr
		}
	}

	// Extract chunking_strategy
	var chunkingStrategy map[string]interface{}
	if v, ok := d.GetOk("chunking_strategy"); ok && len(v.([]interface{})) > 0 {
		chunk := v.([]interface{})[0].(map[string]interface{})
		chunkingStrategy = map[string]interface{}{
			"type": chunk["type"].(string),
		}

		if size, ok := chunk["size"]; ok && size != nil {
			chunkingStrategy["size"] = size.(int)
		}

		if maxTokens, ok := chunk["max_tokens"]; ok && maxTokens != nil {
			chunkingStrategy["max_tokens"] = maxTokens.(int)
		}
	}

	// Create request body
	requestBody := map[string]interface{}{
		"file_ids": fileIDs,
	}

	if len(attributes) > 0 {
		requestBody["attributes"] = attributes
	}

	if chunkingStrategy != nil {
		requestBody["chunking_strategy"] = chunkingStrategy
	}

	// Debug output
	debugJSON, _ := json.MarshalIndent(requestBody, "", "  ")
	log.Printf("[DEBUG] Vector store file batch request body: %s", string(debugJSON))

	// Make API call
	responseBytes, err := client.DoRequest("POST", fmt.Sprintf("/v1/vector_stores/%s/file_batches", vectorStoreID), requestBody)
	if err != nil {
		return diag.Errorf("Error adding file batch to vector store: %s", err.Error())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return diag.Errorf("Error parsing response: %s", err.Error())
	}

	// Set ID and other attributes
	if id, ok := response["id"]; ok && id != nil {
		d.SetId(id.(string))
	} else {
		return diag.Errorf("Response missing required 'id' field")
	}

	if createdAt, ok := response["created_at"]; ok && createdAt != nil {
		if err := d.Set("created_at", int(createdAt.(float64))); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
		}
	}

	if object, ok := response["object"]; ok && object != nil {
		if err := d.Set("object", object.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set object: %v", err))
		}
	}

	if status, ok := response["status"]; ok && status != nil {
		if err := d.Set("status", status.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set status: %v", err))
		}
	}

	return resourceOpenAIVectorStoreFileBatchRead(ctx, d, m)
}

func resourceOpenAIVectorStoreFileBatchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)

	// Make API call
	responseBytes, err := client.DoRequest("GET", fmt.Sprintf("/v1/vector_stores/%s/file_batches/%s", vectorStoreID, d.Id()), nil)
	if err != nil {
		return diag.Errorf("Error reading vector store file batch: %s", err.Error())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return diag.Errorf("Error parsing response: %s", err.Error())
	}

	// Set the computed attributes
	if createdAt, ok := response["created_at"]; ok && createdAt != nil {
		if err := d.Set("created_at", int(createdAt.(float64))); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
		}
	}

	if object, ok := response["object"]; ok && object != nil {
		if err := d.Set("object", object.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set object: %v", err))
		}
	}

	if status, ok := response["status"]; ok && status != nil {
		if err := d.Set("status", status.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set status: %v", err))
		}
	}

	// Handle file_ids
	if fileIDs, ok := response["file_ids"].([]interface{}); ok {
		if err := d.Set("file_ids", fileIDs); err != nil {
			return diag.Errorf("Error setting file_ids: %s", err)
		}
	}

	// Handle attributes
	if attributes, ok := response["attributes"].(map[string]interface{}); ok {
		if err := d.Set("attributes", attributes); err != nil {
			return diag.Errorf("Error setting attributes: %s", err)
		}
	}

	return diag.Diagnostics{}
}

func resourceOpenAIVectorStoreFileBatchUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)

	// Extract attributes
	var attributes map[string]interface{}
	if v, ok := d.GetOk("attributes"); ok {
		attributes = make(map[string]interface{})
		for k, attr := range v.(map[string]interface{}) {
			attributes[k] = attr
		}
	}

	// Only attributes can be updated
	requestBody := map[string]interface{}{}
	if len(attributes) > 0 {
		requestBody["attributes"] = attributes
	}

	// Make API call
	responseBytes, err := client.DoRequest("PUT", fmt.Sprintf("/v1/vector_stores/%s/file_batches/%s", vectorStoreID, d.Id()), requestBody)
	if err != nil {
		return diag.Errorf("Error updating vector store file batch: %s", err.Error())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return diag.Errorf("Error parsing response: %s", err.Error())
	}

	// Set the computed attributes
	if createdAt, ok := response["created_at"]; ok && createdAt != nil {
		if err := d.Set("created_at", int(createdAt.(float64))); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
		}
	}

	if object, ok := response["object"]; ok && object != nil {
		if err := d.Set("object", object.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set object: %v", err))
		}
	}

	if status, ok := response["status"]; ok && status != nil {
		if err := d.Set("status", status.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set status: %v", err))
		}
	}

	return resourceOpenAIVectorStoreFileBatchRead(ctx, d, m)
}

func resourceOpenAIVectorStoreFileBatchDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// OpenAI API doesn't support batch deletion, so this is a no-op.
	// Just clear the ID so Terraform knows the resource is gone.
	d.SetId("")
	return diag.Diagnostics{}
}
