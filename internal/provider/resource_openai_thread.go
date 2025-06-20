package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ThreadResponse represents the API response for an OpenAI thread.
// It contains the thread's identifier, creation timestamp, and associated metadata.
type ThreadResponse struct {
	ID        string                 `json:"id"`
	Object    string                 `json:"object"`
	CreatedAt int                    `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ThreadCreateRequest represents the request payload for creating a thread in the OpenAI API.
// It can include initial messages and metadata for the thread.
type ThreadCreateRequest struct {
	Messages []ThreadMessage        `json:"messages,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ThreadMessage represents a message within a thread.
// Each message has a role, content, optional file attachments, and metadata.
type ThreadMessage struct {
	Role     string                 `json:"role"`
	Content  string                 `json:"content"`
	FileIDs  []string               `json:"file_ids,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// resourceOpenAIThread defines the schema and CRUD operations for OpenAI threads.
// This resource allows users to create, read, update, and delete threads through the OpenAI API.
func resourceOpenAIThread() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIThreadCreate,
		ReadContext:   resourceOpenAIThreadRead,
		UpdateContext: resourceOpenAIThreadUpdate,
		DeleteContext: resourceOpenAIThreadDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"messages": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The role of the entity that is creating the message. Currently only 'user' is supported.",
						},
						"content": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The content of the message",
						},
						"file_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "A list of file IDs that the message should use",
						},
						"metadata": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Set of key-value pairs that can be attached to the message",
						},
					},
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of key-value pairs that can be attached to the thread",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the thread was created",
			},
		},
	}
}

// resourceOpenAIThreadCreate handles the creation of a new OpenAI thread.
// It creates a thread with optional initial messages and metadata.
func resourceOpenAIThreadCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Prepare the request
	createRequest := &ThreadCreateRequest{}

	// Add messages if present
	if messagesRaw, ok := d.GetOk("messages"); ok {
		messagesList := messagesRaw.([]interface{})
		messages := make([]ThreadMessage, 0, len(messagesList))

		for _, msgRaw := range messagesList {
			msgMap := msgRaw.(map[string]interface{})

			msg := ThreadMessage{
				Role:    msgMap["role"].(string),
				Content: msgMap["content"].(string),
			}

			// Add file_ids if present
			if fileIDsRaw, ok := msgMap["file_ids"]; ok {
				fileIDsList := fileIDsRaw.([]interface{})
				fileIDs := make([]string, 0, len(fileIDsList))

				for _, id := range fileIDsList {
					fileIDs = append(fileIDs, id.(string))
				}

				msg.FileIDs = fileIDs
			}

			// Add metadata if present
			if metadataRaw, ok := msgMap["metadata"]; ok {
				metadata := make(map[string]interface{})
				for k, v := range metadataRaw.(map[string]interface{}) {
					metadata[k] = v
				}
				msg.Metadata = metadata
			}

			messages = append(messages, msg)
		}

		createRequest.Messages = messages
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
		return diag.FromErr(fmt.Errorf("error serializing thread request: %s", err))
	}

	// Prepare HTTP request
	url := fmt.Sprintf("%s/threads", client.APIURL)
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
		return diag.FromErr(fmt.Errorf("error creating thread: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse response
	var threadResponse ThreadResponse
	if err := json.Unmarshal(respBody, &threadResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Save ID and other data to state
	d.SetId(threadResponse.ID)
	if err := d.Set("created_at", threadResponse.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
	}

	return diag.Diagnostics{}
}

// resourceOpenAIThreadRead handles reading an existing OpenAI thread.
// It retrieves the thread information from OpenAI and updates the Terraform state.
func resourceOpenAIThreadRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the thread ID
	threadID := d.Id()
	if threadID == "" {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Prepare HTTP request
	url := fmt.Sprintf("%s/threads/%s", client.APIURL, threadID)
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

	// If thread doesn't exist, remove from state
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
		return diag.FromErr(fmt.Errorf("error reading thread: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var threadResponse ThreadResponse
	if err := json.Unmarshal(respBody, &threadResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update thread attributes
	if err := d.Set("created_at", threadResponse.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
	}

	// Update metadata if present in the response
	if len(threadResponse.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range threadResponse.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		if err := d.Set("metadata", metadata); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set metadata: %v", err))
		}
	}

	return diag.Diagnostics{}
}

func resourceOpenAIThreadUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the thread ID
	threadID := d.Id()

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
		url := fmt.Sprintf("%s/threads/%s", client.APIURL, threadID)
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
			return diag.FromErr(fmt.Errorf("error updating thread: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
		}

		// Parse the response
		var threadResponse ThreadResponse
		if err := json.Unmarshal(respBody, &threadResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
		}

		// Update the state
		if err := d.Set("created_at", threadResponse.CreatedAt); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
		}

		// Update metadata if present in the response
		if len(threadResponse.Metadata) > 0 {
			metadata := make(map[string]string)
			for k, v := range threadResponse.Metadata {
				metadata[k] = fmt.Sprintf("%v", v)
			}
			if err := d.Set("metadata", metadata); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set metadata: %v", err))
			}
		}
	}

	return diag.Diagnostics{}
}

func resourceOpenAIThreadDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the thread ID
	threadID := d.Id()

	// Prepare HTTP request
	url := fmt.Sprintf("%s/threads/%s", client.APIURL, threadID)
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

	// If the thread doesn't exist, it's not an error
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
		return diag.FromErr(fmt.Errorf("error deleting thread: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Remove the ID from state
	d.SetId("")

	return diag.Diagnostics{}
}
