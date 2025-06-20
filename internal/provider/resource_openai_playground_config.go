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

// PlaygroundConfigResponse represents the API response for an OpenAI playground configuration
type PlaygroundConfigResponse struct {
	ID        string                 `json:"id"`
	Object    string                 `json:"object"`
	CreatedAt int                    `json:"created_at"`
	UpdatedAt int                    `json:"updated_at"`
	Name      string                 `json:"name"`
	Settings  map[string]interface{} `json:"settings"`
}

// PlaygroundConfigRequest represents the request to create or update a playground configuration
type PlaygroundConfigRequest struct {
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings"`
}

// resourceOpenAIPlaygroundConfig returns a schema.Resource that represents a resource for OpenAI playground configurations.
func resourceOpenAIPlaygroundConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIPlaygroundConfigCreate,
		ReadContext:   resourceOpenAIPlaygroundConfigRead,
		UpdateContext: resourceOpenAIPlaygroundConfigUpdate,
		DeleteContext: resourceOpenAIPlaygroundConfigDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the playground configuration",
			},
			"settings": {
				Type:        schema.TypeMap,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The settings for the playground configuration",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the playground configuration was created",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the playground configuration was last updated",
			},
		},
	}
}

// resourceOpenAIPlaygroundConfigCreate creates a new OpenAI playground configuration.
func resourceOpenAIPlaygroundConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)
	settingsRaw := d.Get("settings").(map[string]interface{})

	// Prepare the settings map
	settings := make(map[string]interface{})
	for k, v := range settingsRaw {
		settings[k] = v
	}

	// Prepare the request
	requestBody, err := json.Marshal(PlaygroundConfigRequest{
		Name:     name,
		Settings: settings,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshaling request: %s", err))
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/playground/configs", client.APIURL)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/playground/configs", client.APIURL)
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
	var config PlaygroundConfigResponse
	err = json.Unmarshal(responseBody, &config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the ID
	d.SetId(config.ID)

	// Set computed attributes
	if err := d.Set("created_at", time.Unix(int64(config.CreatedAt), 0).Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", time.Unix(int64(config.UpdatedAt), 0).Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIPlaygroundConfigRead reads an existing OpenAI playground configuration.
func resourceOpenAIPlaygroundConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	configID := d.Id()
	if configID == "" {
		return diag.Errorf("config ID is empty")
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/playground/configs/%s", client.APIURL, configID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/playground/configs/%s", client.APIURL, configID)
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
	var config PlaygroundConfigResponse
	err = json.Unmarshal(responseBody, &config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update the state
	if err := d.Set("name", config.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("settings", config.Settings); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", time.Unix(int64(config.CreatedAt), 0).Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", time.Unix(int64(config.UpdatedAt), 0).Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIPlaygroundConfigUpdate updates an existing OpenAI playground configuration.
func resourceOpenAIPlaygroundConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	configID := d.Id()
	if configID == "" {
		return diag.Errorf("config ID is empty")
	}

	name := d.Get("name").(string)
	settingsRaw := d.Get("settings").(map[string]interface{})

	// Prepare the settings map
	settings := make(map[string]interface{})
	for k, v := range settingsRaw {
		settings[k] = v
	}

	// Prepare the request
	requestBody, err := json.Marshal(PlaygroundConfigRequest{
		Name:     name,
		Settings: settings,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshaling request: %s", err))
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/playground/configs/%s", client.APIURL, configID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/playground/configs/%s", client.APIURL, configID)
	}

	// Create HTTP request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
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
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, errorResponse.Error.Message))
	}

	// Parse the response
	var config PlaygroundConfigResponse
	err = json.Unmarshal(responseBody, &config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update computed attributes
	if err := d.Set("updated_at", time.Unix(int64(config.UpdatedAt), 0).Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIPlaygroundConfigDelete deletes an OpenAI playground configuration.
func resourceOpenAIPlaygroundConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	configID := d.Id()
	if configID == "" {
		return diag.Errorf("config ID is empty")
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/playground/configs/%s", client.APIURL, configID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/playground/configs/%s", client.APIURL, configID)
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
