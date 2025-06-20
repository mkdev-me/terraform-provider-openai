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

// dataSourceOpenAIAssistant returns a schema.Resource that represents a data source for a single OpenAI assistant.
func dataSourceOpenAIAssistant() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIAssistantRead,
		Schema: map[string]*schema.Schema{
			"assistant_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the assistant to retrieve",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type, which is always 'assistant'",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the assistant was created",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the assistant",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the assistant",
			},
			"model": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The model used by the assistant",
			},
			"instructions": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The system instructions of the assistant",
			},
			"tools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of tool",
						},
						"function": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the function",
									},
									"description": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The description of the function",
									},
									"parameters": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The parameters of the function in JSON format",
									},
								},
							},
						},
					},
				},
				Description: "The tools enabled on the assistant",
			},
			"file_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of file IDs attached to the assistant",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Metadata attached to the assistant",
			},
			"response_format": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The format of responses from the assistant",
			},
			"reasoning_effort": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Constrains the effort spent on reasoning for reasoning models (low, medium, or high)",
			},
			"temperature": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "What sampling temperature to use for this assistant",
			},
			"top_p": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "An alternative to sampling with temperature, called nucleus sampling",
			},
		},
	}
}

// dataSourceOpenAIAssistantRead handles the read operation for the OpenAI assistant data source.
// It retrieves details about a specific assistant from the OpenAI API.
func dataSourceOpenAIAssistantRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %s", err))
	}

	// Get assistant ID from the schema
	assistantID := d.Get("assistant_id").(string)

	// Build URL for the request
	url := fmt.Sprintf("%s/assistants/%s", c.APIURL, assistantID)

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
		return diag.FromErr(fmt.Errorf("error retrieving assistant: %s - %s",
			errResp.Error.Type, errResp.Error.Message))
	}

	// Parse the response to a map first to handle all possible fields
	var assistantMap map[string]interface{}
	if err := json.Unmarshal(respBody, &assistantMap); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing assistant response to map: %s", err))
	}

	// Parse the response to the standard struct
	var assistant AssistantResponse
	if err := json.Unmarshal(respBody, &assistant); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing assistant response: %s", err))
	}

	// Set the ID in the resource data
	d.SetId(assistant.ID)

	// Set the basic assistant properties
	if err := d.Set("object", assistant.Object); err != nil {
		return diag.FromErr(fmt.Errorf("error setting object: %s", err))
	}
	if err := d.Set("created_at", assistant.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created_at: %s", err))
	}
	if err := d.Set("name", assistant.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %s", err))
	}
	if err := d.Set("description", assistant.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %s", err))
	}
	if err := d.Set("model", assistant.Model); err != nil {
		return diag.FromErr(fmt.Errorf("error setting model: %s", err))
	}
	if err := d.Set("instructions", assistant.Instructions); err != nil {
		return diag.FromErr(fmt.Errorf("error setting instructions: %s", err))
	}

	// Set tools
	toolsList := make([]interface{}, 0, len(assistant.Tools))
	for _, tool := range assistant.Tools {
		toolMap := map[string]interface{}{
			"type": tool.Type,
		}

		// Add function details if present
		if tool.Function != nil {
			functionMap := map[string]interface{}{
				"name":       tool.Function.Name,
				"parameters": string(tool.Function.Parameters),
			}
			if tool.Function.Description != "" {
				functionMap["description"] = tool.Function.Description
			}
			toolMap["function"] = []interface{}{functionMap}
		}

		toolsList = append(toolsList, toolMap)
	}
	if err := d.Set("tools", toolsList); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tools: %s", err))
	}

	// Set file_ids
	if err := d.Set("file_ids", assistant.FileIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting file_ids: %s", err))
	}

	// Set metadata
	metadataMap := make(map[string]string)
	for k, v := range assistant.Metadata {
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

	// Set additional properties from the map
	if responseFormat, ok := assistantMap["response_format"]; ok {
		if str, ok := responseFormat.(string); ok {
			if err := d.Set("response_format", str); err != nil {
				return diag.FromErr(fmt.Errorf("error setting response_format: %s", err))
			}
		} else if obj, ok := responseFormat.(map[string]interface{}); ok {
			// Convert map to JSON string for storing complex response_format
			jsonBytes, err := json.Marshal(obj)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error marshaling response_format: %s", err))
			}
			if err := d.Set("response_format", string(jsonBytes)); err != nil {
				return diag.FromErr(fmt.Errorf("error setting response_format as JSON: %s", err))
			}
		}
	}

	if reasoningEffort, ok := assistantMap["reasoning_effort"]; ok {
		if str, ok := reasoningEffort.(string); ok {
			if err := d.Set("reasoning_effort", str); err != nil {
				return diag.FromErr(fmt.Errorf("error setting reasoning_effort: %s", err))
			}
		}
	}

	if temperature, ok := assistantMap["temperature"]; ok {
		if num, ok := temperature.(float64); ok {
			if err := d.Set("temperature", num); err != nil {
				return diag.FromErr(fmt.Errorf("error setting temperature: %s", err))
			}
		}
	}

	if topP, ok := assistantMap["top_p"]; ok {
		if num, ok := topP.(float64); ok {
			if err := d.Set("top_p", num); err != nil {
				return diag.FromErr(fmt.Errorf("error setting top_p: %s", err))
			}
		}
	}

	return nil
}
