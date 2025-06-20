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
)

// dataSourceOpenAIModelResponse provides a data source to retrieve OpenAI model response details.
func dataSourceOpenAIModelResponse() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIModelResponseRead,
		Schema: map[string]*schema.Schema{
			"response_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the model response to retrieve",
			},
			"include": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Additional fields to include in the response. Valid values include: usage.input_tokens_details, usage.output_tokens_details, file_search_results, web_search_results, message_files.url, computation_files.url",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the response was created",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the response (e.g., 'completed')",
			},
			"model": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the model used for the response",
			},
			"temperature": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The temperature used for generation",
			},
			"top_p": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The top_p value used for generation",
			},
			"output": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The output of the model response",
			},
			"usage": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Token usage statistics for the request",
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
					},
				},
				Description: "The input items for the model response",
			},
		},
		Description: "Data source for retrieving OpenAI model response information",
	}
}

// dataSourceOpenAIModelResponseRead reads information about an existing OpenAI model response.
func dataSourceOpenAIModelResponseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	// Build the URL
	var url string
	if strings.HasSuffix(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/responses/%s", client.APIURL, responseID)
	} else {
		url = fmt.Sprintf("%s/v1/responses/%s", client.APIURL, responseID)
	}

	// Add optional include parameter if specified
	if v, ok := d.GetOk("include"); ok {
		includeList := v.([]interface{})
		if len(includeList) > 0 {
			// Validate include parameter values
			validIncludeValues := map[string]bool{
				"usage.input_tokens_details":  true,
				"usage.output_tokens_details": true,
				"file_search_results":         true,
				"web_search_results":          true,
				"message_files.url":           true,
				"computation_files.url":       true,
			}

			includeParams := make([]string, 0, len(includeList))
			for _, item := range includeList {
				includeValue := item.(string)
				if _, ok := validIncludeValues[includeValue]; !ok {
					return diag.FromErr(fmt.Errorf("invalid include value: %s. Valid values are: usage.input_tokens_details, usage.output_tokens_details, file_search_results, web_search_results, message_files.url, computation_files.url", includeValue))
				}
				includeParams = append(includeParams, includeValue)
			}
			url = fmt.Sprintf("%s?include=%s", url, strings.Join(includeParams, ","))
		}
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

	// Parse the response directly as a map
	var modelResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &modelResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %w", err))
	}

	// Set the ID
	d.SetId(responseID)

	// Set basic fields
	if createdAt, ok := modelResponse["created_at"].(float64); ok {
		if err := d.Set("created_at", int(createdAt)); err != nil {
			return diag.FromErr(err)
		}
	}

	if status, ok := modelResponse["status"].(string); ok {
		if err := d.Set("status", status); err != nil {
			return diag.FromErr(err)
		}
	}

	if model, ok := modelResponse["model"].(string); ok {
		if err := d.Set("model", model); err != nil {
			return diag.FromErr(err)
		}
	}

	if temperature, ok := modelResponse["temperature"].(float64); ok {
		if err := d.Set("temperature", temperature); err != nil {
			return diag.FromErr(err)
		}
	}

	if topP, ok := modelResponse["top_p"].(float64); ok {
		if err := d.Set("top_p", topP); err != nil {
			return diag.FromErr(err)
		}
	}

	// Extract text from the output structure
	if output, ok := modelResponse["output"].([]interface{}); ok && len(output) > 0 {
		outputMap := make(map[string]string)

		if firstMessage, ok := output[0].(map[string]interface{}); ok {
			// Add role if available
			if role, ok := firstMessage["role"].(string); ok {
				outputMap["role"] = role
			}

			// Extract text from content array
			if content, ok := firstMessage["content"].([]interface{}); ok && len(content) > 0 {
				if firstContent, ok := content[0].(map[string]interface{}); ok {
					if text, ok := firstContent["text"].(string); ok {
						outputMap["text"] = text
					}
				}
			}
		}

		if err := d.Set("output", outputMap); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle usage
	if usage, ok := modelResponse["usage"].(map[string]interface{}); ok {
		usageMap := make(map[string]string)

		// Extract usage metrics
		for k, v := range usage {
			switch value := v.(type) {
			case float64:
				usageMap[k] = strconv.Itoa(int(value))
			case int:
				usageMap[k] = strconv.Itoa(value)
			case map[string]interface{}:
				// For nested maps like input_tokens_details
				detailsStr := fmt.Sprintf("map[%s", "")
				for dk, dv := range value {
					detailsStr += fmt.Sprintf("%s:%v ", dk, dv)
				}
				detailsStr += "]"
				usageMap[k] = detailsStr
			default:
				usageMap[k] = fmt.Sprintf("%v", value)
			}
		}

		if err := d.Set("usage", usageMap); err != nil {
			return diag.FromErr(err)
		}
	}

	// Fetch input items in a separate request
	inputItemsURL := ""
	if strings.HasSuffix(client.APIURL, "/v1") {
		inputItemsURL = fmt.Sprintf("%s/responses/%s/input_items", client.APIURL, responseID)
	} else {
		inputItemsURL = fmt.Sprintf("%s/v1/responses/%s/input_items", client.APIURL, responseID)
	}

	inputItemsReq, err := http.NewRequest(http.MethodGet, inputItemsURL, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating input items request: %w", err))
	}

	// Add headers for input items request
	inputItemsReq.Header.Set("Content-Type", "application/json")
	inputItemsReq.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		inputItemsReq.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the input items request
	inputItemsResp, err := http.DefaultClient.Do(inputItemsReq)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making input items request: %w", err))
	}
	defer inputItemsResp.Body.Close()

	// Read the input items response
	inputItemsResponseBody, err := io.ReadAll(inputItemsResp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading input items response: %w", err))
	}

	// Only process input items if the request was successful
	if inputItemsResp.StatusCode == http.StatusOK {
		var inputItemsResponse map[string]interface{}
		if err := json.Unmarshal(inputItemsResponseBody, &inputItemsResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing input items response: %w", err))
		}

		// Process input items data if present
		if data, ok := inputItemsResponse["data"].([]interface{}); ok {
			inputItems := make([]map[string]interface{}, 0, len(data))
			for _, item := range data {
				if itemMap, ok := item.(map[string]interface{}); ok {
					inputItem := make(map[string]interface{})

					// Extract item properties
					if itemType, ok := itemMap["type"].(string); ok {
						inputItem["type"] = itemType
					}

					if id, ok := itemMap["id"].(string); ok {
						inputItem["id"] = id
					}

					if role, ok := itemMap["role"].(string); ok {
						inputItem["role"] = role
					}

					// Extract text content from the content array
					if content, ok := itemMap["content"].([]interface{}); ok && len(content) > 0 {
						if contentItem, ok := content[0].(map[string]interface{}); ok {
							if text, ok := contentItem["text"].(string); ok {
								inputItem["content"] = text
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
	}

	return diag.Diagnostics{}
}
