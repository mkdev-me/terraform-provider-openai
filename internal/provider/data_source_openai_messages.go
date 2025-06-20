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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// MessagesListResponse represents the API response for a list of messages
type MessagesListResponse struct {
	Object  string            `json:"object"`
	Data    []MessageResponse `json:"data"`
	FirstID string            `json:"first_id"`
	LastID  string            `json:"last_id"`
	HasMore bool              `json:"has_more"`
}

// dataSourceOpenAIMessages returns a schema.Resource that represents a data source for listing OpenAI messages within a thread.
func dataSourceOpenAIMessages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIMessagesRead,
		Schema: map[string]*schema.Schema{
			"thread_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the thread containing the messages",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "A limit on the number of messages to be returned. Maximum of 100.",
			},
			"order": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "desc",
				Description: "Sort order by creation timestamp. One of: asc, desc (default)",
				ValidateFunc: validation.StringInSlice([]string{
					"asc",
					"desc",
				}, false),
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A cursor for pagination. Return objects after this message ID.",
			},
			"before": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A cursor for pagination. Return objects before this message ID.",
			},
			"messages": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identifier of this message",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The object type, always 'thread.message'",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp for when the message was created",
						},
						"thread_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the thread that contains the message",
						},
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role of the entity that created the message",
						},
						"content": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The content of the message",
						},
						"assistant_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "If applicable, the ID of the assistant that authored this message",
						},
						"run_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "If applicable, the ID of the run that generated this message",
						},
						"metadata": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Set of key-value pairs attached to the message",
						},
						"attachments": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the attachment",
									},
									"type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The type of the attachment",
									},
									"assistant_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "If applicable, the ID of the assistant this attachment is associated with",
									},
									"created_at": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The timestamp for when the attachment was created",
									},
								},
							},
							Description: "A list of attachments in the message",
						},
					},
				},
				Description: "List of messages in the thread",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more items available in the list",
			},
			"first_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the first message in the list",
			},
			"last_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the last message in the list",
			},
		},
	}
}

// dataSourceOpenAIMessagesRead handles the read operation for the OpenAI messages data source.
// It retrieves a list of messages in a thread from the OpenAI API based on the specified filters.
func dataSourceOpenAIMessagesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %s", err))
	}

	// Get parameters
	threadID := d.Get("thread_id").(string)
	limit := d.Get("limit").(int)
	order := d.Get("order").(string)

	// Build the query parameters
	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	params.Add("order", order)

	if after, ok := d.GetOk("after"); ok {
		params.Add("after", after.(string))
	}

	if before, ok := d.GetOk("before"); ok {
		params.Add("before", before.(string))
	}

	// Build URL for the request
	url := fmt.Sprintf("%s/threads/%s/messages?%s", client.APIURL, threadID, params.Encode())

	// Prepare HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := client.HTTPClient.Do(req)
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
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s, status code: %d", err, resp.StatusCode))
		}
		return diag.FromErr(fmt.Errorf("error retrieving messages: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var messagesResponse MessagesListResponse
	if err := json.Unmarshal(respBody, &messagesResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing messages response: %s", err))
	}

	// Generate a unique ID for this resource
	d.SetId(fmt.Sprintf("%s-%d-%s", threadID, limit, order))

	// Set the basic properties
	if err := d.Set("has_more", messagesResponse.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %s", err))
	}
	if err := d.Set("first_id", messagesResponse.FirstID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting first_id: %s", err))
	}
	if err := d.Set("last_id", messagesResponse.LastID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting last_id: %s", err))
	}

	// Process messages
	messages := make([]map[string]interface{}, 0, len(messagesResponse.Data))
	for _, message := range messagesResponse.Data {
		messages = append(messages, convertMessageToMap(message))
	}

	if err := d.Set("messages", messages); err != nil {
		return diag.FromErr(fmt.Errorf("error setting messages: %s", err))
	}

	return nil
}

// Convert file_ids to schema-compatible format
func convertMessageToMap(message MessageResponse) map[string]interface{} {
	m := map[string]interface{}{
		"id":         message.ID,
		"object":     message.Object,
		"created_at": message.CreatedAt,
		"thread_id":  message.ThreadID,
		"role":       message.Role,
	}

	// Extract content text if available
	if len(message.Content) > 0 {
		for _, content := range message.Content {
			if content.Type == "text" && content.Text != nil {
				m["content"] = content.Text.Value
				break
			}
		}
	}

	// Add assistant-specific fields if present
	if message.AssistantID != "" {
		m["assistant_id"] = message.AssistantID
	}
	if message.RunID != "" {
		m["run_id"] = message.RunID
	}

	// Add metadata if present
	if len(message.Metadata) > 0 {
		m["metadata"] = message.Metadata
	}

	// Add attachments if present
	if len(message.Attachments) > 0 {
		attachments := make([]map[string]interface{}, 0, len(message.Attachments))
		for _, attachment := range message.Attachments {
			attachmentMap := map[string]interface{}{
				"id":         attachment.ID,
				"type":       attachment.Type,
				"created_at": attachment.CreatedAt,
			}
			if attachment.AssistantID != "" {
				attachmentMap["assistant_id"] = attachment.AssistantID
			}
			attachments = append(attachments, attachmentMap)
		}
		m["attachments"] = attachments
	}

	return m
}
