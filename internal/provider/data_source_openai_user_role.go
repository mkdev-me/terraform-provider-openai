package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIUserRole returns a schema.Resource that represents a data source for an OpenAI user role.
// This data source allows users to retrieve information about a specific user's role in the organization.
func dataSourceOpenAIUserRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIUserRoleRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the user to retrieve the role for",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "API key for authentication. If not provided, the provider's default API key will be used.",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role of the user in the organization (owner or member)",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address of the user",
			},
		},
	}
}

// dataSourceOpenAIUserRoleRead handles the read operation for the OpenAI user role data source.
// It retrieves information about a specific user role from the OpenAI API.
func dataSourceOpenAIUserRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Get("user_id").(string)
	if userID == "" {
		return diag.FromErr(fmt.Errorf("user_id is required"))
	}

	// Get custom API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Retrieve the user information
	user, exists, err := c.GetUser(userID, apiKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving user: %s", err))
	}

	if !exists {
		return diag.FromErr(fmt.Errorf("user with ID %s not found", userID))
	}

	// Set the resource ID
	d.SetId(userID)

	// Update the state with the user details
	if err := d.Set("role", user.Role); err != nil {
		return diag.FromErr(fmt.Errorf("error setting role: %s", err))
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email: %s", err))
	}

	return nil
}
