package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// resourceOpenAIProjectUser defines the schema and CRUD operations for OpenAI project users.
// This resource allows users to manage project users through Terraform,
// including adding, reading, updating, and removing users from projects.
func resourceOpenAIProjectUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIProjectUserCreate,
		ReadContext:   resourceOpenAIProjectUserRead,
		UpdateContext: resourceOpenAIProjectUserUpdate,
		DeleteContext: resourceOpenAIProjectUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIProjectUserImport,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the project the user will be added to",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the user to add to the project",
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"owner", "member"}, false),
				Description:  "The role to assign to the user (owner or member)",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "API key specific to this project. If not provided, the provider's default API key will be used.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Always suppress the diff for the API key
					return true
				},
				// This ensures the API key never gets stored in the state file
				StateFunc: func(val interface{}) string {
					// Return empty string instead of the actual API key
					return ""
				},
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address of the user",
			},
			"added_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Timestamp when the user was added to the project",
			},
		},
	}
}

// resourceOpenAIProjectUserCreate adds a user to an OpenAI project.
// It requires the project_id, user_id, and role to be specified.
func resourceOpenAIProjectUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	userID := d.Get("user_id").(string)
	role := d.Get("role").(string)

	// Generate a unique ID for the resource
	id := fmt.Sprintf("%s:%s", projectID, userID)
	d.SetId(id)

	// Use the project-specific API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
		tflog.Debug(ctx, "Using custom API key for adding user to project")
	}

	// First check if the user is already in the project
	tflog.Debug(ctx, fmt.Sprintf("Checking if user %s already exists in project %s", userID, projectID))
	existingUser, exists, err := c.FindProjectUser(projectID, userID, apiKey)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error checking if user exists: %v", err))
		return diag.Errorf("Error checking if user exists in project: %s", err)
	}

	if exists {
		tflog.Info(ctx, fmt.Sprintf("User %s already exists in project %s, using existing user", userID, projectID))
		// Update the Terraform state with values from the existing user
		if err := d.Set("email", existingUser.Email); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set email: %v", err))
		}
		if err := d.Set("added_at", existingUser.AddedAt); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set added_at: %v", err))
		}
		if err := d.Set("role", existingUser.Role); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set role: %v", err))
		}

		// If the role in the configuration is different from the existing role, update it
		if existingUser.Role != role {
			tflog.Info(ctx, fmt.Sprintf("User's current role '%s' differs from desired role '%s', updating...", existingUser.Role, role))
			return resourceOpenAIProjectUserUpdate(ctx, d, m)
		}

		return diag.Diagnostics{}
	}

	// Add the user to the project
	tflog.Debug(ctx, fmt.Sprintf("Adding user %s to project %s with role %s", userID, projectID, role))
	projectUser, err := c.AddProjectUser(projectID, userID, role, apiKey)
	if err != nil {
		// Check if error is because user already exists
		if strings.Contains(err.Error(), "already exists in project") {
			tflog.Info(ctx, fmt.Sprintf("User %s already exists in project, continuing: %s", userID, err))

			// Try to get user details again
			existingUser, exists, findErr := c.FindProjectUser(projectID, userID, apiKey)
			if findErr == nil && exists {
				// Update the Terraform state with values from the existing user
				if err := d.Set("email", existingUser.Email); err != nil {
					return diag.FromErr(fmt.Errorf("failed to set email: %v", err))
				}
				if err := d.Set("added_at", existingUser.AddedAt); err != nil {
					return diag.FromErr(fmt.Errorf("failed to set added_at: %v", err))
				}
				if err := d.Set("role", existingUser.Role); err != nil {
					return diag.FromErr(fmt.Errorf("failed to set role: %v", err))
				}

				// If the role in the configuration is different from the existing role, update it
				if existingUser.Role != role {
					tflog.Info(ctx, fmt.Sprintf("User's current role '%s' differs from desired role '%s', updating...", existingUser.Role, role))
					return resourceOpenAIProjectUserUpdate(ctx, d, m)
				}

				return diag.Diagnostics{}
			}

			// If we can't get the user details, just continue with the state as is
			return diag.Diagnostics{}
		}

		// For other errors, return an error diagnostic
		tflog.Error(ctx, fmt.Sprintf("Failed to add user to project: %v", err))
		return diag.Errorf("Error adding user to project: %s", err)
	}

	// Update the Terraform state with computed values
	if err := d.Set("email", projectUser.Email); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set email: %v", err))
	}
	if err := d.Set("added_at", projectUser.AddedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set added_at: %v", err))
	}

	// Explicitly set the api_key to empty in the state
	if err := d.Set("api_key", ""); err != nil {
		return diag.FromErr(fmt.Errorf("failed to reset api_key: %v", err))
	}

	return diag.Diagnostics{}
}

// resourceOpenAIProjectUserRead retrieves information about a user in a project.
// This implementation now tries to verify if the user exists in the project.
func resourceOpenAIProjectUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	userID := d.Get("user_id").(string)

	// Use the project-specific API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
		tflog.Debug(ctx, "Using custom API key for reading user from project")
	}

	// Check if the user exists in the project
	tflog.Debug(ctx, fmt.Sprintf("Checking if user %s exists in project %s", userID, projectID))
	existingUser, exists, err := c.FindProjectUser(projectID, userID, apiKey)
	if err != nil {
		// If it's a permissions error, just keep the state
		if strings.Contains(err.Error(), "insufficient permissions") {
			tflog.Warn(ctx, fmt.Sprintf("Permission error reading user from project, using local state: %s", err))

			// Even in case of error, ensure API key is not stored
			if err := d.Set("api_key", ""); err != nil {
				return diag.FromErr(fmt.Errorf("failed to reset api_key: %v", err))
			}

			return diag.Diagnostics{}
		}

		tflog.Error(ctx, fmt.Sprintf("Error checking if user exists: %v", err))
		return diag.Errorf("Error checking if user exists in project: %s", err)
	}

	if !exists {
		// User doesn't exist in the project, remove from state
		tflog.Warn(ctx, fmt.Sprintf("User %s no longer exists in project %s, removing from state", userID, projectID))
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Update state with non-role values from API
	if err := d.Set("email", existingUser.Email); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set email: %v", err))
	}
	if err := d.Set("added_at", existingUser.AddedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set added_at: %v", err))
	}

	// Get the configured role (what's in .tf file) and the role from API
	configuredRole := d.Get("role").(string)

	// Log if there's a mismatch but DO NOT update the role in the state
	// This ensures the configuration remains the source of truth
	if configuredRole != existingUser.Role {
		tflog.Warn(ctx, fmt.Sprintf("Role mismatch: terraform=%s, API=%s for user %s in project %s. "+
			"Terraform configuration will take precedence.",
			configuredRole, existingUser.Role, userID, projectID))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("Role in sync: terraform=%s, API=%s for user %s in project %s",
			configuredRole, existingUser.Role, userID, projectID))
	}

	// IMPORTANT: We deliberately DO NOT update the role from the API
	// This ensures that the Terraform config remains the source of truth
	// If you want to synchronize with the API state, use terraform import

	// Explicitly set the api_key to empty in the state
	if err := d.Set("api_key", ""); err != nil {
		return diag.FromErr(fmt.Errorf("failed to reset api_key: %v", err))
	}

	return diag.Diagnostics{}
}

// resourceOpenAIProjectUserUpdate updates a user's role in a project.
// Only the role can be updated, as other attributes like user_id and project_id require recreating the resource.
func resourceOpenAIProjectUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if the role has changed
	if !d.HasChange("role") {
		return diag.Diagnostics{}
	}

	projectID := d.Get("project_id").(string)
	userID := d.Get("user_id").(string)
	role := d.Get("role").(string)

	// Use the project-specific API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
		tflog.Debug(ctx, "Using custom API key for updating user in project")
	}

	// First check: Is this user an organization owner?
	// Get information about the user in the organization first
	// This is a preventive check, as attempting to change the role of an org owner will fail
	orgUser, exists, err := c.GetUser(userID, apiKey)
	if err == nil && exists && orgUser.Role == "owner" {
		// This is an organization owner, they can't have their project role changed from owner
		tflog.Error(ctx, fmt.Sprintf("Cannot change role for user %s (email: %s) because they are an organization owner. "+
			"Organization owners must maintain 'owner' role in all projects.",
			userID, orgUser.Email))

		// Return detailed diagnostic explaining the issue
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Cannot change role for organization owner",
				Detail:   fmt.Sprintf("User %s (email: %s) is an organization owner. Organization owners must maintain 'owner' role in all projects. The OpenAI API does not allow changing their role to 'member'.", userID, orgUser.Email),
			},
		}
	}

	// Update the user's role in the project
	tflog.Info(ctx, fmt.Sprintf("Updating user %s in project %s to role %s via API call", userID, projectID, role))

	// Call the API to update the user's role
	updatedUser, err := c.UpdateProjectUser(projectID, userID, role, apiKey)
	if err != nil {
		// Handle specific error for organization owners
		if strings.Contains(err.Error(), "owner of the organization") ||
			strings.Contains(err.Error(), "organization owner") {
			tflog.Error(ctx, fmt.Sprintf("Cannot update user %s as they are an organization owner: %s", userID, err))

			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Cannot change role for organization owner",
					Detail:   fmt.Sprintf("User %s is an organization owner. Organization owners must maintain 'owner' role in all projects. The OpenAI API does not allow changing their role to 'member'. Error from API: %s", userID, err),
				},
			}
		}

		tflog.Error(ctx, fmt.Sprintf("Error updating user role: %v", err))
		return diag.Errorf("Error updating user role in project: %s", err)
	}

	// Verify the update was successful by checking the returned role
	if updatedUser.Role != role {
		tflog.Warn(ctx, fmt.Sprintf("API returned role %s after update, which doesn't match requested role %s. "+
			"This likely means the role change was not applied correctly.",
			updatedUser.Role, role))

		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Role change was not applied",
				Detail:   fmt.Sprintf("The API returned role '%s' after attempting to change to '%s'. This typically happens when trying to change the role of an organization owner or due to permission issues.", updatedUser.Role, role),
			},
		}
	} else {
		tflog.Info(ctx, fmt.Sprintf("Successfully updated user %s to role %s in project %s",
			userID, role, projectID))
	}

	// Update the Terraform state with the updated user info
	if err := d.Set("role", updatedUser.Role); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role: %v", err))
	}
	if err := d.Set("email", updatedUser.Email); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set email: %v", err))
	}
	if err := d.Set("added_at", updatedUser.AddedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set added_at: %v", err))
	}

	// Explicitly set the api_key to empty in the state
	if err := d.Set("api_key", ""); err != nil {
		return diag.FromErr(fmt.Errorf("failed to reset api_key: %v", err))
	}

	return diag.Diagnostics{}
}

// resourceOpenAIProjectUserDelete removes a user from a project.
// It now makes an actual API call to remove the user, with appropriate error handling.
func resourceOpenAIProjectUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	userID := d.Get("user_id").(string)

	// Use the project-specific API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
		tflog.Debug(ctx, "Using custom API key for removing user from project")
	}

	// Attempt to remove the user from the project
	tflog.Debug(ctx, fmt.Sprintf("Removing user %s from project %s", userID, projectID))
	err = c.RemoveProjectUser(projectID, userID, apiKey)
	if err != nil {
		// Check if this is a case where the user is an organization owner
		if strings.Contains(err.Error(), "owner of the organization") {
			tflog.Warn(ctx, fmt.Sprintf("Cannot remove user %s as they are an organization owner: %s", userID, err))

			// Since we can't actually remove the user, just remove from Terraform state
			tflog.Info(ctx, "Removing user from Terraform state only, since they cannot be removed from the project")
			d.SetId("")
			return diag.Diagnostics{}
		}

		// If the error indicates the user doesn't exist, consider the delete successful
		if strings.Contains(err.Error(), "not found") {
			tflog.Info(ctx, fmt.Sprintf("User %s not found in project %s, considering delete successful", userID, projectID))
			d.SetId("")
			return diag.Diagnostics{}
		}

		// For any other error, return an error diagnostic
		tflog.Error(ctx, fmt.Sprintf("Error removing user from project: %v", err))
		return diag.Errorf("Error removing user from project: %s", err)
	}

	// Remove from Terraform state
	d.SetId("")

	// Even though we're removing the resource, ensure the API key is cleared from state
	if err := d.Set("api_key", ""); err != nil {
		return diag.FromErr(fmt.Errorf("failed to reset api_key: %v", err))
	}

	return diag.Diagnostics{}
}

// resourceOpenAIProjectUserImport imports an existing project user into Terraform state.
// The ID should be in the format "project_id:user_id".
func resourceOpenAIProjectUserImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return nil, err
	}

	// Split the ID to get project_id and user_id
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ID format for project user import. Expected 'project_id:user_id', got: %s", d.Id())
	}

	projectID := parts[0]
	userID := parts[1]

	// Set the required fields in the resource data
	if err := d.Set("project_id", projectID); err != nil {
		return nil, fmt.Errorf("error setting project_id: %s", err)
	}
	if err := d.Set("user_id", userID); err != nil {
		return nil, fmt.Errorf("error setting user_id: %s", err)
	}

	// Find the user in the project to get additional details
	existingUser, exists, err := c.FindProjectUser(projectID, userID, "")
	if err != nil {
		return nil, fmt.Errorf("error retrieving project user: %s", err)
	}

	if !exists {
		return nil, fmt.Errorf("user %s not found in project %s", userID, projectID)
	}

	// Set the computed fields based on the API response
	if err := d.Set("email", existingUser.Email); err != nil {
		return nil, fmt.Errorf("error setting email: %s", err)
	}
	if err := d.Set("role", existingUser.Role); err != nil {
		return nil, fmt.Errorf("error setting role: %s", err)
	}
	if err := d.Set("added_at", existingUser.AddedAt); err != nil {
		return nil, fmt.Errorf("error setting added_at: %s", err)
	}

	// Explicitly ensure the api_key is empty in the state
	if err := d.Set("api_key", ""); err != nil {
		return nil, fmt.Errorf("error resetting api_key: %s", err)
	}

	return []*schema.ResourceData{d}, nil
}
