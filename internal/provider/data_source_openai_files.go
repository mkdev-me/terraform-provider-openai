package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ListFilesResponse represents the API response for listing OpenAI files
type ListFilesResponse struct {
	Data   []FileResponse `json:"data"`
	Object string         `json:"object"`
}

// dataSourceOpenAIFiles returns a schema.Resource that represents a data source for OpenAI files.
// This data source allows users to retrieve a list of all OpenAI files.
func dataSourceOpenAIFiles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIFilesRead,
		Schema: map[string]*schema.Schema{
			"purpose": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter files by purpose (e.g., 'fine-tune', 'assistants', etc.)",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The project ID to associate with this file lookup (for Terraform reference only, not sent to OpenAI API)",
			},
			"files": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the file",
						},
						"filename": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the file",
						},
						"bytes": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The size of the file in bytes",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the file was created",
						},
						"purpose": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The purpose of the file",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The object type, always 'file'",
						},
					},
				},
				Description: "List of OpenAI files",
			},
		},
	}
}

// dataSourceOpenAIFilesRead handles the read operation for the OpenAI files data source.
// It retrieves a list of all OpenAI files from the OpenAI API.
func dataSourceOpenAIFilesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get optional purpose filter
	purpose := d.Get("purpose").(string)

	// Set a unique ID for the resource
	d.SetId(fmt.Sprintf("files_%d", time.Now().Unix()))

	// Create URL to list files
	var url string
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/files", client.APIURL)
	} else {
		url = fmt.Sprintf("%s/v1/files", client.APIURL)
	}

	// Add purpose filter if provided
	if purpose != "" {
		url = fmt.Sprintf("%s?purpose=%s", url, purpose)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %w", err))
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %w", err))
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %w", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API error: %s - %s (%s)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code))
	}

	// Parse JSON response
	var filesResponse ListFilesResponse
	if err := json.Unmarshal(responseBody, &filesResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %w", err))
	}

	// Convert the response to the format expected by the schema
	files := make([]map[string]interface{}, len(filesResponse.Data))
	for i, file := range filesResponse.Data {
		fileMap := map[string]interface{}{
			"id":         file.ID,
			"filename":   file.Filename,
			"bytes":      file.Bytes,
			"purpose":    file.Purpose,
			"object":     file.Object,
			"created_at": time.Unix(int64(file.CreatedAt), 0).Format(time.RFC3339),
		}
		files[i] = fileMap
	}

	if err := d.Set("files", files); err != nil {
		return diag.FromErr(fmt.Errorf("error setting files: %s", err))
	}

	return diag.Diagnostics{}
}
