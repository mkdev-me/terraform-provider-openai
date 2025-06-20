package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProjectResponse represents the API response for an OpenAI project
type ProjectResponse struct {
	ID          string      `json:"id"`
	Object      string      `json:"object"`
	Name        string      `json:"name"`
	CreatedAt   int         `json:"created_at"`
	Status      string      `json:"status"`
	UsageLimits UsageLimits `json:"usage_limits"`
}

// UsageLimits represents the usage limits for a project
type UsageLimits struct {
	MaxMonthlyDollars   float64 `json:"max_monthly_dollars"`
	MaxParallelRequests int     `json:"max_parallel_requests"`
	MaxTokens           int     `json:"max_tokens"`
}

// dataSourceOpenAIProject returns a schema.Resource that represents a data source for an OpenAI project.
// This data source allows users to retrieve information about a specific OpenAI project.
func dataSourceOpenAIProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIProjectRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project to retrieve",
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
				Description: "The name of the project",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the project",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Timestamp when the project was created",
			},
			"usage_limits": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_monthly_dollars": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Maximum monthly spend in dollars",
						},
						"max_parallel_requests": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of parallel requests allowed",
						},
						"max_tokens": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of tokens per request",
						},
					},
				},
				Description: "Usage limits for the project",
			},
		},
	}
}

// dataSourceOpenAIProjectRead handles the read operation for the OpenAI project data source.
// It retrieves information about a specific project from the OpenAI API.
func dataSourceOpenAIProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		return diag.Errorf("project_id cannot be empty")
	}

	// Get custom API key if provided
	apiKey := client.APIKey
	if v, ok := d.GetOk("admin_key"); ok {
		apiKey = v.(string)
	}

	// Set the ID to the project ID
	d.SetId(projectID)

	// Construct the API URL - using the correct organization/projects path
	url := fmt.Sprintf("%s/v1/organization/projects/%s", client.APIURL, projectID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/organization/projects/%s", client.APIURL, projectID)
	}

	fmt.Printf("Getting project with ID: %s\n", projectID)
	fmt.Printf("Using URL: %s\n", strings.Replace(url, client.APIURL, "", 1))
	fmt.Printf("OpenAI client config: API URL=%s, Organization ID=%s\n", client.APIURL, client.OrganizationID)
	fmt.Printf("Making API request: GET %s\n", url)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
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
	var project ProjectResponse
	err = json.Unmarshal(responseBody, &project)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the project details in the schema
	if err := d.Set("name", project.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", project.Status); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", project.CreatedAt); err != nil {
		return diag.FromErr(err)
	}

	// Set usage limits
	usageLimits := []map[string]interface{}{
		{
			"max_monthly_dollars":   project.UsageLimits.MaxMonthlyDollars,
			"max_parallel_requests": project.UsageLimits.MaxParallelRequests,
			"max_tokens":            project.UsageLimits.MaxTokens,
		},
	}
	if err := d.Set("usage_limits", usageLimits); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
