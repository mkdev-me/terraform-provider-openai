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

// ProjectResourcesResponse represents the API response for OpenAI project resources
type ProjectResourcesResponse struct {
	APIKeys         []ProjectAPIKeyResponse `json:"api_keys"`
	Assistants      []AssistantSummary      `json:"assistants"`
	Files           []FileResponse          `json:"files"`
	FineTunedModels []FineTunedModelSummary `json:"fine_tuned_models"`
	ServiceAccounts []ServiceAccountSummary `json:"service_accounts"`
}

// AssistantSummary represents a summary of an assistant
type AssistantSummary struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Name      string `json:"name"`
	Model     string `json:"model"`
}

// FineTunedModelSummary represents a summary of a fine-tuned model
type FineTunedModelSummary struct {
	ID             string `json:"id"`
	Object         string `json:"object"`
	CreatedAt      int    `json:"created_at"`
	FineTunedModel string `json:"fine_tuned_model"`
	Status         string `json:"status"`
}

// ServiceAccountSummary represents a summary of a service account
type ServiceAccountSummary struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

// dataSourceOpenAIProjectResources returns a schema.Resource that represents a data source for OpenAI project resources.
// This data source allows users to retrieve information about resources associated with a specific OpenAI project.
func dataSourceOpenAIProjectResources() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIProjectResourcesRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project to retrieve resources for",
			},
			// API Keys
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
				Description: "List of API keys for this project",
			},
			// Assistants
			"assistants": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the assistant",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the assistant",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The model used by the assistant",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the assistant was created",
						},
					},
				},
				Description: "List of assistants for this project",
			},
			// Files
			"files": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the file",
						},
						"filename": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the file",
						},
						"purpose": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The purpose of the file",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the file was created",
						},
					},
				},
				Description: "List of files for this project",
			},
			// Fine-tuned models
			"fine_tuned_models": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the fine-tuning job",
						},
						"fine_tuned_model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The fine-tuned model ID",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the fine-tuning job",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the fine-tuning job was created",
						},
					},
				},
				Description: "List of fine-tuned models for this project",
			},
			// Service accounts
			"service_accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the service account",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the service account",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The email of the service account",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the service account was created",
						},
					},
				},
				Description: "List of service accounts for this project",
			},
		},
	}
}

// dataSourceOpenAIProjectResourcesRead handles the read operation for the OpenAI project resources data source.
// It retrieves resources associated with a specific project from the OpenAI API.
func dataSourceOpenAIProjectResourcesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		return diag.Errorf("project_id cannot be empty")
	}

	// Set a unique ID for this data source
	d.SetId(fmt.Sprintf("%s-%s", projectID, time.Now().Format(time.RFC3339)))

	// Construct the API URL
	url := fmt.Sprintf("%s/v1/projects/%s/resources", client.APIURL, projectID)
	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/projects/%s/resources", client.APIURL, projectID)
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
	var resources ProjectResourcesResponse
	err = json.Unmarshal(responseBody, &resources)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set API keys
	apiKeys := make([]map[string]interface{}, 0, len(resources.APIKeys))
	for _, key := range resources.APIKeys {
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

	// Set assistants
	assistants := make([]map[string]interface{}, 0, len(resources.Assistants))
	for _, asst := range resources.Assistants {
		assistant := map[string]interface{}{
			"id":         asst.ID,
			"name":       asst.Name,
			"model":      asst.Model,
			"created_at": time.Unix(int64(asst.CreatedAt), 0).Format(time.RFC3339),
		}
		assistants = append(assistants, assistant)
	}
	if err := d.Set("assistants", assistants); err != nil {
		return diag.FromErr(err)
	}

	// Set files
	files := make([]map[string]interface{}, 0, len(resources.Files))
	for _, file := range resources.Files {
		fileMap := map[string]interface{}{
			"id":         file.ID,
			"filename":   file.Filename,
			"purpose":    file.Purpose,
			"created_at": time.Unix(int64(file.CreatedAt), 0).Format(time.RFC3339),
		}
		files = append(files, fileMap)
	}
	if err := d.Set("files", files); err != nil {
		return diag.FromErr(err)
	}

	// Set fine-tuned models
	fineTunedModels := make([]map[string]interface{}, 0, len(resources.FineTunedModels))
	for _, model := range resources.FineTunedModels {
		modelMap := map[string]interface{}{
			"id":               model.ID,
			"fine_tuned_model": model.FineTunedModel,
			"status":           model.Status,
			"created_at":       time.Unix(int64(model.CreatedAt), 0).Format(time.RFC3339),
		}
		fineTunedModels = append(fineTunedModels, modelMap)
	}
	if err := d.Set("fine_tuned_models", fineTunedModels); err != nil {
		return diag.FromErr(err)
	}

	// Set service accounts
	serviceAccounts := make([]map[string]interface{}, 0, len(resources.ServiceAccounts))
	for _, sa := range resources.ServiceAccounts {
		saMap := map[string]interface{}{
			"id":         sa.ID,
			"name":       sa.Name,
			"email":      sa.Email,
			"created_at": time.Unix(int64(sa.CreatedAt), 0).Format(time.RFC3339),
		}
		serviceAccounts = append(serviceAccounts, saMap)
	}
	if err := d.Set("service_accounts", serviceAccounts); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
