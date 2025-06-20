package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// MessageResponse represents the API response for an OpenAI message.
// It contains all the fields returned by the OpenAI API when creating or retrieving a message.
type MessageResponse struct {
	ID          string                 `json:"id"`
	Object      string                 `json:"object"`
	CreatedAt   int                    `json:"created_at"`
	ThreadID    string                 `json:"thread_id"`
	Role        string                 `json:"role"`
	Content     []MessageContent       `json:"content"`
	AssistantID string                 `json:"assistant_id,omitempty"`
	RunID       string                 `json:"run_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Attachments []MessageAttachment    `json:"attachments,omitempty"`
}

// MessageContent represents the content of a message.
// It can contain text or other types of content with their respective annotations.
type MessageContent struct {
	Type string              `json:"type"`
	Text *MessageContentText `json:"text,omitempty"`
}

// MessageContentText represents the text content of a message.
// It includes the text value and any associated annotations.
type MessageContentText struct {
	Value       string        `json:"value"`
	Annotations []interface{} `json:"annotations,omitempty"`
}

// MessageAttachment represents an attachment in a message.
type MessageAttachment struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	AssistantID string `json:"assistant_id,omitempty"`
	CreatedAt   int    `json:"created_at"`
}

// MessageCreateRequest represents the payload for creating a message in the OpenAI API.
// It contains all the fields that can be set when creating a new message.
type MessageCreateRequest struct {
	Role        string                 `json:"role"`
	Content     string                 `json:"content"`
	Attachments []AttachmentRequest    `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AttachmentRequest represents a file attachment in a message creation request.
type AttachmentRequest struct {
	FileID string        `json:"file_id"`
	Tools  []ToolRequest `json:"tools"`
}

// ToolRequest represents a tool in an attachment
type ToolRequest struct {
	Type string `json:"type"`
}

func resourceOpenAIMessage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIMessageCreate,
		ReadContext:   resourceOpenAIMessageRead,
		UpdateContext: resourceOpenAIMessageUpdate,
		DeleteContext: resourceOpenAIMessageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIMessageImportState,
		},
		Schema: map[string]*schema.Schema{
			"thread_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the thread to add the message to",
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The role of the entity that is creating the message. Currently only 'user' is supported.",
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The content of the message",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of key-value pairs that can be attached to the message",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the message was created",
			},
			"assistant_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "If applicable, the ID of the assistant that authored this message",
			},
			"run_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "If applicable, the ID of the run associated with this message",
			},
			"attachments": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the file to attach",
						},
						"tools": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The type of tool that should use this file",
									},
								},
							},
							Description: "The tools that should use this file",
						},
					},
				},
				Description: "A list of file attachments to include with the message",
			},
		},
	}
}

func resourceOpenAIMessageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the thread ID
	threadID := d.Get("thread_id").(string)

	// Prepare the request
	createRequest := &MessageCreateRequest{
		Role:    d.Get("role").(string),
		Content: d.Get("content").(string),
	}

	// Add attachments if present
	if attachmentsRaw, ok := d.GetOk("attachments"); ok {
		attachmentsList := attachmentsRaw.([]interface{})
		attachments := make([]AttachmentRequest, 0, len(attachmentsList))

		for _, attachmentRaw := range attachmentsList {
			attachmentMap := attachmentRaw.(map[string]interface{})

			// Process tools
			toolsRaw := attachmentMap["tools"].([]interface{})
			tools := make([]ToolRequest, 0, len(toolsRaw))
			for _, t := range toolsRaw {
				toolMap := t.(map[string]interface{})
				tools = append(tools, ToolRequest{
					Type: toolMap["type"].(string),
				})
			}

			attachments = append(attachments, AttachmentRequest{
				FileID: attachmentMap["file_id"].(string),
				Tools:  tools,
			})
		}

		createRequest.Attachments = attachments
	}

	// Add metadata if present
	if metadataRaw, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		for k, v := range metadataRaw.(map[string]interface{}) {
			metadata[k] = v.(string)
		}
		createRequest.Metadata = metadata
	}

	// Convert request to JSON
	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing message request: %s", err))
	}

	// Prepare HTTP request
	url := fmt.Sprintf("%s/threads/%s/messages", client.APIURL, threadID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
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
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s", err))
		}
		return diag.FromErr(fmt.Errorf("error creating message: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var messageResponse MessageResponse
	if err := json.Unmarshal(respBody, &messageResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Save the ID and other data in state
	d.SetId(messageResponse.ID)
	if err := d.Set("created_at", messageResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("assistant_id", messageResponse.AssistantID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("run_id", messageResponse.RunID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metadata", messageResponse.Metadata); err != nil {
		return diag.FromErr(err)
	}

	// Process attachments if present
	// Don't try to set attachments from response, keep the original request attachments
	// The API returns a different structure than what we send

	return diag.Diagnostics{}
}

// resourceOpenAIMessageRead handles reading an existing OpenAI message.
// It retrieves the message information from OpenAI and updates the Terraform state.
func resourceOpenAIMessageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the thread ID and message ID
	threadID := d.Get("thread_id").(string)
	messageID := d.Id()

	// Create URL to get message information
	url := fmt.Sprintf("%s/threads/%s/messages/%s", client.APIURL, threadID, messageID)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// If message doesn't exist (404), return an error
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s", err))
		}
		return diag.FromErr(fmt.Errorf("error reading message: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse response
	var messageResponse MessageResponse
	if err := json.Unmarshal(respBody, &messageResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update state
	if err := d.Set("thread_id", messageResponse.ThreadID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("role", messageResponse.Role); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", messageResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("assistant_id", messageResponse.AssistantID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("run_id", messageResponse.RunID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("metadata", messageResponse.Metadata); err != nil {
		return diag.FromErr(err)
	}

	// Handle content
	if len(messageResponse.Content) > 0 {
		for _, content := range messageResponse.Content {
			if content.Type == "text" && content.Text != nil {
				if err := d.Set("content", content.Text.Value); err != nil {
					return diag.FromErr(err)
				}
				break
			}
		}
	}

	// Handle attachments - we don't set them from response as they have a different structure
	// than what we expect in the schema for sending

	return diag.Diagnostics{}
}

func resourceOpenAIMessageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the IDs
	messageID := d.Id()
	threadID := d.Get("thread_id").(string)

	// Check what has changed
	if d.HasChange("metadata") {
		// Prepare the update request
		updateRequest := map[string]interface{}{}

		// Add metadata to the request
		if metadataRaw, ok := d.GetOk("metadata"); ok {
			metadata := make(map[string]interface{})
			for k, v := range metadataRaw.(map[string]interface{}) {
				metadata[k] = v.(string)
			}
			updateRequest["metadata"] = metadata
		}

		// Convert request to JSON
		reqBody, err := json.Marshal(updateRequest)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error serializing update request: %s", err))
		}

		// Prepare HTTP request
		url := fmt.Sprintf("%s/threads/%s/messages/%s", client.APIURL, threadID, messageID)
		req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating request: %s", err))
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+client.APIKey)
		req.Header.Set("OpenAI-Beta", "assistants=v2")
		if client.OrganizationID != "" {
			req.Header.Set("OpenAI-Organization", client.OrganizationID)
		}

		// Make the request
		resp, err := http.DefaultClient.Do(req)
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
			var errorResponse ErrorResponse
			if err := json.Unmarshal(respBody, &errorResponse); err != nil {
				return diag.FromErr(fmt.Errorf("error parsing error response: %s", err))
			}
			return diag.FromErr(fmt.Errorf("error updating message: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
		}

		// Parse the response
		var messageResponse MessageResponse
		if err := json.Unmarshal(respBody, &messageResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
		}

		// Update the state
		if err := d.Set("created_at", messageResponse.CreatedAt); err != nil {
			return diag.FromErr(err)
		}

		// Update metadata if present
		if len(messageResponse.Metadata) > 0 {
			metadata := make(map[string]string)
			for k, v := range messageResponse.Metadata {
				metadata[k] = fmt.Sprintf("%v", v)
			}
			if err := d.Set("metadata", metadata); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return diag.Diagnostics{}
}

func resourceOpenAIMessageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the IDs
	messageID := d.Id()
	threadID := d.Get("thread_id").(string)

	// Prepare HTTP request
	url := fmt.Sprintf("%s/threads/%s/messages/%s", client.APIURL, threadID, messageID)
	req, err := http.NewRequest("DELETE", url, nil)
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// If the message doesn't exist, it's not an error
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s", err))
		}
		return diag.FromErr(fmt.Errorf("error deleting message: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Remove the ID from state
	d.SetId("")

	return diag.Diagnostics{}
}

// resourceOpenAIMessageImportState handles importing an existing OpenAI message.
// It requires the ID to be in the format "thread_id:message_id"
func resourceOpenAIMessageImportState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Parse the import ID
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return nil, fmt.Errorf("invalid import ID format. Expected 'thread_id:message_id', got %s", d.Id())
	}

	threadID := idParts[0]
	messageID := idParts[1]

	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Create URL to get message information
	url := fmt.Sprintf("%s/threads/%s/messages/%s", client.APIURL, threadID, messageID)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %s", err)
	}
	defer resp.Body.Close()

	// If message doesn't exist (404), return error
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("message with ID %s not found in thread %s", messageID, threadID)
	}

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %s", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return nil, fmt.Errorf("error parsing error response: %s", err)
		}
		return nil, fmt.Errorf("error retrieving message: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message)
	}

	// Parse response
	var messageResponse MessageResponse
	if err := json.Unmarshal(respBody, &messageResponse); err != nil {
		return nil, fmt.Errorf("error parsing response: %s", err)
	}

	// Set ID
	d.SetId(messageID)

	// Set fields
	if err := d.Set("thread_id", threadID); err != nil {
		return nil, err
	}
	if err := d.Set("role", messageResponse.Role); err != nil {
		return nil, err
	}
	if err := d.Set("created_at", messageResponse.CreatedAt); err != nil {
		return nil, err
	}
	if err := d.Set("assistant_id", messageResponse.AssistantID); err != nil {
		return nil, err
	}
	if err := d.Set("run_id", messageResponse.RunID); err != nil {
		return nil, err
	}
	if err := d.Set("metadata", messageResponse.Metadata); err != nil {
		return nil, err
	}

	// Handle content
	if len(messageResponse.Content) > 0 {
		for _, content := range messageResponse.Content {
			if content.Type == "text" && content.Text != nil {
				if err := d.Set("content", content.Text.Value); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	return []*schema.ResourceData{d}, nil
}
