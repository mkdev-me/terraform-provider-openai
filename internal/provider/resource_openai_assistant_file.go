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

// AssistantFileResponse represents the API response for an OpenAI assistant file
type AssistantFileResponse struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	CreatedAt   int    `json:"created_at"`
	AssistantID string `json:"assistant_id"`
	FileID      string `json:"file_id"`
}

// AssistantFileCreateRequest represents the request to create an assistant file
type AssistantFileCreateRequest struct {
	FileID string `json:"file_id"`
}

// resourceOpenAIAssistantFile returns a schema.Resource that represents a resource for OpenAI assistant files.
func resourceOpenAIAssistantFile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIAssistantFileCreate,
		ReadContext:   resourceOpenAIAssistantFileRead,
		DeleteContext: resourceOpenAIAssistantFileDelete,
		Schema: map[string]*schema.Schema{
			"assistant_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the assistant to attach the file to",
			},
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the file to attach to the assistant",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Timestamp when the file was attached to the assistant",
			},
		},
	}
}

// resourceOpenAIAssistantFileCreate handles the creation of a new assistant file.
func resourceOpenAIAssistantFileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	assistantID := d.Get("assistant_id").(string)
	fileID := d.Get("file_id").(string)

	if assistantID == "" || fileID == "" {
		return diag.Errorf("assistant_id and file_id are required")
	}

	// Prepare the request body
	requestBody, err := json.Marshal(AssistantFileCreateRequest{
		FileID: fileID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshaling request: %s", err))
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/assistants/%s/files", client.APIURL, assistantID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/assistants/%s/files", client.APIURL, assistantID)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
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
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, errorResponse.Error.Message))
	}

	// Parse the response
	var assistantFile AssistantFileResponse
	err = json.Unmarshal(responseBody, &assistantFile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the ID
	d.SetId(assistantFile.ID)

	// Set computed attributes
	if err := d.Set("created_at", assistantFile.CreatedAt); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIAssistantFileRead handles reading an existing assistant file from the OpenAI API.
func resourceOpenAIAssistantFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	assistantID := d.Get("assistant_id").(string)
	fileID := d.Id()

	if assistantID == "" || fileID == "" {
		return diag.Errorf("assistant_id and file_id are required")
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/assistants/%s/files/%s", client.APIURL, assistantID, fileID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/assistants/%s/files/%s", client.APIURL, assistantID, fileID)
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// For 404 errors, we remove the resource from state
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, errorResponse.Error.Message))
	}

	// Parse the response
	var assistantFile AssistantFileResponse
	err = json.Unmarshal(responseBody, &assistantFile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update the state
	if err := d.Set("created_at", assistantFile.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("file_id", assistantFile.FileID); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIAssistantFileDelete handles deleting an assistant file.
func resourceOpenAIAssistantFileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	assistantID := d.Get("assistant_id").(string)
	fileID := d.Id()

	if assistantID == "" || fileID == "" {
		return diag.Errorf("assistant_id and file_id are required")
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/assistants/%s/files/%s", client.APIURL, assistantID, fileID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/assistants/%s/files/%s", client.APIURL, assistantID, fileID)
	}

	// Create HTTP request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// For 404 errors, we just remove the resource from state
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Check for other errors
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error reading response: %s", err))
		}

		var errorResponse ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, errorResponse.Error.Message))
	}

	// Remove the resource from state
	d.SetId("")

	return diag.Diagnostics{}
}
