package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mkdev-me/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

// ListAssistantsResponse represents the API response for listing OpenAI assistants.
type ListAssistantsResponse struct {
	Object  string                     `json:"object"`   // Object type, always "list"
	Data    []client.AssistantResponse `json:"data"`     // Array of assistant objects
	FirstID string                     `json:"first_id"` // ID of the first assistant in the list
	LastID  string                     `json:"last_id"`  // ID of the last assistant in the list
	HasMore bool                       `json:"has_more"` // Whether there are more assistants to fetch
}

// dataSourceOpenAIAssistants returns a schema.Resource that represents a data source for OpenAI assistants.
// This data source allows users to retrieve a paginated list of assistants with optional filtering.
func dataSourceOpenAIAssistants() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIAssistantsRead,
		Schema: map[string]*schema.Schema{
			"order": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "desc",
				Description: "Sort order by the created_at timestamp. Can be 'asc' or 'desc'.",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "Number of assistants to retrieve (max 100)",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier for after which to retrieve assistants",
			},
			"before": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier for before which to retrieve assistants",
			},
			"assistants": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The assistant identifier",
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
					},
				},
				Description: "List of available assistants",
			},
			"first_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the first assistant in the list",
			},
			"last_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the last assistant in the list",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more assistants available",
			},
		},
	}
}

// dataSourceOpenAIAssistantsRead handles the read operation for the OpenAI assistants data source.
// It retrieves a list of assistants from the OpenAI API based on the provided filters and updates the Terraform state.
func dataSourceOpenAIAssistantsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %s", err))
	}

	// Build query parameters
	queryParams := url.Values{}

	// Add parameters if present
	if order, ok := d.GetOk("order"); ok {
		queryParams.Add("order", order.(string))
	}

	if limit, ok := d.GetOk("limit"); ok {
		queryParams.Add("limit", strconv.Itoa(limit.(int)))
	}

	if after, ok := d.GetOk("after"); ok {
		queryParams.Add("after", after.(string))
	}

	if before, ok := d.GetOk("before"); ok {
		queryParams.Add("before", before.(string))
	}

	// Build URL with parameters
	baseURL := fmt.Sprintf("%s/assistants", c.APIURL)
	if len(queryParams) > 0 {
		baseURL = fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
	}

	// Prepare HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	if c.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.OrganizationID)
	}

	// Make the request using the client's HTTP client
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s, status code: %d, body: %s",
				err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error listing assistants: %s - %s",
			errResp.Error.Type, errResp.Error.Message))
	}

	// Parse the response
	var assistantsResponse ListAssistantsResponse
	if err := json.Unmarshal(respBody, &assistantsResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set resource ID with timestamp
	d.SetId(fmt.Sprintf("assistants-%d", time.Now().Unix()))

	// Set meta information attributes
	if err := d.Set("first_id", assistantsResponse.FirstID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_id", assistantsResponse.LastID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("has_more", assistantsResponse.HasMore); err != nil {
		return diag.FromErr(err)
	}

	// Process the list of assistants
	assistants := make([]interface{}, 0, len(assistantsResponse.Data))
	for _, assistant := range assistantsResponse.Data {
		assistantMap := map[string]interface{}{
			"id":           assistant.ID,
			"object":       assistant.Object,
			"created_at":   assistant.CreatedAt,
			"name":         assistant.Name,
			"description":  assistant.Description,
			"model":        assistant.Model,
			"instructions": assistant.Instructions,
			"file_ids":     assistant.FileIDs,
		}

		// Process tools if present
		if len(assistant.Tools) > 0 {
			tools := make([]map[string]interface{}, 0, len(assistant.Tools))

			for _, tool := range assistant.Tools {
				toolMap := map[string]interface{}{
					"type": tool.Type,
				}

				// If type is "function", process function details
				if tool.Type == "function" && tool.Function != nil {
					function := map[string]interface{}{
						"name":       tool.Function.Name,
						"parameters": string(tool.Function.Parameters),
					}

					if tool.Function.Description != "" {
						function["description"] = tool.Function.Description
					}

					toolMap["function"] = []interface{}{function}
				}

				tools = append(tools, toolMap)
			}

			assistantMap["tools"] = tools
		}

		// Process metadata if present
		if len(assistant.Metadata) > 0 {
			metadata := make(map[string]string)
			for k, v := range assistant.Metadata {
				metadata[k] = fmt.Sprintf("%v", v)
			}
			assistantMap["metadata"] = metadata
		}

		assistants = append(assistants, assistantMap)
	}

	if err := d.Set("assistants", assistants); err != nil {
		return diag.FromErr(fmt.Errorf("error setting assistants: %s", err))
	}

	return diag.Diagnostics{}
}
