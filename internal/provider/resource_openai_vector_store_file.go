package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOpenAIVectorStoreFile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIVectorStoreFileCreate,
		ReadContext:   resourceOpenAIVectorStoreFileRead,
		UpdateContext: resourceOpenAIVectorStoreFileUpdate,
		DeleteContext: resourceOpenAIVectorStoreFileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: importVectorStoreFileState,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the vector store file.",
			},
			"vector_store_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the vector store.",
			},
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the file to add to the vector store.",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Dynamic description or tags for the file in the vector store.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"chunking_strategy": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The chunking strategy used to chunk the file.",
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
				Description: "The timestamp for when the file was added to the vector store.",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type (always 'vector_store.file').",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the file in the vector store.",
			},
		},
	}
}

// importVectorStoreFileState handles the import of a vector store file resource
// The ID is expected to be in the format "{vector_store_id}/{file_id}"
func importVectorStoreFileState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return nil, fmt.Errorf("error getting OpenAI client: %s", err)
	}

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid ID format. Expected 'vector_store_id/file_id', got: %s", d.Id())
	}

	vectorStoreID, fileID := parts[0], parts[1]

	// Set the component IDs
	if err := d.Set("vector_store_id", vectorStoreID); err != nil {
		return nil, fmt.Errorf("error setting vector_store_id: %s", err)
	}

	if err := d.Set("file_id", fileID); err != nil {
		return nil, fmt.Errorf("error setting file_id: %s", err)
	}

	// Make API call to get the current state of the resource
	responseBytes, err := client.DoRequest("GET", fmt.Sprintf("/v1/vector_stores/%s/files/%s", vectorStoreID, fileID), nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vector store file: %s", err)
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %s", err)
	}

	// Set the ID first
	id, ok := response["id"]
	if !ok || id == nil {
		return nil, fmt.Errorf("response missing required 'id' field")
	}
	d.SetId(id.(string))

	// Set other fields from API response
	if createdAt, ok := response["created_at"]; ok && createdAt != nil {
		if err := d.Set("created_at", int(createdAt.(float64))); err != nil {
			return nil, fmt.Errorf("error setting created_at: %s", err)
		}
	}

	if object, ok := response["object"]; ok && object != nil {
		if err := d.Set("object", object.(string)); err != nil {
			return nil, fmt.Errorf("error setting object: %s", err)
		}
	}

	if status, ok := response["status"]; ok && status != nil {
		if err := d.Set("status", status.(string)); err != nil {
			return nil, fmt.Errorf("error setting status: %s", err)
		}
	}

	// Handle attributes
	if attributes, ok := response["attributes"].(map[string]interface{}); ok && len(attributes) > 0 {
		attrMap := make(map[string]interface{})
		for k, v := range attributes {
			attrMap[k] = fmt.Sprintf("%v", v)
		}
		if err := d.Set("attributes", attrMap); err != nil {
			return nil, fmt.Errorf("error setting attributes: %s", err)
		}
	}

	// Handle chunking_strategy if present
	if cs, ok := response["chunking_strategy"].(map[string]interface{}); ok && len(cs) > 0 {
		chunkingStrategy := []interface{}{cs}
		if err := d.Set("chunking_strategy", chunkingStrategy); err != nil {
			return nil, fmt.Errorf("error setting chunking_strategy: %s", err)
		}
	}

	return []*schema.ResourceData{d}, nil
}

func resourceOpenAIVectorStoreFileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)
	fileID := d.Get("file_id").(string)

	// Extract attributes
	var attributes map[string]string
	if v, ok := d.GetOk("attributes"); ok {
		attributes = make(map[string]string)
		for k, attr := range v.(map[string]interface{}) {
			attributes[k] = attr.(string)
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
		"file_id": fileID,
	}

	if len(attributes) > 0 {
		requestBody["attributes"] = attributes
	}

	if chunkingStrategy != nil {
		requestBody["chunking_strategy"] = chunkingStrategy
	}

	// Make API call
	responseBytes, err := client.DoRequest("POST", fmt.Sprintf("/v1/vector_stores/%s/files", vectorStoreID), requestBody)
	if err != nil {
		return diag.Errorf("Error adding file to vector store: %s", err.Error())
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
			return diag.FromErr(err)
		}
	}

	if object, ok := response["object"]; ok && object != nil {
		if err := d.Set("object", object.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if status, ok := response["status"]; ok && status != nil {
		if err := d.Set("status", status.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceOpenAIVectorStoreFileRead(ctx, d, m)
}

func resourceOpenAIVectorStoreFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)

	// Make API call
	responseBytes, err := client.DoRequest("GET", fmt.Sprintf("/v1/vector_stores/%s/files/%s", vectorStoreID, d.Id()), nil)
	if err != nil {
		return diag.Errorf("Error reading vector store file: %s", err.Error())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return diag.Errorf("Error parsing response: %s", err.Error())
	}

	// Set attributes
	if fileID, ok := response["file_id"]; ok && fileID != nil {
		if err := d.Set("file_id", fileID.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if createdAt, ok := response["created_at"]; ok && createdAt != nil {
		if err := d.Set("created_at", int(createdAt.(float64))); err != nil {
			return diag.FromErr(err)
		}
	}

	if object, ok := response["object"]; ok && object != nil {
		if err := d.Set("object", object.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if status, ok := response["status"]; ok && status != nil {
		if err := d.Set("status", status.(string)); err != nil {
			return diag.FromErr(err)
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

func resourceOpenAIVectorStoreFileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)

	// Extract attributes
	var attributes map[string]string
	if v, ok := d.GetOk("attributes"); ok {
		attributes = make(map[string]string)
		for k, attr := range v.(map[string]interface{}) {
			attributes[k] = attr.(string)
		}
	}

	// Create request body
	requestBody := map[string]interface{}{
		"attributes": attributes,
	}

	// Make API call
	responseBytes, err := client.DoRequest("POST", fmt.Sprintf("/v1/vector_stores/%s/files/%s", vectorStoreID, d.Id()), requestBody)
	if err != nil {
		return diag.Errorf("Error updating vector store file: %s", err.Error())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return diag.Errorf("Error parsing response: %s", err.Error())
	}

	return resourceOpenAIVectorStoreFileRead(ctx, d, m)
}

func resourceOpenAIVectorStoreFileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)

	// Make delete request
	_, err = client.DoRequest("DELETE", fmt.Sprintf("/v1/vector_stores/%s/files/%s", vectorStoreID, d.Id()), nil)
	if err != nil {
		return diag.Errorf("Error deleting vector store file: %s", err.Error())
	}

	d.SetId("")
	return diag.Diagnostics{}
}
