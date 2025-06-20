package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceOpenAIVectorStores() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIVectorStoresRead,
		Schema: map[string]*schema.Schema{
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 100),
				Description:  "A limit on the number of objects to be returned. Limit can range between 1 and 100, and the default is 20.",
			},
			"order": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "desc",
				ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
				Description:  "Sort order by the created_at timestamp of the objects. asc for ascending order and desc for descending order.",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A cursor for use in pagination. after is an object ID that defines your place in the list.",
			},
			"before": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A cursor for use in pagination. before is an object ID that defines your place in the list.",
			},
			"vector_stores": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the vector store.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the vector store.",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The object type, always 'vector_store'.",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The Unix timestamp of when the vector store was created.",
						},
						"file_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of files in the vector store.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the vector store.",
						},
					},
				},
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more vector stores available.",
			},
		},
	}
}

func dataSourceOpenAIVectorStoresRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build query parameters
	queryParams := map[string]string{}

	if limit, ok := d.GetOk("limit"); ok {
		queryParams["limit"] = strconv.Itoa(limit.(int))
	}

	if order, ok := d.GetOk("order"); ok {
		queryParams["order"] = order.(string)
	}

	if after, ok := d.GetOk("after"); ok {
		queryParams["after"] = after.(string)
	}

	if before, ok := d.GetOk("before"); ok {
		queryParams["before"] = before.(string)
	}

	// Construct the URL with query parameters
	url := "/v1/vector_stores"
	first := true
	for key, value := range queryParams {
		if first {
			url += "?"
			first = false
		} else {
			url += "&"
		}
		url += key + "=" + value
	}

	// Make the request
	responseBytes, err := client.DoRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing vector stores: %s", err))
	}

	// Parse the response
	var responseData map[string]interface{}
	if err := json.Unmarshal(responseBytes, &responseData); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set has_more
	if hasMore, ok := responseData["has_more"].(bool); ok {
		if err := d.Set("has_more", hasMore); err != nil {
			return diag.FromErr(fmt.Errorf("error setting has_more: %s", err))
		}
	}

	// Process vector stores
	var vectorStores []map[string]interface{}
	if data, ok := responseData["data"].([]interface{}); ok {
		for _, item := range data {
			if store, ok := item.(map[string]interface{}); ok {
				storeMap := map[string]interface{}{
					"id":     store["id"],
					"name":   store["name"],
					"object": store["object"],
					"status": store["status"],
				}

				if createdAt, ok := store["created_at"].(float64); ok {
					storeMap["created_at"] = int(createdAt)
				}

				if fileCount, ok := store["file_count"].(float64); ok {
					storeMap["file_count"] = int(fileCount)
				}

				vectorStores = append(vectorStores, storeMap)
			}
		}
	}

	if err := d.Set("vector_stores", vectorStores); err != nil {
		return diag.FromErr(fmt.Errorf("error setting vector_stores: %s", err))
	}

	// Generate a unique ID based on the query parameters
	id := fmt.Sprintf("vector_stores_%d", schema.HashString(fmt.Sprintf("%v", queryParams)))
	d.SetId(id)

	return diag.Diagnostics{}
}
