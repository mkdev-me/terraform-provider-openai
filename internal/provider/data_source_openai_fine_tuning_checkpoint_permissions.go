package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOpenAIFineTuningCheckpointPermissions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIFineTuningCheckpointPermissionsRead,
		Schema: map[string]*schema.Schema{
			"checkpoint_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the checkpoint to fetch permissions for",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "Number of permissions to retrieve (default: 20)",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier for the last permission from the previous pagination request",
			},
			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique identifier for this permission",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of object (checkpoint.permission)",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Unix timestamp when the permission was created",
						},
						"checkpoint_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the checkpoint",
						},
						"project_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of project IDs that have access",
						},
						"allow_view": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether viewing the checkpoint is allowed",
						},
						"allow_create": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether creating from the checkpoint is allowed",
						},
					},
				},
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more permissions to retrieve",
			},
		},
	}
}

func dataSourceOpenAIFineTuningCheckpointPermissionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	checkpointID := d.Get("checkpoint_id").(string)

	// Build the query parameters
	queryParams := url.Values{}

	if v, ok := d.GetOk("limit"); ok {
		queryParams.Set("limit", strconv.Itoa(v.(int)))
	}

	if v, ok := d.GetOk("after"); ok {
		queryParams.Set("after", v.(string))
	}

	// Build the URL
	apiURL := fmt.Sprintf("%s/fine_tuning/checkpoints/%s/permissions", client.APIURL, checkpointID)
	if len(queryParams) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, queryParams.Encode())
	}

	// Log the URL for debugging
	fmt.Printf("[DEBUG] Fetching checkpoint permissions with URL: %s\n", apiURL)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Use the provider's API key
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))

	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error requesting checkpoint permissions: %s", err))
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		// Handle specific HTTP status codes gracefully
		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			// Set the ID anyway so Terraform has something to reference
			d.SetId(fmt.Sprintf("checkpoint-permissions-%s", checkpointID))

			// Set has_more to false
			if err := d.Set("has_more", false); err != nil {
				return diag.FromErr(err)
			}

			// Set empty permissions list
			if err := d.Set("permissions", []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}

			// Parse error message if possible
			var errorMessage string
			var errorResponse struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}

			if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error.Message != "" {
				errorMessage = errorResponse.Error.Message
			} else {
				errorMessage = string(body)
			}

			// Return a warning instead of error
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Insufficient permissions for checkpoint",
					Detail:   fmt.Sprintf("Unable to access checkpoint permissions for '%s': %s. This may be because you need admin privileges or the 'api.fine_tuning.checkpoints.read' scope. Returning empty permissions list.", checkpointID, errorMessage),
				},
			}

		case http.StatusNotFound:
			// Set the ID anyway so Terraform has something to reference
			d.SetId(fmt.Sprintf("checkpoint-permissions-%s", checkpointID))

			// Set has_more to false
			if err := d.Set("has_more", false); err != nil {
				return diag.FromErr(err)
			}

			// Set empty permissions list
			if err := d.Set("permissions", []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}

			// Return a warning instead of error
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Checkpoint not found",
					Detail:   fmt.Sprintf("Checkpoint with ID '%s' could not be found. This may be because it has been deleted or has expired. Returning empty permissions list.", checkpointID),
				},
			}
		}

		// For other error types, return the normal error
		return diag.FromErr(fmt.Errorf("error fetching checkpoint permissions: %s - %s", resp.Status, string(body)))
	}

	// Parse the response
	var permissionsResponse struct {
		Object  string                   `json:"object"`
		Data    []map[string]interface{} `json:"data"`
		HasMore bool                     `json:"has_more"`
	}

	if err := json.Unmarshal(body, &permissionsResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the ID
	d.SetId(fmt.Sprintf("checkpoint-permissions-%s", checkpointID))

	// Set has_more
	if err := d.Set("has_more", permissionsResponse.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %s", err))
	}

	// Convert permissions to the format expected by the schema
	permissions := make([]map[string]interface{}, 0, len(permissionsResponse.Data))
	for _, permission := range permissionsResponse.Data {
		permissionMap := map[string]interface{}{
			"id":            permission["id"],
			"object":        permission["object"],
			"created_at":    int(permission["created_at"].(float64)),
			"checkpoint_id": permission["fine_tuned_model_checkpoint"],
		}

		// Handle project IDs
		if projectIDs, ok := permission["project_ids"].([]interface{}); ok {
			projectIDStrings := make([]string, len(projectIDs))
			for i, pid := range projectIDs {
				projectIDStrings[i] = pid.(string)
			}
			permissionMap["project_ids"] = projectIDStrings
		}

		// Handle permission flags
		if allowView, ok := permission["allow_view"].(bool); ok {
			permissionMap["allow_view"] = allowView
		} else {
			permissionMap["allow_view"] = true // Default to true if not specified
		}

		if allowCreate, ok := permission["allow_create"].(bool); ok {
			permissionMap["allow_create"] = allowCreate
		} else {
			permissionMap["allow_create"] = true // Default to true if not specified
		}

		permissions = append(permissions, permissionMap)
	}

	if err := d.Set("permissions", permissions); err != nil {
		return diag.FromErr(fmt.Errorf("error setting permissions: %s", err))
	}

	return nil
}
