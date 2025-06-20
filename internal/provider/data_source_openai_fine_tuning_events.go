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

// FineTuningEventsResponse represents the API response for fine-tuning events
type FineTuningEventsResponse struct {
	Object  string                `json:"object"`   // Type of object returned (list)
	Data    []FineTuningEventData `json:"data"`     // List of events
	HasMore bool                  `json:"has_more"` // Whether there are more events to fetch
}

// FineTuningEventData represents a single fine-tuning event
type FineTuningEventData struct {
	Object    string `json:"object"`         // Type of object (event)
	ID        string `json:"id"`             // Unique identifier for this event
	CreatedAt int    `json:"created_at"`     // Unix timestamp when the event was created
	Level     string `json:"level"`          // Event level (info, warning, error)
	Message   string `json:"message"`        // The message describing the event
	Type      string `json:"type"`           // Event type (e.g., metrics, status_update)
	Data      any    `json:"data,omitempty"` // Additional data about the event
}

func dataSourceOpenAIFineTuningEvents() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIFineTuningEventsRead,
		Schema: map[string]*schema.Schema{
			"fine_tuning_job_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the fine-tuning job to get events for",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "Number of events to retrieve (default: 20)",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier for the last event from the previous pagination request",
			},
			"events": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique identifier for this event",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of object (event)",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Unix timestamp when the event was created",
						},
						"level": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Event level (info, warning, error)",
						},
						"message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The message describing the event",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Event type (e.g., metrics, status_update)",
						},
						"data_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Additional data about the event in JSON format",
						},
					},
				},
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more events to retrieve",
			},
		},
	}
}

func dataSourceOpenAIFineTuningEventsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	apiURL := fmt.Sprintf("%s/fine_tuning/jobs/%s/events", client.APIURL, jobID)
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
		return diag.FromErr(fmt.Errorf("error requesting fine-tuning events: %s", err))
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
			d.SetId(fmt.Sprintf("fine-tuning-events-%s", jobID))

			// Set has_more to false
			if err := d.Set("has_more", false); err != nil {
				return diag.FromErr(err)
			}

			// Set empty events list
			if err := d.Set("events", []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}

			// Return a warning instead of error
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Fine-tuning job not found",
					Detail:   fmt.Sprintf("Fine-tuning job with ID '%s' could not be found. This may be because it has been deleted or has expired. Returning empty events list.", jobID),
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
				d.SetId(fmt.Sprintf("fine-tuning-events-%s", jobID))

				// Set has_more to false
				if err := d.Set("has_more", false); err != nil {
					return diag.FromErr(err)
				}

				// Set empty events list
				if err := d.Set("events", []map[string]interface{}{}); err != nil {
					return diag.FromErr(err)
				}

				// Return a warning instead of error
				return diag.Diagnostics{
					diag.Diagnostic{
						Severity: diag.Warning,
						Summary:  "Fine-tuning job not found",
						Detail:   fmt.Sprintf("Fine-tuning job with ID '%s' could not be found: %s. Returning empty events list.", jobID, errorResponse.Error.Message),
					},
				}
			}
		}

		// For other error types, return the normal error
		return diag.FromErr(fmt.Errorf("error fetching fine-tuning events: %s - %s", resp.Status, string(body)))
	}

	// Parse the response
	var eventsResponse FineTuningEventsResponse
	if err := json.Unmarshal(body, &eventsResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the ID
	d.SetId(fmt.Sprintf("fine-tuning-events-%s", jobID))

	// Set has_more
	if err := d.Set("has_more", eventsResponse.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %s", err))
	}

	// Convert events to the format expected by the schema
	events := make([]map[string]interface{}, 0, len(eventsResponse.Data))
	for _, event := range eventsResponse.Data {
		eventMap := map[string]interface{}{
			"id":         event.ID,
			"object":     event.Object,
			"created_at": event.CreatedAt,
			"level":      event.Level,
			"message":    event.Message,
			"type":       event.Type,
		}

		// Convert event data to JSON if present
		if event.Data != nil {
			dataJSON, err := json.Marshal(event.Data)
			if err == nil {
				eventMap["data_json"] = string(dataJSON)
			}
		}

		events = append(events, eventMap)
	}

	if err := d.Set("events", events); err != nil {
		return diag.FromErr(fmt.Errorf("error setting events: %s", err))
	}

	return nil
}
