package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ChatCompletionMessage represents a message in a chat completion.
type ChatCompletionMessageResponse struct {
	Object  string                  `json:"object"`
	Data    []ChatCompletionMessage `json:"data"`
	FirstID string                  `json:"first_id"`
	LastID  string                  `json:"last_id"`
	HasMore bool                    `json:"has_more"`
}

// dataSourceOpenAIChatCompletionMessages returns a schema.Resource that represents a data source for retrieving
// messages from a specific OpenAI chat completion.
func dataSourceOpenAIChatCompletionMessages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIChatCompletionMessagesRead,
		Schema: map[string]*schema.Schema{
			"completion_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the chat completion to retrieve messages from (format: chat_xxx)",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier for the last message from the previous pagination request",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 100),
				Description:  "Number of messages to retrieve (defaults to 20, max 100)",
			},
			"order": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "asc",
				ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
				Description:  "Sort order for messages by timestamp. Use 'asc' for ascending order or 'desc' for descending order. Defaults to 'asc'.",
			},
			"messages": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of messages from the chat completion",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role of the message author (system, user, assistant, or function)",
						},
						"content": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The content of the message",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the author of this message",
						},
						"function_call": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The function call in the message",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the function to call",
									},
									"arguments": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The arguments to call the function with, as a JSON string",
									},
								},
							},
						},
					},
				},
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more messages to retrieve",
			},
			"first_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the first message in the response",
			},
			"last_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the last message in the response",
			},
		},
	}
}

// dataSourceOpenAIChatCompletionMessagesRead handles the read operation for the OpenAI chat completion messages data source.
// It retrieves messages from a specific chat completion.
func dataSourceOpenAIChatCompletionMessagesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get OpenAI client: %s", err))
	}

	completionID := d.Get("completion_id").(string)
	url := fmt.Sprintf("/v1/chat/completions/%s/messages", completionID)

	// Add query parameters if provided
	params := make(map[string]string)
	if limit, ok := d.GetOk("limit"); ok {
		params["limit"] = fmt.Sprintf("%d", limit.(int))
	}
	if order, ok := d.GetOk("order"); ok {
		params["order"] = order.(string)
	}
	if after, ok := d.GetOk("after"); ok {
		params["after"] = after.(string)
	}
	if before, ok := d.GetOk("before"); ok {
		params["before"] = before.(string)
	}

	// Add query parameters to the URL
	if len(params) > 0 {
		url += "?"
		for key, value := range params {
			url += fmt.Sprintf("%s=%s&", key, value)
		}
		url = url[:len(url)-1] // Remove trailing &
	}

	respBody, err := client.DoRequest("GET", url, nil)
	if err != nil {
		// Check if it's a not found error, and handle gracefully
		if strings.Contains(err.Error(), "not found") {
			// Set ID to a derived ID to prevent Terraform from failing
			d.SetId(fmt.Sprintf("%s-messages", completionID))
			_ = d.Set("has_more", false)
			// Return empty messages array
			if err := d.Set("messages", []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Chat completion not found",
					Detail:   fmt.Sprintf("Chat completion with ID '%s' could not be found. This may be because it has expired or was deleted from the OpenAI Chat Completions Store.", completionID),
				},
			}
		}
		return diag.Errorf("Error retrieving messages for chat completion with ID '%s': %s", completionID, err)
	}

	var response ChatCompletionMessageResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing chat completion messages response: %s", err))
	}

	d.SetId(fmt.Sprintf("%s-messages", completionID))
	if err := d.Set("has_more", response.HasMore); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("first_id", response.FirstID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_id", response.LastID); err != nil {
		return diag.FromErr(err)
	}

	// Process messages
	if len(response.Data) > 0 {
		messages := make([]map[string]interface{}, 0, len(response.Data))
		for _, msg := range response.Data {
			message := map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			}

			if msg.Name != "" {
				message["name"] = msg.Name
			}

			// Add function_call if present
			if msg.FunctionCall != nil {
				message["function_call"] = []map[string]interface{}{
					{
						"name":      msg.FunctionCall.Name,
						"arguments": msg.FunctionCall.Arguments,
					},
				}
			}

			messages = append(messages, message)
		}

		if err := d.Set("messages", messages); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Set empty messages array for consistency
		if err := d.Set("messages", []map[string]interface{}{}); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}
