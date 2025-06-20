package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// BatchesResponse represents the API response for listing OpenAI batch jobs
type BatchesResponse struct {
	Object  string          `json:"object"`
	Data    []BatchResponse `json:"data"`
	HasMore bool            `json:"has_more"`
}

// dataSourceOpenAIBatches defines the schema and read operation for the OpenAI batches data source.
// This data source allows retrieving information about all batch jobs for a specific OpenAI project.
func dataSourceOpenAIBatches() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIBatchesRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the project associated with the batch jobs. If not specified, the API key's default project will be used.",
			},
			"batches": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the batch job",
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
						"metadata": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Custom metadata attached to the batch job",
						},
					},
				},
			},
		},
	}
}

// dataSourceOpenAIBatchesRead fetches information about all batch jobs for a specific project from OpenAI.
func dataSourceOpenAIBatchesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, ok := m.(*OpenAIClient)
	if !ok {
		return diag.Errorf("error getting OpenAI client")
	}

	// Get project_id if present
	var projectID string
	if v, ok := d.GetOk("project_id"); ok {
		projectID = v.(string)
	}

	// Use the provider's API key
	apiKey := client.APIKey

	// Prepare the HTTP request
	url := fmt.Sprintf("%s/batches", client.APIURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
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

		// Handle permission errors
		if strings.Contains(responsePreview, "insufficient permissions") ||
			strings.Contains(responsePreview, "Missing scopes") {
			tflog.Info(ctx, fmt.Sprintf("Permission error reading batches: %s", responsePreview))
			return diag.FromErr(fmt.Errorf("error retrieving batches: API error: You have insufficient permissions for this operation"))
		}

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
		return diag.FromErr(fmt.Errorf("error retrieving batches: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var batchesResponse BatchesResponse
	if err := json.Unmarshal(respBody, &batchesResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Generate a consistent ID
	d.SetId(fmt.Sprintf("batches-%s", projectID))

	// Transform the response into the expected format
	batches := make([]map[string]interface{}, 0, len(batchesResponse.Data))
	for _, batch := range batchesResponse.Data {
		batchMap := map[string]interface{}{
			"id":                batch.ID,
			"input_file_id":     batch.InputFileID,
			"endpoint":          batch.Endpoint,
			"completion_window": batch.CompletionWindow,
			"output_file_id":    batch.OutputFileID,
			"error_file_id":     batch.ErrorFileID,
			"status":            batch.Status,
			"created_at":        batch.CreatedAt,
			"expires_at":        batch.ExpiresAt,
		}

		// Handle optional fields
		if batch.InProgressAt != nil {
			batchMap["in_progress_at"] = *batch.InProgressAt
		}

		if batch.CompletedAt != nil {
			batchMap["completed_at"] = *batch.CompletedAt
		}

		// Handle request counts
		if batch.RequestCounts != nil {
			batchMap["request_counts"] = batch.RequestCounts
		}

		// Handle metadata
		if len(batch.Metadata) > 0 {
			metadataMap := make(map[string]string)
			for k, v := range batch.Metadata {
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
			batchMap["metadata"] = metadataMap
		}

		batches = append(batches, batchMap)
	}

	// Set the result
	if err := d.Set("batches", batches); err != nil {
		return diag.FromErr(fmt.Errorf("error setting batches: %w", err))
	}

	return diag.Diagnostics{}
}
