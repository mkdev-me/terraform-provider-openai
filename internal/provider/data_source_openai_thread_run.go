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

// dataSourceOpenAIThreadRun provides a data source to retrieve details about a specific OpenAI thread run.
func dataSourceOpenAIThreadRun() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIThreadRunRead,
		Schema: map[string]*schema.Schema{
			"run_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the run to retrieve",
			},
			"thread_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the thread the run belongs to",
			},
			"assistant_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the assistant used for the run",
			},
			"model": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the model used for the run",
			},
			"instructions": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instructions used for the run",
			},
			"tools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				Description: "The tools available to the assistant for the run",
			},
			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Metadata associated with the run",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the run (queued, in_progress, completed, failed, etc.)",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type, which is always 'thread.run'",
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
			"file_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The IDs of the files used in the run",
			},
			"usage": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "Usage statistics for the run",
			},
		},
	}
}

// dataSourceOpenAIThreadRunRead fetches information about an existing OpenAI thread run.
func dataSourceOpenAIThreadRunRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*OpenAIClient)

	// Get the run ID and thread ID from the data source configuration
	runID := d.Get("run_id").(string)
	threadID := d.Get("thread_id").(string)

	// Construct the API URL
	url := fmt.Sprintf("%s/threads/%s/runs/%s", client.APIURL, threadID, runID)

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
	var threadRunResponse ThreadRunResponse
	if err := json.Unmarshal(respBody, &threadRunResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response body: %w", err))
	}

	// Set the resource ID
	d.SetId(threadRunResponse.ID)

	// Set the run data in the state
	if err := d.Set("assistant_id", threadRunResponse.AssistantID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting assistant_id: %w", err))
	}
	if err := d.Set("model", threadRunResponse.Model); err != nil {
		return diag.FromErr(fmt.Errorf("error setting model: %w", err))
	}
	if err := d.Set("instructions", threadRunResponse.Instructions); err != nil {
		return diag.FromErr(fmt.Errorf("error setting instructions: %w", err))
	}
	if err := d.Set("status", threadRunResponse.Status); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %w", err))
	}
	if err := d.Set("object", threadRunResponse.Object); err != nil {
		return diag.FromErr(fmt.Errorf("error setting object: %w", err))
	}
	if err := d.Set("created_at", threadRunResponse.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created_at: %w", err))
	}
	if err := d.Set("file_ids", threadRunResponse.FileIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting file_ids: %w", err))
	}

	// Set optional fields
	if threadRunResponse.StartedAt != nil {
		if err := d.Set("started_at", *threadRunResponse.StartedAt); err != nil {
			return diag.FromErr(fmt.Errorf("error setting started_at: %w", err))
		}
	}
	if threadRunResponse.CompletedAt != nil {
		if err := d.Set("completed_at", *threadRunResponse.CompletedAt); err != nil {
			return diag.FromErr(fmt.Errorf("error setting completed_at: %w", err))
		}
	}

	// Set usage data if available
	if threadRunResponse.Usage != nil {
		usageData := map[string]interface{}{
			"prompt_tokens":     threadRunResponse.Usage.PromptTokens,
			"completion_tokens": threadRunResponse.Usage.CompletionTokens,
			"total_tokens":      threadRunResponse.Usage.TotalTokens,
		}
		if err := d.Set("usage", usageData); err != nil {
			return diag.FromErr(fmt.Errorf("error setting usage: %w", err))
		}
	}

	// Set tools data
	if len(threadRunResponse.Tools) > 0 {
		if err := d.Set("tools", threadRunResponse.Tools); err != nil {
			return diag.FromErr(fmt.Errorf("error setting tools: %w", err))
		}
	}

	// Set metadata if present
	if threadRunResponse.Metadata != nil {
		metadata := make(map[string]string)
		for k, v := range threadRunResponse.Metadata {
			if strVal, ok := v.(string); ok {
				metadata[k] = strVal
			} else {
				// Convert non-string values to JSON strings
				jsonData, err := json.Marshal(v)
				if err == nil {
					metadata[k] = string(jsonData)
				}
			}
		}
		if err := d.Set("metadata", metadata); err != nil {
			return diag.FromErr(fmt.Errorf("error setting metadata: %w", err))
		}
	}

	return nil
}
