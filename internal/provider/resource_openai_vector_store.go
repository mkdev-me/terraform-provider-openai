package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOpenAIVectorStore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIVectorStoreCreate,
		ReadContext:   resourceOpenAIVectorStoreRead,
		UpdateContext: resourceOpenAIVectorStoreUpdate,
		DeleteContext: resourceOpenAIVectorStoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the vector store.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the vector store.",
			},
			"file_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of file IDs to use in the vector store.",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of key-value pairs that can be attached to the vector store.",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the vector store was created.",
			},
			"file_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of files in the vector store.",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type (always 'vector_store').",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the vector store.",
			},
			"expires_after": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The expiration policy for the vector store.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"days": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Number of days after which the vector store should expire.",
						},
						"anchor": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The anchor time for the expiration (usually 'now').",
						},
					},
				},
			},
			"chunking_strategy": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The chunking strategy used to chunk the files.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of chunking strategy (auto or static).",
						},
					},
				},
			},
		},
	}
}

func resourceOpenAIVectorStoreCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)

	// Convert file_ids to []string
	var fileIDs []string
	if v, ok := d.GetOk("file_ids"); ok {
		for _, id := range v.([]interface{}) {
			fileIDs = append(fileIDs, id.(string))
		}
	}

	// Convert metadata to map[string]string
	var metadata map[string]string
	if v, ok := d.GetOk("metadata"); ok {
		metadata = make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			metadata[k] = v.(string)
		}
	}

	// Extract expires_after
	var expiresAfter map[string]interface{}
	if v, ok := d.GetOk("expires_after"); ok && len(v.([]interface{})) > 0 {
		exp := v.([]interface{})[0].(map[string]interface{})
		expiresAfter = make(map[string]interface{})

		if days, ok := exp["days"]; ok && days != nil {
			expiresAfter["days"] = days.(int)
		}

		if anchor, ok := exp["anchor"]; ok && anchor != nil {
			expiresAfter["anchor"] = anchor.(string)
		}
	}

	// Extract chunking_strategy
	var chunkingStrategy map[string]interface{}
	if v, ok := d.GetOk("chunking_strategy"); ok && len(v.([]interface{})) > 0 {
		chunk := v.([]interface{})[0].(map[string]interface{})
		chunkingStrategy = map[string]interface{}{
			"type": chunk["type"].(string),
		}
	}

	// Create request body
	requestBody := map[string]interface{}{
		"name": name,
	}

	if len(fileIDs) > 0 {
		requestBody["file_ids"] = fileIDs
	}

	if len(metadata) > 0 {
		requestBody["metadata"] = metadata
	}

	if expiresAfter != nil {
		requestBody["expires_after"] = expiresAfter
	}

	if chunkingStrategy != nil {
		requestBody["chunking_strategy"] = chunkingStrategy
	}

	// Make API call
	responseBytes, err := client.DoRequest("POST", "/v1/vector_stores", requestBody)
	if err != nil {
		return diag.Errorf("Error creating vector store: %s", err.Error())
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

	if fileCount, ok := response["file_count"]; ok && fileCount != nil {
		if err := d.Set("file_count", int(fileCount.(float64))); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set file_count: %v", err))
		}
	}

	if name, ok := response["name"]; ok && name != nil {
		if err := d.Set("name", name.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set name: %v", err))
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

	return resourceOpenAIVectorStoreRead(ctx, d, m)
}

func resourceOpenAIVectorStoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Make API call to fetch the vector store details
	responseBytes, err := client.DoRequest("GET", fmt.Sprintf("/v1/vector_stores/%s", d.Id()), nil)
	if err != nil {
		return diag.Errorf("Error reading vector store: %s", err.Error())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return diag.Errorf("Error parsing response: %s", err.Error())
	}

	// Set the parameters from the response
	if name, ok := response["name"]; ok && name != nil {
		_ = d.Set("name", name.(string))
	}
	if createdAt, ok := response["created_at"]; ok && createdAt != nil {
		_ = d.Set("created_at", int(createdAt.(float64)))
	}
	if fileCount, ok := response["file_count"]; ok && fileCount != nil {
		_ = d.Set("file_count", int(fileCount.(float64)))
	}
	if object, ok := response["object"]; ok && object != nil {
		_ = d.Set("object", object.(string))
	}
	if status, ok := response["status"]; ok && status != nil {
		_ = d.Set("status", status.(string))
	}

	// Handle optional fields with checks
	if fileIDs, ok := response["file_ids"].([]interface{}); ok {
		_ = d.Set("file_ids", fileIDs)
	}
	if metadata, ok := response["metadata"].(map[string]interface{}); ok {
		_ = d.Set("metadata", metadata)
	}

	// Optional fields: Set defaults or handle missing fields
	if expiresAfter, ok := response["expires_after"].(map[string]interface{}); ok {
		_ = d.Set("expires_after", expiresAfter)
	}
	if chunkingStrategy, ok := response["chunking_strategy"].(map[string]interface{}); ok {
		_ = d.Set("chunking_strategy", chunkingStrategy)
	}

	return diag.Diagnostics{}
}

func resourceOpenAIVectorStoreUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Prepare request body
	requestBody := make(map[string]interface{})

	// Add name if changed
	if d.HasChange("name") {
		requestBody["name"] = d.Get("name").(string)
	}

	// Add metadata if changed
	if d.HasChange("metadata") {
		metadata := make(map[string]string)
		if v, ok := d.GetOk("metadata"); ok {
			for k, v := range v.(map[string]interface{}) {
				metadata[k] = v.(string)
			}
			requestBody["metadata"] = metadata
		}
	}

	// Add file_ids if changed
	if d.HasChange("file_ids") {
		var fileIDs []string
		if v, ok := d.GetOk("file_ids"); ok {
			for _, id := range v.([]interface{}) {
				fileIDs = append(fileIDs, id.(string))
			}
			requestBody["file_ids"] = fileIDs
		}
	}

	// Only make an API call if there are changes
	if len(requestBody) > 0 {
		// Make API call
		responseBytes, err := client.DoRequest("POST", fmt.Sprintf("/v1/vector_stores/%s", d.Id()), requestBody)
		if err != nil {
			return diag.Errorf("Error updating vector store: %s", err.Error())
		}

		// Parse response
		var response map[string]interface{}
		if err := json.Unmarshal(responseBytes, &response); err != nil {
			return diag.Errorf("Error parsing response: %s", err.Error())
		}
	}

	return resourceOpenAIVectorStoreRead(ctx, d, m)
}

func resourceOpenAIVectorStoreDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Make delete request
	_, err = client.DoRequest("DELETE", fmt.Sprintf("/v1/vector_stores/%s", d.Id()), nil)
	if err != nil {
		return diag.Errorf("Error deleting vector store: %s", err.Error())
	}

	d.SetId("")
	return diag.Diagnostics{}
}
