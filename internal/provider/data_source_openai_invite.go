package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIInvite returns a schema.Resource that represents a data source for a single OpenAI invite.
// This data source allows users to retrieve information about a specific invitation in an OpenAI organization.
func dataSourceOpenAIInvite() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIInviteRead,
		Schema: map[string]*schema.Schema{
			"invite_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the invitation to retrieve",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "API key for authentication. If not provided, the provider's default API key will be used.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address of the invited user",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role assigned to the invited user (owner or reader)",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the invitation",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "When the invitation was created",
			},
			"expires_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "When the invitation expires",
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
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role assigned to the user within the project (owner or member)",
						},
					},
				},
				Description: "Projects assigned to the invited user",
			},
		},
	}
}

// dataSourceOpenAIInviteRead handles the read operation for the OpenAI invite data source.
// It retrieves information about a specific invitation from the OpenAI API.
func dataSourceOpenAIInviteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	inviteID := d.Get("invite_id").(string)
	if inviteID == "" {
		return diag.FromErr(fmt.Errorf("invite_id is required"))
	}

	// Get custom API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Retrieve the invitation
	invite, err := c.GetInvite(inviteID, apiKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving invitation: %s", err))
	}

	// Set the resource ID
	d.SetId(invite.ID)

	// Update the state with the invitation details
	if err := d.Set("email", invite.Email); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email: %s", err))
	}
	if err := d.Set("role", invite.Role); err != nil {
		return diag.FromErr(fmt.Errorf("error setting role: %s", err))
	}
	if err := d.Set("status", invite.Status); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %s", err))
	}
	if err := d.Set("created_at", invite.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created_at: %s", err))
	}
	if err := d.Set("expires_at", invite.ExpiresAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting expires_at: %s", err))
	}

	// Set project assignments if present
	if len(invite.Projects) > 0 {
		projects := make([]map[string]interface{}, len(invite.Projects))
		for i, project := range invite.Projects {
			projects[i] = map[string]interface{}{
				"id":   project.ID,
				"role": project.Role,
			}
		}
		if err := d.Set("projects", projects); err != nil {
			return diag.FromErr(fmt.Errorf("error setting projects: %s", err))
		}
	}

	return nil
}
