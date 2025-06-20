package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOpenAIVectorStoreFileContent() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIVectorStoreFileContentRead,
		Schema: map[string]*schema.Schema{
			"vector_store_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the vector store.",
			},
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the file within the vector store.",
			},
			"content": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The content of the file.",
			},
		},
	}
}

func dataSourceOpenAIVectorStoreFileContentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	vectorStoreID := d.Get("vector_store_id").(string)
	fileID := d.Get("file_id").(string)

	// Prepare request URL
	url := fmt.Sprintf("/v1/vector_stores/%s/files/%s/content", vectorStoreID, fileID)

	// Make API request
	responseBytes, err := client.DoRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving vector store file content %s: %s", fileID, err))
	}

	// Set ID as a composite of vector_store_id and file_id
	d.SetId(fmt.Sprintf("%s_%s_content", vectorStoreID, fileID))

	// Set content
	if err := d.Set("content", string(responseBytes)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting content: %s", err))
	}

	return diag.Diagnostics{}
}
