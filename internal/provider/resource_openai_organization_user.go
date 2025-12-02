package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// resourceOpenAIOrganizationUser defines the schema and CRUD operations for OpenAI organization users.
// This resource allows users to manage organization users through Terraform.
// Note: Users cannot be created through this resource, they must already exist.
// The resource allows updating user roles and removing users from the organization.
func resourceOpenAIOrganizationUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIOrganizationUserCreate,
		ReadContext:   resourceOpenAIOrganizationUserRead,
		UpdateContext: resourceOpenAIOrganizationUserUpdate,
		DeleteContext: resourceOpenAIOrganizationUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the user to manage in the organization",
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"owner", "reader"}, false),
				Description:  "The role to assign to the user (owner or reader)",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address of the user",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the user",
			},
			"created": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The Unix timestamp when the user was added to the organization",
			},
		},
	}
}

// resourceOpenAIOrganizationUserCreate handles the creation of the organization user resource.
// Since users cannot be created through the API, this function verifies the user exists
// and updates their role if necessary.
func resourceOpenAIOrganizationUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Get("user_id").(string)
	role := d.Get("role").(string)

	// Check if the user exists in the organization
	tflog.Debug(ctx, fmt.Sprintf("Checking if user %s exists in the organization", userID))
	user, exists, err := c.GetUser(userID)
	if err != nil {
		return diag.Errorf("Error checking if user exists: %s", err)
	}

	if !exists {
		return diag.Errorf("User with ID %s does not exist in the organization", userID)
	}

	// Set the resource ID
	d.SetId(userID)

	// Update the user's role if it's different from the current role
	if user.Role != role {
		tflog.Info(ctx, fmt.Sprintf("Updating user %s role from %s to %s", userID, user.Role, role))
		updatedUser, err := c.UpdateUserRole(userID, role)
		if err != nil {
			return diag.Errorf("Error updating user role: %s", err)
		}
		user = updatedUser
	}

	// Set the computed fields
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email: %s", err))
	}
	if err := d.Set("name", user.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %s", err))
	}
	if err := d.Set("created", user.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %s", err))
	}

	return nil
}

// resourceOpenAIOrganizationUserRead retrieves information about a user in the organization.
func resourceOpenAIOrganizationUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Id()

	// Get the user information
	tflog.Debug(ctx, fmt.Sprintf("Retrieving user with ID: %s", userID))
	user, exists, err := c.GetUser(userID)
	if err != nil {
		return diag.Errorf("Error retrieving user: %s", err)
	}

	if !exists {
		// User no longer exists, remove from state
		tflog.Warn(ctx, fmt.Sprintf("User %s no longer exists in the organization, removing from state", userID))
		d.SetId("")
		return nil
	}

	// Update the state with the user information
	if err := d.Set("user_id", user.ID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_id: %s", err))
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email: %s", err))
	}
	if err := d.Set("name", user.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %s", err))
	}
	if err := d.Set("role", user.Role); err != nil {
		return diag.FromErr(fmt.Errorf("error setting role: %s", err))
	}
	if err := d.Set("created", user.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %s", err))
	}

	return nil
}

// resourceOpenAIOrganizationUserUpdate updates a user's role in the organization.
func resourceOpenAIOrganizationUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Id()

	if d.HasChange("role") {
		newRole := d.Get("role").(string)

		tflog.Info(ctx, fmt.Sprintf("Updating user %s role to %s", userID, newRole))
		user, err := c.UpdateUserRole(userID, newRole)
		if err != nil {
			return diag.Errorf("Error updating user role: %s", err)
		}

		// Update the computed fields
		if err := d.Set("email", user.Email); err != nil {
			return diag.FromErr(fmt.Errorf("error setting email: %s", err))
		}
		if err := d.Set("name", user.Name); err != nil {
			return diag.FromErr(fmt.Errorf("error setting name: %s", err))
		}
		if err := d.Set("created", user.Created); err != nil {
			return diag.FromErr(fmt.Errorf("error setting created: %s", err))
		}
	}

	return nil
}

// resourceOpenAIOrganizationUserDelete removes a user from the organization.
func resourceOpenAIOrganizationUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Id()

	tflog.Info(ctx, fmt.Sprintf("Deleting user %s from organization", userID))
	err = c.DeleteUser(userID)
	if err != nil {
		return diag.Errorf("Error deleting user: %s", err)
	}

	d.SetId("")
	return nil
}
