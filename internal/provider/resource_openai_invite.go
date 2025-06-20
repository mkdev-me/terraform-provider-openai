package provider

import (
	"context"
	"fmt"

	"github.com/fjcorp/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceOpenAIInvite() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIInviteCreate,
		ReadContext:   resourceOpenAIInviteRead,
		DeleteContext: resourceOpenAIInviteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIInviteImport,
		},
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The email address of the user to invite",
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"owner", "reader"}, false),
				Description:  "The role to assign to the user (owner or reader)",
			},
			"projects": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "Projects to assign to the invited user",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the project",
						},
						"role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"owner", "member"}, false),
							Description:  "The role to assign to the user within the project (owner or member)",
						},
					},
				},
				MinItems: 1,
			},
			"invite_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the invitation",
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
		},
	}
}

func resourceOpenAIInviteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	email := d.Get("email").(string)
	role := d.Get("role").(string)

	// Process project assignments if present
	var projects []client.InviteProject
	if projectsRaw, ok := d.GetOk("projects"); ok {
		projectsList := projectsRaw.([]interface{})
		projects = make([]client.InviteProject, 0, len(projectsList))

		for _, projectRaw := range projectsList {
			projectMap := projectRaw.(map[string]interface{})
			project := client.InviteProject{
				ID:   projectMap["id"].(string),
				Role: projectMap["role"].(string),
			}
			projects = append(projects, project)
		}
	}

	// Create the invitation
	invite, err := c.CreateInvite(email, role, projects)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating invitation: %s", err))
	}

	// Set the resource ID and computed fields
	d.SetId(invite.ID)
	if err := d.Set("invite_id", invite.ID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting invite_id: %s", err))
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

	return resourceOpenAIInviteRead(ctx, d, m)
}

func resourceOpenAIInviteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	inviteID := d.Id()

	// Retrieve the invitation
	invite, err := c.GetInvite(inviteID)
	if err != nil {
		// If the invite doesn't exist anymore, remove it from state
		d.SetId("")
		return nil
	}

	// Update computed fields
	if err := d.Set("email", invite.Email); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email: %s", err))
	}
	if err := d.Set("role", invite.Role); err != nil {
		return diag.FromErr(fmt.Errorf("error setting role: %s", err))
	}
	if err := d.Set("invite_id", invite.ID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting invite_id: %s", err))
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

	// Update projects if present
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

func resourceOpenAIInviteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	inviteID := d.Id()

	// Delete the invitation
	err = c.DeleteInvite(inviteID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting invitation: %s", err))
	}

	// Remove resource from state
	d.SetId("")
	return nil
}

func resourceOpenAIInviteImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	inviteID := d.Id()

	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return nil, err
	}

	// Use the provider's configured API key via the DoRequest method, which will use the default key
	invite, err := c.GetInvite(inviteID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving invite for import: %s", err)
	}

	d.SetId(invite.ID)
	_ = d.Set("email", invite.Email)
	_ = d.Set("role", invite.Role)
	_ = d.Set("invite_id", invite.ID)
	_ = d.Set("status", invite.Status)
	_ = d.Set("created_at", invite.CreatedAt)
	_ = d.Set("expires_at", invite.ExpiresAt)

	if len(invite.Projects) > 0 {
		projects := make([]map[string]interface{}, len(invite.Projects))
		for i, project := range invite.Projects {
			projects[i] = map[string]interface{}{
				"id":   project.ID,
				"role": project.Role,
			}
		}
		_ = d.Set("projects", projects)
	}

	return []*schema.ResourceData{d}, nil
}
