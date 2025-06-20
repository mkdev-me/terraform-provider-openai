package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOpenAIVectorStoreFile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIVectorStoreFileRead,
		Schema: map[string]*schema.Schema{
			"vector_store_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the vector store.",
			},
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the file to retrieve details for.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the file.",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type (always 'vector_store.file').",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the file was created.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the file.",
			},
			"attributes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The size of the file in bytes.",
						},
						"filename": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the file.",
						},
						"purpose": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The purpose of the file.",
						},
					},
				},
				Description: "Attributes of the file.",
			},
		},
	}
}

func dataSourceOpenAIVectorStoreFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)
	fileID := d.Get("file_id").(string)

	// Prepare request URL
	url := fmt.Sprintf("/v1/vector_stores/%s/files/%s", vectorStoreID, fileID)

	// Make API request
	responseBytes, err := client.DoRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading vector store file %s: %s", fileID, err))
	}

	// Parse response
	var file map[string]interface{}
	if err := json.Unmarshal(responseBytes, &file); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set ID
	if id, ok := file["id"].(string); ok {
		d.SetId(id)
	} else {
		return diag.FromErr(fmt.Errorf("file ID not found in response"))
	}

	// Set basic attributes
	if err := d.Set("object", file["object"]); err != nil {
		return diag.FromErr(fmt.Errorf("error setting object: %s", err))
	}

	if err := d.Set("status", file["status"]); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %s", err))
	}

	// Set numeric attributes
	if createdAt, ok := file["created_at"].(float64); ok {
		if err := d.Set("created_at", int(createdAt)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting created_at: %s", err))
		}
	}

	// Set attributes
	if attributes, ok := file["attributes"].(map[string]interface{}); ok {
		attributesList := []map[string]interface{}{}
		attributesMap := map[string]interface{}{}

		// Set size
		if size, ok := attributes["size"].(float64); ok {
			attributesMap["size"] = int(size)
		}

		// Set filename
		if filename, ok := attributes["filename"].(string); ok {
			attributesMap["filename"] = filename
		}

		// Set purpose
		if purpose, ok := attributes["purpose"].(string); ok {
			attributesMap["purpose"] = purpose
		}

		if len(attributesMap) > 0 {
			attributesList = append(attributesList, attributesMap)
			if err := d.Set("attributes", attributesList); err != nil {
				return diag.FromErr(fmt.Errorf("error setting attributes: %s", err))
			}
		}
	}

	return diag.Diagnostics{}
}
