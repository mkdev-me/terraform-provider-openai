package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceOpenAIUserRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIUserRoleCreate,
		ReadContext:   resourceOpenAIUserRoleRead,
		UpdateContext: resourceOpenAIUserRoleUpdate,
		DeleteContext: resourceOpenAIUserRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIUserRoleImport,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the user",
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
				Sensitive:   true,
				Description: "API key for authentication. If not provided, the provider's default API key will be used.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address of the user",
			},
		},
	}
}

func resourceOpenAIUserRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Get("user_id").(string)
	role := d.Get("role").(string)

	// Set the ID to be the user_id since this is a user-specific resource
	d.SetId(userID)

	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Check if user exists and get current role
	user, exists, err := c.GetUser(userID, apiKey)
	if err != nil {
		return diag.Errorf("Error checking user: %s", err)
	}

	if !exists {
		return diag.Errorf("User %s not found", userID)
	}

	// Update the user's role if it's different
	if user.Role != role {
		updatedUser, err := c.UpdateUserRole(userID, role, apiKey)
		if err != nil {
			return diag.Errorf("Error updating user role: %s", err)
		}

		if err := d.Set("email", updatedUser.Email); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set email: %v", err))
		}
	}

	return resourceOpenAIUserRoleRead(ctx, d, m)
}

func resourceOpenAIUserRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Id()
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	user, exists, err := c.GetUser(userID, apiKey)
	if err != nil {
		return diag.Errorf("Error reading user: %s", err)
	}

	if !exists {
		d.SetId("")
		return nil
	}

	if err := d.Set("role", user.Role); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role: %v", err))
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set email: %v", err))
	}

	return nil
}

func resourceOpenAIUserRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	if !d.HasChange("role") {
		return nil
	}

	userID := d.Id()
	role := d.Get("role").(string)
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	updatedUser, err := c.UpdateUserRole(userID, role, apiKey)
	if err != nil {
		return diag.Errorf("Error updating user role: %s", err)
	}

	if err := d.Set("email", updatedUser.Email); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set email: %v", err))
	}

	return resourceOpenAIUserRoleRead(ctx, d, m)
}

func resourceOpenAIUserRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Since we can't delete users, we'll just remove it from state
	d.SetId("")
	return nil
}

// resourceOpenAIUserRoleImport imports an existing user role into Terraform state.
// The ID should be the user_id.
func resourceOpenAIUserRoleImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return nil, err
	}

	userID := d.Id()

	// Set the required fields in the resource data
	if err := d.Set("user_id", userID); err != nil {
		return nil, fmt.Errorf("error setting user_id: %s", err)
	}

	// Get the user details from the API
	user, exists, err := c.GetUser(userID, "")
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %s", err)
	}

	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	// Set the computed fields based on the API response
	if err := d.Set("email", user.Email); err != nil {
		return nil, fmt.Errorf("error setting email: %s", err)
	}
	if err := d.Set("role", user.Role); err != nil {
		return nil, fmt.Errorf("error setting role: %s", err)
	}

	return []*schema.ResourceData{d}, nil
}
