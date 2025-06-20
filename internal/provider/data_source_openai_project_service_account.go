package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIProjectServiceAccount returns the schema and operations for the OpenAI Project Service Account data source
func dataSourceOpenAIProjectServiceAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIProjectServiceAccountRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project to which the service account belongs",
			},
			"service_account_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the service account to retrieve",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Custom API key to use for this resource. If not provided, the provider's default API key will be used",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the service account",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp (in Unix time) when the service account was created",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role of the service account",
			},
			"api_key_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the API key associated with the service account",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

// dataSourceOpenAIProjectServiceAccountRead reads an existing OpenAI project service account
func dataSourceOpenAIProjectServiceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics

	// Get input values
	projectID := d.Get("project_id").(string)
	serviceAccountID := d.Get("service_account_id").(string)
	apiKey := d.Get("api_key").(string)

	// Get the service account
	tflog.Debug(ctx, "Reading OpenAI project service account", map[string]interface{}{
		"project_id":         projectID,
		"service_account_id": serviceAccountID,
	})

	serviceAccount, err := c.GetProjectServiceAccount(projectID, serviceAccountID, apiKey)
	if err != nil {
		// Handle permission errors gracefully
		if strings.Contains(err.Error(), "insufficient permissions") || strings.Contains(err.Error(), "Missing scopes") {
			tflog.Info(ctx, "Permission error reading service account", map[string]interface{}{
				"error": err.Error(),
			})
			return diag.FromErr(fmt.Errorf("cannot retrieve service account due to permission issues: %w", err))
		}

		// Check if the error is a 404, which means the service account is not found
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return diag.FromErr(fmt.Errorf("service account with ID %s not found in project %s", serviceAccountID, projectID))
		}

		return diag.FromErr(fmt.Errorf("error reading project service account: %w", err))
	}

	// Set ID to {project_id}:{service_account_id}
	d.SetId(fmt.Sprintf("%s:%s", projectID, serviceAccount.ID))

	// Set values in state
	if err := d.Set("name", serviceAccount.Name); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set name: %v", err))
	}
	if err := d.Set("created_at", serviceAccount.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
	}
	if err := d.Set("role", serviceAccount.Role); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role: %v", err))
	}

	// API key value won't be available during read, only the ID if available
	if serviceAccount.APIKey != nil {
		if err := d.Set("api_key_id", serviceAccount.APIKey.ID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set api_key_id: %v", err))
		}
	}

	return diags
}
