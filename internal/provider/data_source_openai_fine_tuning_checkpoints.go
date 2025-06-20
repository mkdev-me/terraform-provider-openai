package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FineTuningCheckpointsResponse represents the API response for fine-tuning checkpoints
type FineTuningCheckpointsResponse struct {
	Object  string                     `json:"object"`   // Type of object returned (list)
	Data    []FineTuningCheckpointData `json:"data"`     // List of checkpoints
	HasMore bool                       `json:"has_more"` // Whether there are more checkpoints to fetch
}

// FineTuningCheckpointData represents a single fine-tuning checkpoint
type FineTuningCheckpointData struct {
	ID               string  `json:"id"`                 // Unique identifier for this checkpoint
	Object           string  `json:"object"`             // Type of object (fine_tuning.checkpoint)
	FineTuningJobID  string  `json:"fine_tuning_job_id"` // ID of the fine-tuning job that created this checkpoint
	CreatedAt        int     `json:"created_at"`         // Unix timestamp when the checkpoint was created
	Status           string  `json:"status"`             // Status of the checkpoint (e.g., active, deleted)
	TrainedTokens    int     `json:"trained_tokens"`     // Number of tokens processed during training until this checkpoint
	TrainingProgress float64 `json:"training_progress"`  // Progress percentage of the fine-tuning job when this checkpoint was created
}

func dataSourceOpenAIFineTuningCheckpoints() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIFineTuningCheckpointsRead,
		Schema: map[string]*schema.Schema{
			"fine_tuning_job_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the fine-tuning job to get checkpoints for",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "Number of checkpoints to retrieve (default: 20)",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier for the last checkpoint from the previous pagination request",
			},
			"checkpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique identifier for this checkpoint",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of object (fine_tuning.checkpoint)",
						},
						"fine_tuning_job_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the fine-tuning job that created this checkpoint",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Unix timestamp when the checkpoint was created",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the checkpoint (e.g., active, deleted)",
						},
						"trained_tokens": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of tokens processed during training until this checkpoint",
						},
						"training_progress": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Progress percentage of the fine-tuning job when this checkpoint was created",
						},
					},
				},
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more checkpoints to retrieve",
			},
		},
	}
}

func dataSourceOpenAIFineTuningCheckpointsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	jobID := d.Get("fine_tuning_job_id").(string)

	// Build the query parameters
	queryParams := url.Values{}

	if v, ok := d.GetOk("limit"); ok {
		queryParams.Set("limit", strconv.Itoa(v.(int)))
	}

	if v, ok := d.GetOk("after"); ok {
		queryParams.Set("after", v.(string))
	}

	// Build the URL
	apiURL := fmt.Sprintf("%s/fine_tuning/jobs/%s/checkpoints", client.APIURL, jobID)
	if len(queryParams) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, queryParams.Encode())
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error requesting fine-tuning checkpoints: %s", err))
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		// Handle 404 Not Found errors gracefully
		if resp.StatusCode == http.StatusNotFound {
			// Set the ID anyway so Terraform has something to reference
			d.SetId(fmt.Sprintf("fine-tuning-checkpoints-%s", jobID))

			// Set has_more to false
			if err := d.Set("has_more", false); err != nil {
				return diag.FromErr(err)
			}

			// Set empty checkpoints list
			if err := d.Set("checkpoints", []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}

			// Return a warning instead of error
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Fine-tuning job not found",
					Detail:   fmt.Sprintf("Fine-tuning job with ID '%s' could not be found. This may be because it has been deleted or has expired. Returning empty checkpoints list.", jobID),
				},
			}
		}

		// For other API errors that include an error message in the response
		var errorResponse struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			// If we have a "fine_tune_not_found" error, handle it gracefully
			if errorResponse.Error.Code == "fine_tune_not_found" {
				// Set the ID anyway so Terraform has something to reference
				d.SetId(fmt.Sprintf("fine-tuning-checkpoints-%s", jobID))

				// Set has_more to false
				if err := d.Set("has_more", false); err != nil {
					return diag.FromErr(err)
				}

				// Set empty checkpoints list
				if err := d.Set("checkpoints", []map[string]interface{}{}); err != nil {
					return diag.FromErr(err)
				}

				// Return a warning instead of error
				return diag.Diagnostics{
					diag.Diagnostic{
						Severity: diag.Warning,
						Summary:  "Fine-tuning job not found",
						Detail:   fmt.Sprintf("Fine-tuning job with ID '%s' could not be found: %s. Returning empty checkpoints list.", jobID, errorResponse.Error.Message),
					},
				}
			}
		}

		// For other error types, return the normal error
		return diag.FromErr(fmt.Errorf("error fetching fine-tuning checkpoints: %s - %s", resp.Status, string(body)))
	}

	// Parse the response
	var checkpointsResponse FineTuningCheckpointsResponse
	if err := json.Unmarshal(body, &checkpointsResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the ID
	d.SetId(fmt.Sprintf("fine-tuning-checkpoints-%s", jobID))

	// Set has_more
	if err := d.Set("has_more", checkpointsResponse.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %s", err))
	}

	// Convert checkpoints to the format expected by the schema
	checkpoints := make([]map[string]interface{}, 0, len(checkpointsResponse.Data))
	for _, checkpoint := range checkpointsResponse.Data {
		checkpointMap := map[string]interface{}{
			"id":                 checkpoint.ID,
			"object":             checkpoint.Object,
			"fine_tuning_job_id": checkpoint.FineTuningJobID,
			"created_at":         checkpoint.CreatedAt,
			"status":             checkpoint.Status,
			"trained_tokens":     checkpoint.TrainedTokens,
			"training_progress":  checkpoint.TrainingProgress,
		}

		checkpoints = append(checkpoints, checkpointMap)
	}

	if err := d.Set("checkpoints", checkpoints); err != nil {
		return diag.FromErr(fmt.Errorf("error setting checkpoints: %s", err))
	}

	return nil
}
