package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// AdminAPIKeyResponse represents the API response for an OpenAI admin API key
type AdminAPIKeyResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	CreatedAt int64    `json:"created_at"`
	ExpiresAt *int64   `json:"expires_at,omitempty"`
	Object    string   `json:"object"`
	Scopes    []string `json:"scopes,omitempty"`
	Key       string   `json:"key"`
}

// resourceOpenAIAdminAPIKey returns the schema and CRUD operations for the OpenAI Admin API Key resource
func resourceOpenAIAdminAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIAdminAPIKeyCreate,
		ReadContext:   resourceOpenAIAdminAPIKeyRead,
		DeleteContext: resourceOpenAIAdminAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the API key",
			},
			"expires_at": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Unix timestamp when the API key should expire (optional)",
			},
			"scopes": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Scopes to assign to the API key (e.g., 'api.management.read', 'api.management.write')",
			},
			// Computed fields
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the API key was created",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type",
			},
			"api_key_value": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The actual API key value (only available upon creation)",
			},
		},
	}
}

// resourceOpenAIAdminAPIKeyCreate creates a new OpenAI admin API key
func resourceOpenAIAdminAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := GetOpenAIClientWithAdminKey(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the name from the resource data
	name := d.Get("name").(string)
	if name == "" {
		return diag.Errorf("name cannot be empty")
	}

	// Prepare optional parameters
	var expiresAt *int64
	if v, ok := d.GetOk("expires_at"); ok {
		t := int64(v.(int))
		expiresAt = &t
	}

	var scopes []string
	if v, ok := d.GetOk("scopes"); ok {
		scopesList := v.([]interface{})
		scopes = make([]string, len(scopesList))
		for i, scope := range scopesList {
			scopes[i] = scope.(string)
		}
	}

	// Create the API key using the provider's API key
	resp, err := client.CreateAPIKey(name, expiresAt, scopes)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating API key: %v", err))
	}

	// Convert from client.AdminAPIKeyResponse to local AdminAPIKeyResponse
	apiKey := &AdminAPIKeyResponse{
		ID:        resp.ID,
		Name:      resp.Name,
		CreatedAt: resp.CreatedAt,
		ExpiresAt: resp.ExpiresAt,
		Object:    resp.Object,
		Scopes:    resp.Scopes,
		Key:       resp.Key,
	}

	// Set the resource ID to the API key ID
	d.SetId(apiKey.ID)

	// Set the computed values in the state
	if err := d.Set("name", apiKey.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", time.Unix(apiKey.CreatedAt, 0).Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("object", apiKey.Object); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("api_key_value", apiKey.Key); err != nil {
		return diag.FromErr(err)
	}

	// For scopes, we need to handle possibly nil scopes
	if apiKey.Scopes != nil {
		if err := d.Set("scopes", apiKey.Scopes); err != nil {
			return diag.FromErr(err)
		}
	}

	// For expires_at, we only set it if it's provided by the API
	if apiKey.ExpiresAt != nil {
		if err := d.Set("expires_at", *apiKey.ExpiresAt); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIAdminAPIKeyRead reads an existing OpenAI admin API key
func resourceOpenAIAdminAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := GetOpenAIClientWithAdminKey(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the ID from the resource
	id := d.Id()
	if id == "" {
		return diag.Errorf("API key ID is empty")
	}

	// Get the API key
	apiKey, err := client.GetAPIKey(id)
	if err != nil {
		// Check if the error suggests the API key doesn't exist
		if isResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading API key: %v", err))
	}

	// Update the resource data from the API response
	if err := d.Set("name", apiKey.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", time.Unix(apiKey.CreatedAt, 0).Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("object", apiKey.Object); err != nil {
		return diag.FromErr(err)
	}

	// For scopes, we need to handle possibly nil scopes
	if apiKey.Scopes != nil {
		if err := d.Set("scopes", apiKey.Scopes); err != nil {
			return diag.FromErr(err)
		}
	}

	// For expires_at, we only set it if it's provided by the API
	if apiKey.ExpiresAt != nil {
		if err := d.Set("expires_at", *apiKey.ExpiresAt); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIAdminAPIKeyDelete deletes an OpenAI admin API key
func resourceOpenAIAdminAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := GetOpenAIClientWithAdminKey(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the ID from the resource
	id := d.Id()
	if id == "" {
		return diag.Errorf("API key ID is empty")
	}

	// Delete the API key
	err = client.DeleteAPIKey(id)
	if err != nil {
		// Don't fail if the API key is already gone
		if isResourceNotFoundError(err) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting API key: %v", err))
	}

	// Clear the ID to indicate the resource no longer exists
	d.SetId("")

	return diag.Diagnostics{}
}

// isResourceNotFoundError checks if an error indicates the resource was not found
func isResourceNotFoundError(err error) bool {
	return err != nil && (contains(err.Error(), "404") ||
		contains(err.Error(), "not found") ||
		contains(err.Error(), "doesn't exist"))
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s[0:len(substr)] == substr
}
