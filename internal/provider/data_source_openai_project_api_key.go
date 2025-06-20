package provider

import (
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

// dataSourceOpenAIProjectAPIKey returns a schema.Resource that represents a data source for an OpenAI project API key.
// This data source allows users to retrieve information about a specific API key within a project.
func dataSourceOpenAIProjectAPIKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIProjectAPIKeyRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project to which the API key belongs",
			},
			"api_key_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the API key to retrieve",
			},
			"admin_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Admin API key for authentication. If not provided, the provider's default API key will be used.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the API key",
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
		},
	}
}

// dataSourceOpenAIProjectAPIKeyRead handles the read operation for the OpenAI project API key data source.
// It retrieves information about a specific API key from the OpenAI API.
func dataSourceOpenAIProjectAPIKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		return diag.Errorf("project_id cannot be empty")
	}

	apiKeyID := d.Get("api_key_id").(string)
	if apiKeyID == "" {
		return diag.Errorf("api_key_id cannot be empty")
	}

	// Get custom admin key if provided
	customKey := client.APIKey
	if v, ok := d.GetOk("admin_key"); ok {
		customKey = v.(string)
	}

	// Set the ID to the combined project_id:api_key_id
	d.SetId(fmt.Sprintf("%s:%s", projectID, apiKeyID))

	// Construct the API URL with the correct path for organization/projects/api_keys
	url := fmt.Sprintf("%s/v1/organization/projects/%s/api_keys/%s", client.APIURL, projectID, apiKeyID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/organization/projects/%s/api_keys/%s", client.APIURL, projectID, apiKeyID)
	}

	fmt.Printf("Getting project API key with ID: %s for project: %s\n", apiKeyID, projectID)
	fmt.Printf("Using URL: %s\n", url)
	fmt.Printf("OpenAI client config: API URL=%s, Organization ID=%s\n", client.APIURL, client.OrganizationID)
	fmt.Printf("Making API request: GET %s\n", url)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+customKey)
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
	var apiKey ProjectAPIKeyResponse
	err = json.Unmarshal(responseBody, &apiKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the API key details in the schema
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
