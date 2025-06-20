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

func dataSourceOpenAIVectorStoreFiles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIVectorStoreFilesRead,
		Schema: map[string]*schema.Schema{
			"vector_store_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the vector store that the files belong to.",
			},
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
			"filter": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"in_progress", "completed", "failed", "cancelled"}, false),
				Description:  "Filter by file status. One of in_progress, completed, failed, cancelled.",
			},
			"files": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
							Description: "The Unix timestamp when the file was created.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the file.",
						},
					},
				},
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean indicating whether there are more files available beyond the current response.",
			},
		},
	}
}

func dataSourceOpenAIVectorStoreFilesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)

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

	if filter, ok := d.GetOk("filter"); ok {
		queryParams["filter"] = filter.(string)
	}

	// Construct the URL with query parameters
	url := fmt.Sprintf("/v1/vector_stores/%s/files", vectorStoreID)
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
		return diag.FromErr(fmt.Errorf("error listing vector store files: %s", err))
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

	// Process files
	var files []map[string]interface{}
	if data, ok := responseData["data"].([]interface{}); ok {
		for _, item := range data {
			if file, ok := item.(map[string]interface{}); ok {
				fileMap := map[string]interface{}{
					"id":     file["id"],
					"object": file["object"],
					"status": file["status"],
				}

				if createdAt, ok := file["created_at"].(float64); ok {
					fileMap["created_at"] = int(createdAt)
				}

				files = append(files, fileMap)
			}
		}
	}

	if err := d.Set("files", files); err != nil {
		return diag.FromErr(fmt.Errorf("error setting files: %s", err))
	}

	// Generate a unique ID based on the query parameters
	id := fmt.Sprintf("vector_store_files_%s_%d", vectorStoreID, schema.HashString(fmt.Sprintf("%v", queryParams)))
	d.SetId(id)

	return diag.Diagnostics{}
}
