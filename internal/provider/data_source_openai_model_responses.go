package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIModelResponses provides a data source to list OpenAI model responses.
func dataSourceOpenAIModelResponses() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIModelResponsesRead,
		Schema: map[string]*schema.Schema{
			"filter_by_user": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter responses by user ID",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "Limit the number of responses returned (default: 20, max: 100)",
			},
			"order": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "desc",
				Description: "Sort order by created_at timestamp (asc or desc)",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Pagination cursor for fetching responses after this response ID",
			},
			"before": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Pagination cursor for fetching responses before this response ID",
			},
			"responses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the model response",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp when the response was created",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the model used for this response",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the response (e.g., 'completed')",
						},
						"usage": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Token usage statistics for the request",
						},
					},
				},
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more responses to fetch",
			},
		},
		Description: "Data source for listing OpenAI model responses",
	}
}

// dataSourceOpenAIModelResponsesRead reads a list of OpenAI model responses.
func dataSourceOpenAIModelResponsesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build the base URL
	var baseURL string
	if strings.Contains(client.APIURL, "/v1") {
		baseURL = fmt.Sprintf("%s/responses", client.APIURL)
	} else {
		baseURL = fmt.Sprintf("%s/v1/responses", client.APIURL)
	}

	// Add query parameters for filtering and pagination
	params := make([]string, 0)

	if v, ok := d.GetOk("filter_by_user"); ok {
		params = append(params, fmt.Sprintf("user=%s", v.(string)))
	}

	if v, ok := d.GetOk("limit"); ok {
		params = append(params, fmt.Sprintf("limit=%d", v.(int)))
	}

	if v, ok := d.GetOk("order"); ok {
		params = append(params, fmt.Sprintf("order=%s", v.(string)))
	}

	if v, ok := d.GetOk("after"); ok {
		params = append(params, fmt.Sprintf("after=%s", v.(string)))
	}

	if v, ok := d.GetOk("before"); ok {
		params = append(params, fmt.Sprintf("before=%s", v.(string)))
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
	var listResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &listResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %w", err))
	}

	// Set the resource ID as the current timestamp
	d.SetId(fmt.Sprintf("model_responses_%d", time.Now().Unix()))

	// Extract has_more flag
	if hasMore, ok := listResponse["has_more"].(bool); ok {
		if err := d.Set("has_more", hasMore); err != nil {
			return diag.FromErr(err)
		}
	}

	// Process the responses
	if data, ok := listResponse["data"].([]interface{}); ok {
		responses := make([]map[string]interface{}, 0, len(data))
		for _, item := range data {
			if responseItem, ok := item.(map[string]interface{}); ok {
				response := make(map[string]interface{})

				// Extract basic fields
				if id, ok := responseItem["id"].(string); ok {
					response["id"] = id
				}
				if createdAt, ok := responseItem["created_at"].(float64); ok {
					response["created_at"] = int(createdAt)
				}
				if model, ok := responseItem["model"].(string); ok {
					response["model"] = model
				}
				if status, ok := responseItem["status"].(string); ok {
					response["status"] = status
				}

				// Handle usage
				if usage, ok := responseItem["usage"].(map[string]interface{}); ok {
					usageMap := make(map[string]string)
					for k, v := range usage {
						switch value := v.(type) {
						case float64:
							usageMap[k] = fmt.Sprintf("%d", int(value))
						case int:
							usageMap[k] = fmt.Sprintf("%d", value)
						default:
							usageMap[k] = fmt.Sprintf("%v", value)
						}
					}
					response["usage"] = usageMap
				}

				responses = append(responses, response)
			}
		}

		if err := d.Set("responses", responses); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}
