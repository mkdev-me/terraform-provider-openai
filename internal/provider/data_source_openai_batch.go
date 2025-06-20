package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIBatch returns a schema.Resource that represents a data source for an OpenAI batch job.
// This data source allows users to retrieve information about batch processing jobs they've created.
func dataSourceOpenAIBatch() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIBatchRead,
		Schema: map[string]*schema.Schema{
			"batch_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the batch job to retrieve",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the project associated with the batch job. If not specified, the API key's default project will be used.",
			},
			"input_file_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the input file used for the batch",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint used for the batch request (e.g., '/v1/chat/completions')",
			},
			"completion_window": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time window specified for batch completion",
			},
			"output_file_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the output file (if available)",
			},
			"error_file_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the error file (if available)",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the batch job",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the batch job was created",
			},
			"in_progress_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the batch job began processing",
			},
			"expires_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the batch job expires",
			},
			"completed_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the batch job completed",
			},
			"request_counts": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Statistics about request processing",
			},
			"error": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Information about errors that occurred during processing",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Custom metadata attached to the batch job",
			},
		},
	}
}

// dataSourceOpenAIBatchRead handles the read operation for the OpenAI batch data source.
// It retrieves information about an existing batch job from the OpenAI API.
func dataSourceOpenAIBatchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, ok := m.(*OpenAIClient)
	if !ok {
		return diag.Errorf("error getting OpenAI client")
	}

	// Get the batch ID
	batchID := d.Get("batch_id").(string)
	if batchID == "" {
		return diag.Errorf("batch_id is required")
	}

	// Get project_id if present
	var projectID string
	if v, ok := d.GetOk("project_id"); ok {
		projectID = v.(string)
	}

	// Prepare the HTTP request
	url := fmt.Sprintf("%s/batches/%s", client.APIURL, batchID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}
	// Add project_id header if present
	if projectID != "" {
		req.Header.Set("OpenAI-Project", projectID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check if there was an error
	if resp.StatusCode != http.StatusOK {
		// Print the response body to help with debugging
		responsePreview := string(respBody)
		if len(responsePreview) > 1000 {
			responsePreview = responsePreview[:1000] + "... (truncated)"
		}
		fmt.Printf("Error response from API (Status %d): %s\n", resp.StatusCode, responsePreview)

		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			// Check if response looks like HTML (starts with '<')
			if len(respBody) > 0 && respBody[0] == '<' {
				return diag.FromErr(fmt.Errorf("received HTML response instead of JSON (likely auth or endpoint issue): Status %d, Response starts with: %s",
					resp.StatusCode, responsePreview[:100]))
			}
			return diag.FromErr(fmt.Errorf("error parsing error response (Status %d): %s\nResponse body: %s",
				resp.StatusCode, err, responsePreview))
		}
		return diag.FromErr(fmt.Errorf("error retrieving batch: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var batchResponse BatchResponse
	if err := json.Unmarshal(respBody, &batchResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the resource ID
	d.SetId(batchResponse.ID)

	// Set the attributes
	if err := d.Set("input_file_id", batchResponse.InputFileID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("endpoint", batchResponse.Endpoint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("completion_window", batchResponse.CompletionWindow); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("output_file_id", batchResponse.OutputFileID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("error_file_id", batchResponse.ErrorFileID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", batchResponse.Status); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", batchResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("expires_at", batchResponse.ExpiresAt); err != nil {
		return diag.FromErr(err)
	}

	// Set optional attributes that may be null
	if batchResponse.InProgressAt != nil {
		if err := d.Set("in_progress_at", *batchResponse.InProgressAt); err != nil {
			return diag.FromErr(err)
		}
	}
	if batchResponse.CompletedAt != nil {
		if err := d.Set("completed_at", *batchResponse.CompletedAt); err != nil {
			return diag.FromErr(err)
		}
	}

	// Convert request counts to map
	if batchResponse.RequestCounts != nil {
		if err := d.Set("request_counts", batchResponse.RequestCounts); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle errors if any
	if batchResponse.Errors != nil {
		errorStr, err := json.Marshal(batchResponse.Errors)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("error", string(errorStr)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Convert metadata to strings for Terraform
	if len(batchResponse.Metadata) > 0 {
		metadataMap := make(map[string]string)
		for k, v := range batchResponse.Metadata {
			switch val := v.(type) {
			case string:
				metadataMap[k] = val
			default:
				// Convert non-string values to JSON string
				jsonBytes, err := json.Marshal(val)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error marshaling metadata value: %s", err))
				}
				metadataMap[k] = string(jsonBytes)
			}
		}
		if err := d.Set("metadata", metadataMap); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}
