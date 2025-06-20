package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIFile provides a data source to retrieve OpenAI file details.
// This allows users to reference existing files in their Terraform configurations.
func dataSourceOpenAIFile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIFileRead,
		Schema: map[string]*schema.Schema{
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the file to retrieve",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The project ID to associate with this file lookup (for Terraform reference only, not sent to OpenAI API)",
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
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the file was created",
			},
			"purpose": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The purpose of the file",
			},
		},
	}
}

// dataSourceOpenAIFileRead reads information about an existing OpenAI file.
// It fetches the file details from the OpenAI API and populates the Terraform state.
func dataSourceOpenAIFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the file ID
	fileID := d.Get("file_id").(string)
	if fileID == "" {
		return diag.FromErr(fmt.Errorf("file_id is required"))
	}

	// Create URL to get file information
	var url string
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/files/%s", client.APIURL, fileID)
	} else {
		url = fmt.Sprintf("%s/v1/files/%s", client.APIURL, fileID)
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

	// If file doesn't exist (404), return error
	if resp.StatusCode == http.StatusNotFound {
		return diag.FromErr(fmt.Errorf("file with ID %s not found", fileID))
	}

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
	var fileResponse FileResponse
	if err := json.Unmarshal(responseBody, &fileResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %w", err))
	}

	// Set the resource ID and data
	d.SetId(fileResponse.ID)
	if err := d.Set("filename", fileResponse.Filename); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("bytes", fileResponse.Bytes); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", fileResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("purpose", fileResponse.Purpose); err != nil {
		return diag.FromErr(err)
	}

	// Keep project_id if it was set
	// The API doesn't return project_id, so we have to keep it from the state

	return diag.Diagnostics{}
}
