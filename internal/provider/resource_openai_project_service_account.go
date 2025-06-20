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

// resourceOpenAIProjectServiceAccount returns the schema and CRUD operations for the OpenAI Project Service Account resource
func resourceOpenAIProjectServiceAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIProjectServiceAccountCreate,
		ReadContext:   resourceOpenAIProjectServiceAccountRead,
		DeleteContext: resourceOpenAIProjectServiceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the project to which the service account belongs",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the service account",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Custom API key to use for this resource. If not provided, the provider's default API key will be used",
			},
			"service_account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the service account",
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
			"api_key_value": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The value of the API key associated with the service account (only available upon creation)",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

// resourceOpenAIProjectServiceAccountCreate creates a new OpenAI project service account
func resourceOpenAIProjectServiceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics

	// Get values from schema
	projectID := d.Get("project_id").(string)
	name := d.Get("name").(string)
	apiKey := d.Get("api_key").(string)

	// Create the service account
	tflog.Debug(ctx, "Creating OpenAI project service account", map[string]interface{}{
		"project_id": projectID,
		"name":       name,
	})

	serviceAccount, err := c.CreateProjectServiceAccount(projectID, name, apiKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating project service account: %w", err))
	}

	// Set ID to {project_id}:{service_account_id}
	d.SetId(fmt.Sprintf("%s:%s", projectID, serviceAccount.ID))

	// Set computed fields
	if err := d.Set("service_account_id", serviceAccount.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", serviceAccount.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
	}
	if err := d.Set("role", serviceAccount.Role); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role: %v", err))
	}

	// Set API key information if available
	if serviceAccount.APIKey != nil {
		if err := d.Set("api_key_id", serviceAccount.APIKey.ID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set api_key_id: %v", err))
		}
		if serviceAccount.APIKey.Value != "" {
			if err := d.Set("api_key_value", serviceAccount.APIKey.Value); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set api_key_value: %v", err))
			}
		}
	}

	return diags
}

// resourceOpenAIProjectServiceAccountRead reads an existing OpenAI project service account
func resourceOpenAIProjectServiceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics

	// Parse ID
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("invalid ID format. Expected 'project_id:service_account_id', got '%s'", d.Id()))
	}

	projectID := idParts[0]
	serviceAccountID := idParts[1]
	apiKey := d.Get("api_key").(string)

	// Get the service account
	tflog.Debug(ctx, "Reading OpenAI project service account", map[string]interface{}{
		"project_id":         projectID,
		"service_account_id": serviceAccountID,
	})

	serviceAccount, err := c.GetProjectServiceAccount(projectID, serviceAccountID, apiKey)
	if err != nil {
		// Check if the error is a 404, which means the service account is gone
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			tflog.Info(ctx, "Service account not found, removing from state", map[string]interface{}{
				"project_id":         projectID,
				"service_account_id": serviceAccountID,
			})
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading project service account: %w", err))
	}

	// Set project ID in state
	if err := d.Set("project_id", projectID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set project_id: %v", err))
	}
	if err := d.Set("name", serviceAccount.Name); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set name: %v", err))
	}
	if err := d.Set("service_account_id", serviceAccount.ID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set service_account_id: %v", err))
	}
	if err := d.Set("created_at", serviceAccount.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
	}
	if err := d.Set("role", serviceAccount.Role); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role: %v", err))
	}

	// API key value won't be available during read, only during create
	// But we can still set the API key ID if it's available
	if serviceAccount.APIKey != nil {
		if err := d.Set("api_key_id", serviceAccount.APIKey.ID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set api_key_id: %v", err))
		}
	}

	return diags
}

// resourceOpenAIProjectServiceAccountDelete deletes an OpenAI project service account
func resourceOpenAIProjectServiceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics

	// Parse ID
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("invalid ID format. Expected 'project_id:service_account_id', got '%s'", d.Id()))
	}

	projectID := idParts[0]
	serviceAccountID := idParts[1]
	apiKey := d.Get("api_key").(string)

	// Delete the service account
	tflog.Debug(ctx, "Deleting OpenAI project service account", map[string]interface{}{
		"project_id":         projectID,
		"service_account_id": serviceAccountID,
	})

	err = c.DeleteProjectServiceAccount(projectID, serviceAccountID, apiKey)
	if err != nil {
		// If the service account is already gone, just log a warning
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			tflog.Warn(ctx, "Service account was already deleted", map[string]interface{}{
				"project_id":         projectID,
				"service_account_id": serviceAccountID,
			})
		} else {
			return diag.FromErr(fmt.Errorf("error deleting project service account: %w", err))
		}
	}

	// Remove from state
	d.SetId("")

	return diags
}
