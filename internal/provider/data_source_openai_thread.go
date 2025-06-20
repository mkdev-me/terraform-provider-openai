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

// dataSourceOpenAIThread returns a schema.Resource that represents a data source for a single OpenAI thread.
func dataSourceOpenAIThread() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIThreadRead,
		Schema: map[string]*schema.Schema{
			"thread_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the thread to retrieve",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type, which is always 'thread'",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the thread was created",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Metadata attached to the thread",
			},
		},
	}
}

// dataSourceOpenAIThreadRead handles the read operation for the OpenAI thread data source.
// It retrieves details about a specific thread from the OpenAI API.
func dataSourceOpenAIThreadRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %s", err))
	}

	// Get thread ID from the schema
	threadID := d.Get("thread_id").(string)

	// Build URL for the request
	url := fmt.Sprintf("%s/threads/%s", c.APIURL, threadID)

	// Prepare HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	if c.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.OrganizationID)
	}

	// Make the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %s", err))
	}

	// Check for errors in the response
	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s, status code: %d, body: %s",
				err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error retrieving thread: %s - %s",
			errResp.Error.Type, errResp.Error.Message))
	}

	// Parse the response to a map first to handle all possible fields
	var threadMap map[string]interface{}
	if err := json.Unmarshal(respBody, &threadMap); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing thread response to map: %s", err))
	}

	// Parse the response to the standard struct
	var thread ThreadResponse
	if err := json.Unmarshal(respBody, &thread); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing thread response: %s", err))
	}

	// Set the ID in the resource data
	d.SetId(thread.ID)

	// Set the basic thread properties
	if err := d.Set("object", thread.Object); err != nil {
		return diag.FromErr(fmt.Errorf("error setting object: %s", err))
	}
	if err := d.Set("created_at", thread.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created_at: %s", err))
	}

	// Set metadata if present
	if thread.Metadata != nil && len(thread.Metadata) > 0 {
		metadataMap := make(map[string]string)
		for k, v := range thread.Metadata {
			if strVal, ok := v.(string); ok {
				metadataMap[k] = strVal
			} else {
				// Convert non-string values to JSON string
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error marshaling metadata value for key %s: %s", k, err))
				}
				metadataMap[k] = string(jsonBytes)
			}
		}
		if err := d.Set("metadata", metadataMap); err != nil {
			return diag.FromErr(fmt.Errorf("error setting metadata: %s", err))
		}
	}

	return nil
}
