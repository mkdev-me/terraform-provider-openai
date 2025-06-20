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

// dataSourceOpenAIMessage returns a schema.Resource that represents a data source for a single OpenAI message.
func dataSourceOpenAIMessage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIMessageRead,
		Schema: map[string]*schema.Schema{
			"thread_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the thread that contains the message",
			},
			"message_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the message to retrieve",
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
	}
}

// dataSourceOpenAIMessageRead handles the read operation for the OpenAI message data source.
// It retrieves details about a specific message from the OpenAI API.
func dataSourceOpenAIMessageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %s", err))
	}

	// Get thread ID and message ID from the schema
	threadID := d.Get("thread_id").(string)
	messageID := d.Get("message_id").(string)

	// Build URL for the request
	url := fmt.Sprintf("%s/threads/%s/messages/%s", client.APIURL, threadID, messageID)

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
		return diag.FromErr(fmt.Errorf("error retrieving message: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var messageResponse MessageResponse
	if err := json.Unmarshal(respBody, &messageResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing message response: %s", err))
	}

	// Set the ID in the resource data
	d.SetId(messageResponse.ID)

	// Set various message properties
	if err := d.Set("thread_id", messageResponse.ThreadID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting thread_id: %s", err))
	}
	if err := d.Set("role", messageResponse.Role); err != nil {
		return diag.FromErr(fmt.Errorf("error setting role: %s", err))
	}
	if err := d.Set("created_at", messageResponse.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created_at: %s", err))
	}
	if err := d.Set("assistant_id", messageResponse.AssistantID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting assistant_id: %s", err))
	}
	if err := d.Set("run_id", messageResponse.RunID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting run_id: %s", err))
	}

	// Set metadata if present
	if messageResponse.Metadata != nil {
		if err := d.Set("metadata", messageResponse.Metadata); err != nil {
			return diag.FromErr(fmt.Errorf("error setting metadata: %s", err))
		}
	}

	// Handle content
	if len(messageResponse.Content) > 0 {
		for _, content := range messageResponse.Content {
			if content.Type == "text" && content.Text != nil {
				if err := d.Set("content", content.Text.Value); err != nil {
					return diag.FromErr(fmt.Errorf("error setting content: %s", err))
				}
				break
			}
		}
	}

	// Handle attachments
	if len(messageResponse.Attachments) > 0 {
		attachments := make([]map[string]interface{}, 0, len(messageResponse.Attachments))
		for _, attachment := range messageResponse.Attachments {
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
		if err := d.Set("attachments", attachments); err != nil {
			return diag.FromErr(fmt.Errorf("error setting attachments: %s", err))
		}
	}

	return nil
}
