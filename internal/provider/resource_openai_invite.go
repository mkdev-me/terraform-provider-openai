package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fjcorp/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				Optional:    true,
				ForceNew:    true,
				Description: "Projects to assign to the invited user (Note: User will be assigned to projects after invitation is accepted)",
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
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "API key for authentication. If not provided, the provider's default API key will be used.",
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
			"user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the user (available after user exists in organization)",
			},
		},
	}
}

func resourceOpenAIInviteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	email := d.Get("email").(string)
	role := d.Get("role").(string)

	// Get custom API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Process project assignments if present
	var projectAssignments []map[string]interface{}
	if projectsRaw, ok := d.GetOk("projects"); ok {
		projectsList := projectsRaw.([]interface{})
		projectAssignments = make([]map[string]interface{}, 0, len(projectsList))

		for _, projectRaw := range projectsList {
			projectMap := projectRaw.(map[string]interface{})
			projectAssignments = append(projectAssignments, map[string]interface{}{
				"id":   projectMap["id"].(string),
				"role": projectMap["role"].(string),
			})
		}
	}

	// Step 1: Try to create the invitation without projects
	// The OpenAI API doesn't support project assignments during invitation
	tflog.Info(ctx, fmt.Sprintf("Creating invitation for email: %s with organization role: %s", email, role))

	invite, err := c.CreateInvite(email, role, []client.InviteProject{}, apiKey)
	if err != nil {
		// Check if the user already exists
		if strings.Contains(err.Error(), "already exists in this organization") {
			tflog.Info(ctx, fmt.Sprintf("User %s already exists in organization, proceeding with project assignments", email))

			// Generate a synthetic invite ID for existing users
			d.SetId(fmt.Sprintf("existing-user-%s", email))
			d.Set("invite_id", d.Id())
			d.Set("status", "existing_user")
			d.Set("created_at", time.Now().Unix())
			d.Set("expires_at", time.Now().Add(24*time.Hour).Unix())

			// Proceed to assign projects
		} else {
			return diag.FromErr(fmt.Errorf("error creating invitation: %s", err))
		}
	} else {
		// Set the resource ID and computed fields for new invitation
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
	}

	// Step 2: If projects were specified, assign the user to projects
	if len(projectAssignments) > 0 {
		tflog.Info(ctx, fmt.Sprintf("Assigning user %s to %d projects", email, len(projectAssignments)))

		// Wait for user to be available in the organization (with retry logic)
		var userID string
		err = resource.Retry(30*time.Second, func() *resource.RetryError {
			user, exists, err := c.FindUserByEmail(email, apiKey)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("error finding user by email: %s", err))
			}
			if !exists {
				tflog.Debug(ctx, fmt.Sprintf("User %s not yet found in organization, retrying...", email))
				return resource.RetryableError(fmt.Errorf("user not yet found in organization"))
			}
			userID = user.ID
			return nil
		})

		if err != nil {
			// If we can't find the user, log a warning but don't fail
			tflog.Warn(ctx, fmt.Sprintf("Could not find user %s to assign to projects: %s. User may need to accept invitation first.", email, err))
		} else {
			// Set the user ID
			d.Set("user_id", userID)

			// Assign user to each project
			for _, projectMap := range projectAssignments {
				projectID := projectMap["id"].(string)
				projectRole := projectMap["role"].(string)

				tflog.Info(ctx, fmt.Sprintf("Assigning user %s to project %s with role %s", userID, projectID, projectRole))

				_, err := c.AddProjectUser(projectID, userID, projectRole, apiKey)
				if err != nil {
					// Check if user is already in project
					if strings.Contains(err.Error(), "already exists in project") {
						tflog.Info(ctx, fmt.Sprintf("User %s already exists in project %s, updating role if needed", userID, projectID))

						// Try to update the role
						_, updateErr := c.UpdateProjectUser(projectID, userID, projectRole, apiKey)
						if updateErr != nil {
							tflog.Warn(ctx, fmt.Sprintf("Could not update user role in project %s: %s", projectID, updateErr))
						}
					} else {
						tflog.Warn(ctx, fmt.Sprintf("Could not add user to project %s: %s", projectID, err))
					}
				}
			}
		}
	}

	return resourceOpenAIInviteRead(ctx, d, m)
}

func resourceOpenAIInviteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	inviteID := d.Id()

	// Get custom API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Handle synthetic IDs for existing users
	if strings.HasPrefix(inviteID, "existing-user-") {
		// For existing users, check if they still exist
		email := d.Get("email").(string)
		user, exists, err := c.FindUserByEmail(email, apiKey)
		if err != nil || !exists {
			// User no longer exists, remove from state
			d.SetId("")
			return nil
		}
		// Update user ID if we have it
		d.Set("user_id", user.ID)
		return nil
	}

	// Retrieve the invitation
	invite, err := c.GetInvite(inviteID, apiKey)
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

	// Try to get user ID if invitation was accepted
	if invite.Status == "accepted" {
		user, exists, _ := c.FindUserByEmail(invite.Email, apiKey)
		if exists {
			d.Set("user_id", user.ID)
		}
	}

	return nil
}

func resourceOpenAIInviteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	inviteID := d.Id()

	// Get custom API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Handle synthetic IDs for existing users
	if strings.HasPrefix(inviteID, "existing-user-") {
		// For existing users, we don't delete anything, just remove from state
		d.SetId("")
		return nil
	}

	// Delete the invitation
	err = c.DeleteInvite(inviteID, apiKey)
	if err != nil {
		// If the error is that the invitation is already accepted, that's OK
		if strings.Contains(err.Error(), "already accepted") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting invitation: %s", err))
	}

	// Remove resource from state
	d.SetId("")
	return nil
}

func resourceOpenAIInviteImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	inviteID := d.Id()

	c, err := GetOpenAIClient(m)
	if err != nil {
		return nil, err
	}

	// Use the provider's configured API key via the DoRequest method, which will use the default key
	invite, err := c.GetInvite(inviteID, "")
	if err != nil {
		return nil, fmt.Errorf("error retrieving invite for import: %s", err)
	}

	d.SetId(invite.ID)
	d.Set("email", invite.Email)
	d.Set("role", invite.Role)
	d.Set("invite_id", invite.ID)
	d.Set("status", invite.Status)
	d.Set("created_at", invite.CreatedAt)
	d.Set("expires_at", invite.ExpiresAt)

	if len(invite.Projects) > 0 {
		projects := make([]map[string]interface{}, len(invite.Projects))
		for i, project := range invite.Projects {
			projects[i] = map[string]interface{}{
				"id":   project.ID,
				"role": project.Role,
			}
		}
		d.Set("projects", projects)
	}

	return []*schema.ResourceData{d}, nil
}
