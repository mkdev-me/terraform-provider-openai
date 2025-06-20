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

// ProjectAPIKeysResponse represents the API response for a list of OpenAI project API keys
type ProjectAPIKeysResponse struct {
	Object string                  `json:"object"`
	Data   []ProjectAPIKeyResponse `json:"data"`
}

// dataSourceOpenAIProjectAPIKeys returns a schema.Resource that represents a data source for OpenAI project API keys.
// This data source allows users to retrieve a list of all API keys for a specific project.
func dataSourceOpenAIProjectAPIKeys() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIProjectAPIKeysRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project to retrieve API keys for",
			},
			"admin_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Admin API key for authentication. If not provided, the provider's default API key will be used.",
			},
			"api_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the API key",
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
				},
				Description: "List of API keys for the project",
			},
		},
	}
}

// dataSourceOpenAIProjectAPIKeysRead handles the read operation for the OpenAI project API keys data source.
// It retrieves a list of all API keys for a specific project from the OpenAI API.
func dataSourceOpenAIProjectAPIKeysRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		return diag.Errorf("project_id cannot be empty")
	}

	// Get custom admin key if provided
	customKey := client.APIKey
	if v, ok := d.GetOk("admin_key"); ok {
		customKey = v.(string)
	}

	// Set the ID to the project_id
	d.SetId(projectID)

	// Construct the API URL with the correct path for organization/projects/api_keys
	url := fmt.Sprintf("%s/v1/organization/projects/%s/api_keys", client.APIURL, projectID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/organization/projects/%s/api_keys", client.APIURL, projectID)
	}

	fmt.Printf("Getting API keys for project: %s\n", projectID)
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
	var apiKeysResponse ProjectAPIKeysResponse
	err = json.Unmarshal(responseBody, &apiKeysResponse)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Transform the API keys into a format appropriate for the schema
	apiKeys := make([]map[string]interface{}, 0, len(apiKeysResponse.Data))
	for _, key := range apiKeysResponse.Data {
		apiKey := map[string]interface{}{
			"id":         key.ID,
			"name":       key.Name,
			"created_at": time.Unix(int64(key.CreatedAt), 0).Format(time.RFC3339),
		}

		if key.LastUsedAt > 0 {
			apiKey["last_used_at"] = time.Unix(int64(key.LastUsedAt), 0).Format(time.RFC3339)
		}

		apiKeys = append(apiKeys, apiKey)
	}

	if err := d.Set("api_keys", apiKeys); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
