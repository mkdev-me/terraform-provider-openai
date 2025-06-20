package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOpenAIVectorStore() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIVectorStoreRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the vector store to retrieve.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the vector store.",
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
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of key-value pairs attached to the vector store.",
			},
			"file_ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of file IDs in the vector store.",
			},
			"expires_after": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"days": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of days after which the vector store should expire.",
						},
						"anchor": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The anchor time for the expiration.",
						},
					},
				},
				Description: "The expiration policy for the vector store.",
			},
			"chunking_strategy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of chunking strategy.",
						},
					},
				},
				Description: "The chunking strategy used for the files in the store.",
			},
		},
	}
}

func dataSourceOpenAIVectorStoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("id").(string)

	// Prepare request URL
	url := fmt.Sprintf("/v1/vector_stores/%s", vectorStoreID)

	// Make API request using DoRequest
	responseBytes, err := client.DoRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading vector store %s: %s", vectorStoreID, err))
	}

	// Parse response
	var vectorStore map[string]interface{}
	if err := json.Unmarshal(responseBytes, &vectorStore); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set ID
	d.SetId(vectorStoreID)

	// Set basic attributes
	if err := d.Set("name", vectorStore["name"]); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %s", err))
	}

	if err := d.Set("object", vectorStore["object"]); err != nil {
		return diag.FromErr(fmt.Errorf("error setting object: %s", err))
	}

	if err := d.Set("status", vectorStore["status"]); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %s", err))
	}

	// Set numeric attributes
	if createdAt, ok := vectorStore["created_at"].(float64); ok {
		if err := d.Set("created_at", int(createdAt)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting created_at: %s", err))
		}
	}

	if fileCount, ok := vectorStore["file_count"].(float64); ok {
		if err := d.Set("file_count", int(fileCount)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting file_count: %s", err))
		}
	}

	// Set metadata
	if metadata, ok := vectorStore["metadata"].(map[string]interface{}); ok {
		metadataMap := make(map[string]string)
		for key, value := range metadata {
			if strValue, ok := value.(string); ok {
				metadataMap[key] = strValue
			}
		}
		if err := d.Set("metadata", metadataMap); err != nil {
			return diag.FromErr(fmt.Errorf("error setting metadata: %s", err))
		}
	}

	// Set file_ids
	if fileIDs, ok := vectorStore["file_ids"].([]interface{}); ok {
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

	// Set expires_after
	if expiresAfter, ok := vectorStore["expires_after"].(map[string]interface{}); ok {
		expiresAfterList := []map[string]interface{}{}

		expiresAfterMap := map[string]interface{}{}

		if days, ok := expiresAfter["days"].(float64); ok {
			expiresAfterMap["days"] = int(days)
		}

		if anchor, ok := expiresAfter["anchor"].(string); ok {
			expiresAfterMap["anchor"] = anchor
		}

		if len(expiresAfterMap) > 0 {
			expiresAfterList = append(expiresAfterList, expiresAfterMap)

			if err := d.Set("expires_after", expiresAfterList); err != nil {
				return diag.FromErr(fmt.Errorf("error setting expires_after: %s", err))
			}
		}
	}

	// Set chunking_strategy
	if chunkingStrategy, ok := vectorStore["chunking_strategy"].(map[string]interface{}); ok {
		chunkingStrategyList := []map[string]interface{}{}

		chunkingStrategyMap := map[string]interface{}{}

		if strategyType, ok := chunkingStrategy["type"].(string); ok {
			chunkingStrategyMap["type"] = strategyType
		}

		if len(chunkingStrategyMap) > 0 {
			chunkingStrategyList = append(chunkingStrategyList, chunkingStrategyMap)

			if err := d.Set("chunking_strategy", chunkingStrategyList); err != nil {
				return diag.FromErr(fmt.Errorf("error setting chunking_strategy: %s", err))
			}
		}
	}

	return diag.Diagnostics{}
}
