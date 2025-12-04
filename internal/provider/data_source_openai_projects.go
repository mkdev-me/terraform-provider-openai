package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProjectsListResponse represents the API response for a list of OpenAI projects
type ProjectsListResponse struct {
	Object  string            `json:"object"`   // Object type, always "list"
	Data    []ProjectResponse `json:"data"`     // Array of project objects
	FirstID string            `json:"first_id"` // ID of the first project in the list
	LastID  string            `json:"last_id"`  // ID of the last project in the list
	HasMore bool              `json:"has_more"` // Whether there are more projects to fetch
}

// dataSourceOpenAIProjects returns a schema.Resource that represents a data source for OpenAI projects.
func dataSourceOpenAIProjects() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIProjectsRead,
		Schema: map[string]*schema.Schema{
			"admin_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Admin API key for authentication. If not provided, the provider's default API key will be used.",
			},
			"projects": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the project",
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
					},
				},
				Description: "List of available projects",
			},
		},
	}
}

// dataSourceOpenAIProjectsRead handles the read operation for the OpenAI projects data source.
// It retrieves the list of available projects from the OpenAI API and updates the Terraform state.
func dataSourceOpenAIProjectsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get custom API key if provided
	apiKey := client.APIKey
	if v, ok := d.GetOk("admin_key"); ok {
		apiKey = v.(string)
		if len(apiKey) >= 4 {
			fmt.Printf("Using custom Admin API key provided in data source (first 4 chars: %s)\n", apiKey[:4])
		} else {
			fmt.Printf("Using custom Admin API key provided in data source (length: %d)\n", len(apiKey))
		}
	} else {
		if len(apiKey) >= 4 {
			fmt.Printf("Using default client API key (first 4 chars: %s)\n", apiKey[:4])
		} else {
			fmt.Printf("Using default client API key (length: %d)\n", len(apiKey))
		}
	}

	// Verify API key is not empty
	if apiKey == "" {
		return diag.Errorf("API key cannot be empty")
	}

	// Collect all projects with pagination
	var allProjects []ProjectResponse
	var after string
	hasMore := true
	limit := 100 // Use maximum limit per page for efficiency

	for hasMore {
		// Construct the API URL with query parameters
		baseURL := fmt.Sprintf("%s/v1/organization/projects", client.APIURL)
		// If the APIURL already contains /v1, adjust the URL construction
		if strings.Contains(client.APIURL, "/v1") {
			baseURL = fmt.Sprintf("%s/organization/projects", client.APIURL)
		}

		// Parse URL and add query parameters
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error parsing URL: %s", err))
		}

		// Add query parameters
		q := parsedURL.Query()
		q.Set("limit", strconv.Itoa(limit))
		if after != "" {
			q.Set("after", after)
		}
		parsedURL.RawQuery = q.Encode()

		// Debug output to help troubleshoot API URL
		fmt.Printf("Making API request to URL: %s\n", parsedURL.String())
		fmt.Printf("OpenAI client config: API URL=%s, Organization ID=%s\n", client.APIURL, client.OrganizationID)

		// Create HTTP request
		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating request: %s", err))
		}

		// Add required headers
		req.Header.Set("Authorization", "Bearer "+apiKey)
		if client.OrganizationID != "" {
			req.Header.Set("OpenAI-Organization", client.OrganizationID)
		}

		fmt.Printf("Request headers: %v\n", req.Header)

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
		var projectsList ProjectsListResponse
		err = json.Unmarshal(responseBody, &projectsList)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
		}

		// Append projects from this page
		allProjects = append(allProjects, projectsList.Data...)

		// Check if there are more pages
		hasMore = projectsList.HasMore
		if hasMore && projectsList.LastID != "" {
			after = projectsList.LastID
		} else {
			hasMore = false
		}

		fmt.Printf("Fetched %d projects (total so far: %d), has_more: %v\n", len(projectsList.Data), len(allProjects), hasMore)
	}

	// Generate a unique ID for this data source invocation
	d.SetId("openai-projects-" + time.Now().Format(time.RFC3339))

	// Convert response to format expected by schema
	projects := make([]map[string]interface{}, 0, len(allProjects))
	for _, project := range allProjects {
		projectMap := map[string]interface{}{
			"id":         project.ID,
			"name":       project.Name,
			"status":     project.Status,
			"created_at": project.CreatedAt,
		}
		projects = append(projects, projectMap)
	}

	// Set the projects in the schema
	if err := d.Set("projects", projects); err != nil {
		return diag.Errorf("Error setting projects: %s", err)
	}

	fmt.Printf("Successfully fetched all %d projects\n", len(allProjects))

	return diag.Diagnostics{}
}
