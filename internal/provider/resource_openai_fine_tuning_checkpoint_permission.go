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

// CheckpointPermissionRequest represents the request body for creating a checkpoint permission
type CheckpointPermissionRequest struct {
	ProjectIDs []string `json:"project_ids"`
}

func resourceOpenAIFineTuningCheckpointPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIFineTuningCheckpointPermissionCreate,
		ReadContext:   resourceOpenAIFineTuningCheckpointPermissionRead,
		DeleteContext: resourceOpenAIFineTuningCheckpointPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the checkpoint permission",
			},
			"checkpoint_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the checkpoint to set permissions for",
			},
			"project_ids": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of project IDs to grant access to",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Unix timestamp of when the permission was created",
			},
			"admin_api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Admin API key to use for this operation instead of the provider's API key",
			},
		},
	}
}

func resourceOpenAIFineTuningCheckpointPermissionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	checkpointID := d.Get("checkpoint_id").(string)

	// Get project IDs and handle nil values properly
	projectIDsRaw := d.Get("project_ids").([]interface{})
	var projectIDs []string
	for _, v := range projectIDsRaw {
		if v == nil {
			continue // Skip nil values
		}
		strVal, ok := v.(string)
		if !ok {
			return diag.FromErr(fmt.Errorf("invalid project_id: %v is not a string", v))
		}
		strVal = strings.TrimSpace(strVal)
		if strVal != "" {
			projectIDs = append(projectIDs, strVal)
		}
	}

	// Ensure we have at least one project ID
	if len(projectIDs) == 0 {
		return diag.FromErr(fmt.Errorf("at least one valid project_id is required"))
	}

	// Create request
	request := CheckpointPermissionRequest{
		ProjectIDs: projectIDs,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if we have an admin API key specified in the resource config
	apiKey := client.APIKey
	if adminKey, ok := d.GetOk("admin_api_key"); ok && adminKey.(string) != "" {
		apiKey = adminKey.(string)
	}

	// Ensure we have an API key
	if apiKey == "" {
		return diag.FromErr(fmt.Errorf("No API key provided. Checkpoint permissions require an admin API key with api.fine_tuning.checkpoints.write scope"))
	}

	apiURL := fmt.Sprintf("%s/fine_tuning/checkpoints/%s/permissions", client.APIURL, checkpointID)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating checkpoint permission: %s", err))
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == http.StatusUnauthorized {
			return diag.FromErr(fmt.Errorf("error creating checkpoint permission: 401 Unauthorized - %s. Make sure you are using an admin API key with the api.fine_tuning.checkpoints.write scope", string(body)))
		}
		return diag.FromErr(fmt.Errorf("error creating checkpoint permission: %s - %s", resp.Status, string(body)))
	}

	// Print full response for debugging
	fmt.Printf("[DEBUG] Full response body: %s\n", string(body))

	// Parse the response
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Log full response structure for debugging
	fmt.Printf("[DEBUG] Response structure: %+v\n", responseData)

	// Check if response is an array of permissions in a "data" field
	if dataArray, ok := responseData["data"].([]interface{}); ok && len(dataArray) > 0 {
		// It's a list response, extract the first permission
		if permission, ok := dataArray[0].(map[string]interface{}); ok {
			if permissionID, ok := permission["id"].(string); ok {
				d.SetId(permissionID)

				// Set created_at if available
				if createdAt, ok := permission["created_at"].(float64); ok {
					d.Set("created_at", int(createdAt))
				}

				return resourceOpenAIFineTuningCheckpointPermissionRead(ctx, d, m)
			}
		}
	}

	// If we get here, try to extract ID directly
	permissionID, ok := responseData["id"].(string)
	if ok {
		d.SetId(permissionID)

		// Set created_at if available
		if createdAt, ok := responseData["created_at"].(float64); ok {
			d.Set("created_at", int(createdAt))
		}

		return resourceOpenAIFineTuningCheckpointPermissionRead(ctx, d, m)
	}

	// For checkpoint permissions, sometimes the ID is in different formats
	// Log each possible field for debugging
	for key, value := range responseData {
		fmt.Printf("[DEBUG] Response field: %s = %v\n", key, value)
	}

	// If we've gotten here, we couldn't find an ID
	// Use the checkpoint_id + first project_id as a synthetic ID
	if len(projectIDs) > 0 {
		syntheticID := fmt.Sprintf("perm_%s_%s", checkpointID, projectIDs[0])
		fmt.Printf("[DEBUG] Using synthetic ID: %s\n", syntheticID)
		d.SetId(syntheticID)
		return resourceOpenAIFineTuningCheckpointPermissionRead(ctx, d, m)
	}

	return diag.FromErr(fmt.Errorf("missing ID in response: %s", string(body)))
}

// Helper function to mask the API key for debugging
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 4 {
		return "****"
	}
	return apiKey[:4] + "****"
}

func resourceOpenAIFineTuningCheckpointPermissionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Log the current state for debugging import issues
	fmt.Printf("[DEBUG] Reading checkpoint permission with ID: %s\n", d.Id())

	// In case we're importing, try to get the checkpoint_id from the configuration
	// If it's not available, we can't fetch the data properly
	checkpointID, checkpointExists := d.GetOk("checkpoint_id")
	if !checkpointExists {
		// During import, we need to leave the checkpoint_id empty
		// This will cause a diff to appear, which is expected - the user needs to specify the checkpoint_id
		fmt.Printf("[DEBUG] checkpoint_id not available during import. User must specify this in the configuration.\n")

		// If we're in an import operation and only have the ID, we need to set enough
		// information to allow terraform refresh to succeed, but indicate a diff is needed
		if d.Id() != "" {
			fmt.Printf("[DEBUG] Import operation detected. Setting minimal state for ID: %s\n", d.Id())
			// We can at least set the ID to allow the import to complete
			// User will need to specify the checkpoint_id and project_ids in their configuration
			return nil
		}
		return diag.Errorf("checkpoint_id is required")
	}

	// Check if we have an admin API key specified in the resource config
	apiKey := client.APIKey
	if adminKey, ok := d.GetOk("admin_api_key"); ok && adminKey.(string) != "" {
		fmt.Printf("[DEBUG] Using admin API key from resource config for reading\n")
		apiKey = adminKey.(string)
	} else {
		fmt.Printf("[DEBUG] Using API key from provider config for reading\n")
	}

	// Ensure we have an API key
	if apiKey == "" {
		return diag.FromErr(fmt.Errorf("No API key provided. Checkpoint permissions require an admin API key with api.fine_tuning.checkpoints.write scope"))
	}

	// List permissions for this checkpoint
	apiURL := fmt.Sprintf("%s/fine_tuning/checkpoints/%s/permissions", client.APIURL, checkpointID.(string))
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading checkpoint permissions: %s", err))
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %s", err))
	}

	// Print response for debugging, especially useful during import
	fmt.Printf("[DEBUG] Response status for reading permissions: %s\n", resp.Status)
	fmt.Printf("[DEBUG] Response body for reading permissions: %s\n", string(body))

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return diag.FromErr(fmt.Errorf("error reading checkpoint permissions: 401 Unauthorized - %s. Make sure you are using an admin API key with the api.fine_tuning.checkpoints.write scope", string(body)))
		}
		return diag.FromErr(fmt.Errorf("error reading checkpoint permissions: %s - %s", resp.Status, string(body)))
	}

	// Parse the response which is a list
	var responseData struct {
		Data []struct {
			ID        string `json:"id"`
			CreatedAt int    `json:"created_at"`
			ProjectID string `json:"project_id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &responseData); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Find the permission with matching ID
	permissionID := d.Id()
	for _, permission := range responseData.Data {
		if permission.ID == permissionID {
			d.Set("created_at", permission.CreatedAt)

			// Set project_ids as a list
			projectIds := []string{permission.ProjectID}
			d.Set("project_ids", projectIds)

			// Set checkpoint_id if it's not already set
			if !checkpointExists {
				d.Set("checkpoint_id", checkpointID)
			}

			return nil
		}
	}

	// Permission not found, mark it as deleted
	d.SetId("")
	return nil
}

func resourceOpenAIFineTuningCheckpointPermissionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	checkpointID := d.Get("checkpoint_id").(string)
	permissionID := d.Id()

	// Check if we have an admin API key specified in the resource config
	apiKey := client.APIKey
	if adminKey, ok := d.GetOk("admin_api_key"); ok && adminKey.(string) != "" {
		fmt.Printf("[DEBUG] Using admin API key from resource config for deletion\n")
		apiKey = adminKey.(string)
	} else {
		fmt.Printf("[DEBUG] Using API key from provider config for deletion\n")
	}

	apiURL := fmt.Sprintf("%s/fine_tuning/checkpoints/%s/permissions/%s", client.APIURL, checkpointID, permissionID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting checkpoint permission: %s", err))
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %s", err))
	}

	// If the permission is not found, consider it deleted
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	// Check for other errors
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return diag.FromErr(fmt.Errorf("error deleting checkpoint permission: %s - %s", resp.Status, string(body)))
	}

	// Successfully deleted
	d.SetId("")
	return nil
}
