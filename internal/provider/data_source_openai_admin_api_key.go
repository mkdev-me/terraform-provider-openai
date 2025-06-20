package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// AdminAPIKey represents the API response for getting an OpenAI admin API key
type AdminAPIKey struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	CreatedAt int64    `json:"created_at"`
	ExpiresAt *int64   `json:"expires_at,omitempty"`
	Object    string   `json:"object"`
	Scopes    []string `json:"scopes,omitempty"`
}

// dataSourceOpenAIAdminAPIKey returns a schema.Resource that represents a data source for an OpenAI admin API key.
// This data source allows users to retrieve information about a specific admin API key.
func dataSourceOpenAIAdminAPIKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIAdminAPIKeyRead,
		Schema: map[string]*schema.Schema{
			"api_key_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the admin API key to retrieve",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the admin API key",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the admin API key was created",
			},
			"expires_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Timestamp when the admin API key expires (optional)",
			},
			"scopes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Scopes assigned to the admin API key",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type",
			},
		},
	}
}

// dataSourceOpenAIAdminAPIKeyRead handles the read operation for the OpenAI admin API key data source.
// It retrieves information about a specific admin API key from the OpenAI API.
func dataSourceOpenAIAdminAPIKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	apiKeyID := d.Get("api_key_id").(string)
	if apiKeyID == "" {
		return diag.Errorf("api_key_id cannot be empty")
	}

	// Set the ID to the API key ID
	d.SetId(apiKeyID)

	// Use the provider's API key
	clientAPIKey, err := client.GetAPIKey(apiKeyID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving admin API key: %s", err))
	}

	// Convert client.AdminAPIKey to our AdminAPIKey
	apiKey := &AdminAPIKey{
		ID:        clientAPIKey.ID,
		Name:      clientAPIKey.Name,
		CreatedAt: clientAPIKey.CreatedAt,
		ExpiresAt: clientAPIKey.ExpiresAt,
		Object:    clientAPIKey.Object,
		Scopes:    clientAPIKey.Scopes,
	}

	if err := setAdminAPIKeyAttributes(d, apiKey); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// setAdminAPIKeyAttributes sets the attributes of an admin API key in the resource data
func setAdminAPIKeyAttributes(d *schema.ResourceData, apiKey *AdminAPIKey) error {
	if err := d.Set("name", apiKey.Name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}

	if err := d.Set("created_at", time.Unix(apiKey.CreatedAt, 0).Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting created_at: %s", err)
	}

	if err := d.Set("object", apiKey.Object); err != nil {
		return fmt.Errorf("error setting object: %s", err)
	}

	if apiKey.ExpiresAt != nil {
		if err := d.Set("expires_at", *apiKey.ExpiresAt); err != nil {
			return fmt.Errorf("error setting expires_at: %s", err)
		}
	}

	if apiKey.Scopes != nil {
		if err := d.Set("scopes", apiKey.Scopes); err != nil {
			return fmt.Errorf("error setting scopes: %s", err)
		}
	}

	return nil
}
