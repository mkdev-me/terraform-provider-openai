package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// AdminAPIKeyWithLastUsed extends AdminAPIKeyResponse to include last_used_at field
type AdminAPIKeyWithLastUsed struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	CreatedAt  int64    `json:"created_at"`
	ExpiresAt  *int64   `json:"expires_at,omitempty"`
	LastUsedAt *int64   `json:"last_used_at,omitempty"`
	Object     string   `json:"object"`
	Scopes     []string `json:"scopes,omitempty"`
	Key        string   `json:"key,omitempty"`
}

// ListAPIKeysResponse represents the API response for listing admin API keys
type ListAPIKeysResponse struct {
	Data    []AdminAPIKeyWithLastUsed `json:"data"`
	HasMore bool                      `json:"has_more"`
	Object  string                    `json:"object"`
}

// dataSourceOpenAIAdminAPIKeys returns a schema.Resource that represents a data source for OpenAI admin API keys.
// This data source allows users to retrieve a list of all admin API keys.
func dataSourceOpenAIAdminAPIKeys() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIAdminAPIKeysRead,
		Schema: map[string]*schema.Schema{
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "Maximum number of API keys to return",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Cursor for pagination, API key ID to fetch results after",
			},
			"api_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the admin API key",
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
						"last_used_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the admin API key was last used",
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
				},
				Description: "List of admin API keys",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more API keys available beyond the limit",
			},
		},
	}
}

// dataSourceOpenAIAdminAPIKeysRead handles the read operation for the OpenAI admin API keys data source.
// It retrieves a list of all admin API keys from the OpenAI API.
func dataSourceOpenAIAdminAPIKeysRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get pagination parameters
	limit := d.Get("limit").(int)
	after := d.Get("after").(string)

	// Set a unique ID for the resource
	d.SetId(fmt.Sprintf("admin_api_keys_%d", time.Now().Unix()))

	// Use the provider's API key
	clientResponse, err := client.ListAPIKeys(limit, after)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing admin API keys: %s", err))
	}

	// Convert client response to our ListAPIKeysResponse
	response := &ListAPIKeysResponse{
		HasMore: clientResponse.HasMore,
		Object:  clientResponse.Object,
	}

	// Convert each API key in the response
	response.Data = make([]AdminAPIKeyWithLastUsed, len(clientResponse.Data))
	for i, key := range clientResponse.Data {
		response.Data[i] = AdminAPIKeyWithLastUsed{
			ID:         key.ID,
			Name:       key.Name,
			CreatedAt:  key.CreatedAt,
			ExpiresAt:  key.ExpiresAt,
			Object:     key.Object,
			Scopes:     key.Scopes,
			LastUsedAt: key.LastUsedAt,
		}
	}

	// Convert the response to the format expected by the schema
	apiKeys := make([]map[string]interface{}, len(response.Data))
	for i, key := range response.Data {
		apiKey := map[string]interface{}{
			"id":         key.ID,
			"name":       key.Name,
			"object":     key.Object,
			"created_at": time.Unix(key.CreatedAt, 0).Format(time.RFC3339),
		}

		// Handle optional fields
		if key.ExpiresAt != nil {
			apiKey["expires_at"] = *key.ExpiresAt
		}

		if key.LastUsedAt != nil {
			apiKey["last_used_at"] = time.Unix(*key.LastUsedAt, 0).Format(time.RFC3339)
		}

		if key.Scopes != nil {
			apiKey["scopes"] = key.Scopes
		}

		apiKeys[i] = apiKey
	}

	if err := d.Set("api_keys", apiKeys); err != nil {
		return diag.FromErr(fmt.Errorf("error setting api_keys: %s", err))
	}

	if err := d.Set("has_more", response.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %s", err))
	}

	return diag.Diagnostics{}
}
