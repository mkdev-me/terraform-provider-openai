package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ListRunsResponse represents the API response for listing runs in a thread.
// It provides details about the list of runs and pagination metadata.
type ListRunsResponse struct {
	Object  string        `json:"object"`   // Object type, always "list"
	Data    []RunResponse `json:"data"`     // Array of runs
	FirstID string        `json:"first_id"` // ID of the first item in the list
	LastID  string        `json:"last_id"`  // ID of the last item in the list
	HasMore bool          `json:"has_more"` // Whether there are more items to fetch
}

// dataSourceOpenAIRuns provides a data source to list runs in a thread.
// This allows users to retrieve and manage multiple runs in their Terraform configurations.
func dataSourceOpenAIRuns() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIRunsRead,
		Schema: map[string]*schema.Schema{
			"thread_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the thread to list runs for",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 100),
				Description:  "A limit on the number of runs to be returned. Limit can range between 1 and 100, default is 20",
			},
			"order": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "desc",
				ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
				Description:  "Sort order by the created_at timestamp of the runs (asc for ascending, desc for descending)",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A cursor for pagination. This is a run ID that defines your place in the list",
			},
			"before": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A cursor for pagination. This is a run ID that defines your place in the list",
			},
			"runs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the run",
						},
						"thread_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the thread this run belongs to",
						},
						"assistant_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the assistant used for the run",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the run (queued, in_progress, completed, failed, etc.)",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the model used for the run",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp for when the run was created",
						},
						"started_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp for when the run was started",
						},
						"completed_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp for when the run was completed",
						},
					},
				},
				Description: "List of runs for the thread",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more runs to fetch",
			},
			"first_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the first run in the list",
			},
			"last_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the last run in the list",
			},
		},
	}
}

// dataSourceOpenAIRunsRead fetches a list of runs for a thread from the OpenAI API.
// It processes pagination parameters and extracts run information.
func dataSourceOpenAIRunsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*OpenAIClient)

	// Get the thread ID
	threadID := d.Get("thread_id").(string)
	if threadID == "" {
		return diag.FromErr(fmt.Errorf("thread_id is required"))
	}

	// Extract the query parameters
	limit := d.Get("limit").(int)
	order := d.Get("order").(string)
	after := d.Get("after").(string)
	before := d.Get("before").(string)

	// Construct the API URL with query parameters
	baseURL := fmt.Sprintf("%s/threads/%s/runs", client.APIURL, threadID)
	url := fmt.Sprintf("%s?limit=%d&order=%s", baseURL, limit, order)
	if after != "" {
		url = fmt.Sprintf("%s&after=%s", url, after)
	}
	if before != "" {
		url = fmt.Sprintf("%s&before=%s", url, before)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %w", err))
	}

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	// Send the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %w", err))
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %w", err))
	}

	// Check for error responses
	if resp.StatusCode != http.StatusOK {
		return diag.FromErr(fmt.Errorf("API returned error - status code: %d, body: %s", resp.StatusCode, string(respBody)))
	}

	// Parse the response
	var listResponse ListRunsResponse
	if err := json.Unmarshal(respBody, &listResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response body: %w", err))
	}

	// Extract run information for Terraform state
	runs := make([]map[string]interface{}, 0, len(listResponse.Data))
	for _, run := range listResponse.Data {
		runData := map[string]interface{}{
			"id":           run.ID,
			"thread_id":    run.ThreadID,
			"assistant_id": run.AssistantID,
			"status":       run.Status,
			"model":        run.Model,
			"created_at":   run.CreatedAt,
		}

		// Add optional fields
		if run.StartedAt != nil {
			runData["started_at"] = *run.StartedAt
		}
		if run.CompletedAt != nil {
			runData["completed_at"] = *run.CompletedAt
		}

		runs = append(runs, runData)
	}

	// Set the data source ID using the thread ID and parameters
	// This ensures proper refresh behavior when the params change
	idParts := []string{
		threadID,
		strconv.Itoa(limit),
		order,
	}
	if after != "" {
		idParts = append(idParts, "after="+after)
	}
	if before != "" {
		idParts = append(idParts, "before="+before)
	}
	d.SetId(strings.Join(idParts, ":"))

	// Set the retrieved data in the state
	if err := d.Set("runs", runs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting runs: %w", err))
	}
	if err := d.Set("has_more", listResponse.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %w", err))
	}
	if err := d.Set("first_id", listResponse.FirstID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting first_id: %w", err))
	}
	if err := d.Set("last_id", listResponse.LastID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting last_id: %w", err))
	}

	return nil
}
