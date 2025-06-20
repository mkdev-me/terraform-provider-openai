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

// dataSourceOpenAIModelResponseInputItems provides a data source to retrieve input items for a model response.
func dataSourceOpenAIModelResponseInputItems() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIModelResponseInputItemsRead,
		Schema: map[string]*schema.Schema{
			"response_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the model response to retrieve input items for",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An item ID to list items after, used in pagination",
			},
			"before": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An item ID to list items before, used in pagination",
			},
			"include": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Additional fields to include in the response",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "A limit on the number of objects to be returned (1-100, default: 20)",
			},
			"order": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "asc",
				Description: "The order to return items in (asc or desc, default: asc)",
			},
			"input_items": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the input item",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the input item",
						},
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role of the input item",
						},
						"content": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The content of the input item",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the input item",
						},
					},
				},
				Description: "The input items for the model response",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more items to fetch",
			},
			"first_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the first item in the list",
			},
			"last_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the last item in the list",
			},
		},
		Description: "Data source for retrieving input items for an OpenAI model response",
	}
}

// dataSourceOpenAIModelResponseInputItemsRead reads input items for an existing OpenAI model response.
func dataSourceOpenAIModelResponseInputItemsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the response ID
	responseID := d.Get("response_id").(string)
	if responseID == "" {
		return diag.FromErr(fmt.Errorf("response_id is required"))
	}

	// Build the base URL
	var baseURL string
	if strings.HasSuffix(client.APIURL, "/v1") {
		baseURL = fmt.Sprintf("%s/responses/%s/input_items", client.APIURL, responseID)
	} else {
		baseURL = fmt.Sprintf("%s/v1/responses/%s/input_items", client.APIURL, responseID)
	}

	// Add query parameters
	params := make([]string, 0)

	if v, ok := d.GetOk("after"); ok {
		params = append(params, fmt.Sprintf("after=%s", v.(string)))
	}

	if v, ok := d.GetOk("before"); ok {
		params = append(params, fmt.Sprintf("before=%s", v.(string)))
	}

	if v, ok := d.GetOk("limit"); ok {
		params = append(params, fmt.Sprintf("limit=%d", v.(int)))
	}

	if v, ok := d.GetOk("order"); ok {
		params = append(params, fmt.Sprintf("order=%s", v.(string)))
	}

	// Add optional include parameter if specified
	if v, ok := d.GetOk("include"); ok {
		includeList := v.([]interface{})
		if len(includeList) > 0 {
			includeParams := make([]string, len(includeList))
			for i, item := range includeList {
				includeParams[i] = item.(string)
			}
			params = append(params, fmt.Sprintf("include=%s", strings.Join(includeParams, ",")))
		}
	}

	// Construct the full URL with query parameters
	url := baseURL
	if len(params) > 0 {
		url = fmt.Sprintf("%s?%s", baseURL, strings.Join(params, "&"))
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %w", err))
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
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

	// If response doesn't exist (404), return error
	if resp.StatusCode == http.StatusNotFound {
		return diag.FromErr(fmt.Errorf("model response with ID %s not found", responseID))
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

	// Parse the response
	var inputItemsResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &inputItemsResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %w", err))
	}

	// Set the resource ID based on the response ID and current timestamp
	d.SetId(responseID)

	// Set has_more if present
	if hasMore, ok := inputItemsResponse["has_more"].(bool); ok {
		if err := d.Set("has_more", hasMore); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set first_id if present
	if firstID, ok := inputItemsResponse["first_id"].(string); ok {
		if err := d.Set("first_id", firstID); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set last_id if present
	if lastID, ok := inputItemsResponse["last_id"].(string); ok {
		if err := d.Set("last_id", lastID); err != nil {
			return diag.FromErr(err)
		}
	}

	// Process input items if present
	if data, ok := inputItemsResponse["data"].([]interface{}); ok {
		inputItems := make([]map[string]interface{}, 0, len(data))
		for _, item := range data {
			if itemMap, ok := item.(map[string]interface{}); ok {
				inputItem := make(map[string]interface{})

				// Extract common fields
				if itemType, ok := itemMap["type"].(string); ok {
					inputItem["type"] = itemType
				}

				if id, ok := itemMap["id"].(string); ok {
					inputItem["id"] = id
				}

				if role, ok := itemMap["role"].(string); ok {
					inputItem["role"] = role
				}

				if status, ok := itemMap["status"].(string); ok {
					inputItem["status"] = status
				}

				// Extract content text from content array
				if content, ok := itemMap["content"].([]interface{}); ok && len(content) > 0 {
					if contentItem, ok := content[0].(map[string]interface{}); ok {
						if contentType, ok := contentItem["type"].(string); ok {
							if contentType == "input_text" || contentType == "text" {
								if text, ok := contentItem["text"].(string); ok {
									inputItem["content"] = text
								}
							}
						}
					}
				}

				inputItems = append(inputItems, inputItem)
			}
		}

		if err := d.Set("input_items", inputItems); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}
