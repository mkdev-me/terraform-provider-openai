package provider

import (
	"bytes"
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

// ProjectAPIKeyResponse represents the API response for an OpenAI project API key
type ProjectAPIKeyResponse struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Name       string `json:"name"`
	CreatedAt  int    `json:"created_at"`
	LastUsedAt int    `json:"last_used_at,omitempty"`
	Value      string `json:"value,omitempty"` // Only returned on creation
}

// ProjectAPIKeyCreateRequest represents the request to create a project API key
type ProjectAPIKeyCreateRequest struct {
	Name string `json:"name,omitempty"`
}

// resourceOpenAIProjectAPIKey returns the schema and CRUD operations for the OpenAI Project API Key resource
func resourceOpenAIProjectAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIProjectAPIKeyCreate,
		ReadContext:   resourceOpenAIProjectAPIKeyRead,
		DeleteContext: resourceOpenAIProjectAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIProjectAPIKeyImport,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the project this API key belongs to",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the API key",
			},
			// Computed fields
			"api_key_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the API key",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the API key was created",
			},
			"last_used_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the API key was last used",
			},
			"api_key_value": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The actual API key value (only available upon creation)",
			},
		},
	}
}

// resourceOpenAIProjectAPIKeyCreate creates a new OpenAI project API key
func resourceOpenAIProjectAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		return diag.Errorf("project_id cannot be empty")
	}

	// Prepare create request
	createRequest := ProjectAPIKeyCreateRequest{}
	if name, ok := d.GetOk("name"); ok {
		createRequest.Name = name.(string)
	}

	requestBody, err := json.Marshal(createRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request: %s", err))
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/projects/%s/api_keys", client.APIURL, projectID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/projects/%s/api_keys", client.APIURL, projectID)
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
	var apiKey ProjectAPIKeyResponse
	err = json.Unmarshal(responseBody, &apiKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the resource ID in the format "project_id:api_key_id"
	d.SetId(fmt.Sprintf("%s:%s", projectID, apiKey.ID))

	// Set the computed values in the state
	if err := d.Set("api_key_id", apiKey.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", apiKey.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", time.Unix(int64(apiKey.CreatedAt), 0).Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if apiKey.LastUsedAt > 0 {
		if err := d.Set("last_used_at", time.Unix(int64(apiKey.LastUsedAt), 0).Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
	}
	if apiKey.Value != "" {
		if err := d.Set("api_key_value", apiKey.Value); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIProjectAPIKeyRead reads an existing OpenAI project API key
func resourceOpenAIProjectAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the ID from the resource
	id := d.Id()
	if id == "" {
		return diag.Errorf("API key ID is empty")
	}

	// Split the ID into project_id and api_key_id
	var projectID, apiKeyID string
	if strings.Contains(id, ":") {
		parts := strings.Split(id, ":")
		if len(parts) != 2 {
			return diag.Errorf("invalid ID format, expected 'project_id:api_key_id'")
		}
		projectID = parts[0]
		apiKeyID = parts[1]

		// Set these values in the resource data
		if err := d.Set("project_id", projectID); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("api_key_id", apiKeyID); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// If we just have a single ID, it might be just the API key ID
		// Try to get the project_id from the resource data
		projectID = d.Get("project_id").(string)
		apiKeyID = id

		if projectID == "" {
			return diag.Errorf("project_id is required")
		}

		if err := d.Set("api_key_id", apiKeyID); err != nil {
			return diag.FromErr(err)
		}
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/projects/%s/api_keys/%s", client.APIURL, projectID, apiKeyID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/projects/%s/api_keys/%s", client.APIURL, projectID, apiKeyID)
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
	var apiKey ProjectAPIKeyResponse
	err = json.Unmarshal(responseBody, &apiKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update the resource data
	if err := d.Set("name", apiKey.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", time.Unix(int64(apiKey.CreatedAt), 0).Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if apiKey.LastUsedAt > 0 {
		if err := d.Set("last_used_at", time.Unix(int64(apiKey.LastUsedAt), 0).Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIProjectAPIKeyDelete deletes an OpenAI project API key
func resourceOpenAIProjectAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the ID from the resource
	id := d.Id()
	if id == "" {
		return diag.Errorf("API key ID is empty")
	}

	// Split the ID into project_id and api_key_id
	var projectID, apiKeyID string
	if strings.Contains(id, ":") {
		parts := strings.Split(id, ":")
		if len(parts) != 2 {
			return diag.Errorf("invalid ID format, expected 'project_id:api_key_id'")
		}
		projectID = parts[0]
		apiKeyID = parts[1]
	} else {
		// If we just have a single ID, try to get the project_id from the resource data
		projectID = d.Get("project_id").(string)
		apiKeyID = id

		if projectID == "" {
			return diag.Errorf("project_id is required")
		}
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/projects/%s/api_keys/%s", client.APIURL, projectID, apiKeyID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/projects/%s/api_keys/%s", client.APIURL, projectID, apiKeyID)
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

// resourceOpenAIProjectAPIKeyImport imports an existing OpenAI project API key
func resourceOpenAIProjectAPIKeyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The import ID is expected to be in the format "project_id:api_key_id"
	id := d.Id()

	// Split the ID into project_id and api_key_id
	if strings.Contains(id, ":") {
		parts := strings.Split(id, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid ID format, expected 'project_id:api_key_id'")
		}

		projectID := parts[0]
		apiKeyID := parts[1]

		// Ensure the API key ID is in the expected format
		if !strings.HasPrefix(apiKeyID, "key_") {
			// This is likely a name, not an ID. For improved user experience,
			// we'll just accept the name as is since the OpenAI API doesn't allow
			// retrieving the actual key ID.
			fmt.Printf("Warning: API key ID '%s' does not start with 'key_'. "+
				"This may or may not work depending on how the OpenAI API validates IDs.\n", apiKeyID)
		}

		d.SetId(id)
		if err := d.Set("project_id", projectID); err != nil {
			return nil, fmt.Errorf("failed to set project_id: %v", err)
		}
		if err := d.Set("api_key_id", apiKeyID); err != nil {
			return nil, fmt.Errorf("failed to set api_key_id: %v", err)
		}

		// Set some default values for required fields in case the API read fails
		if err := d.Set("name", "imported-key"); err != nil {
			return nil, fmt.Errorf("failed to set name: %v", err)
		}

		// For imported keys, we'll simulate some values since the API might not provide them
		now := time.Now().Format(time.RFC3339)
		if err := d.Set("created_at", now); err != nil {
			return nil, fmt.Errorf("failed to set created_at: %v", err)
		}
	} else {
		return nil, fmt.Errorf("invalid ID format, expected 'project_id:api_key_id'")
	}

	// Try to read the resource, but don't fail the import if it can't be read
	diags := resourceOpenAIProjectAPIKeyRead(ctx, d, meta)
	if diags.HasError() {
		fmt.Printf("Warning: Could not read API key from OpenAI API. Using placeholder values instead.\n")
		// Don't return an error - we'll use the placeholder values we set above
	}

	return []*schema.ResourceData{d}, nil
}
