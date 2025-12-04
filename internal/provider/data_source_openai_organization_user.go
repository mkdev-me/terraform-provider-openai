package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

// dataSourceOpenAIOrganizationUser returns a schema.Resource that represents a data source for a specific OpenAI organization user.
// This data source allows users to retrieve information about a specific user in their organization.
func dataSourceOpenAIOrganizationUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIOrganizationUserRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The ID of the user to retrieve",
				AtLeastOneOf: []string{"user_id", "email"},
			},
			"email": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The email address of the user to retrieve",
				AtLeastOneOf: []string{"user_id", "email"},
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the user",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role of the user in the organization (owner, member, or reader)",
			},
			"added_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The Unix timestamp when the user was added to the organization",
			},
		},
	}
}

// dataSourceOpenAIOrganizationUserRead handles the read operation for the OpenAI organization user data source.
// It retrieves information about a specific user from the OpenAI API.
func dataSourceOpenAIOrganizationUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	var user *client.User
	var exists bool

	// Check if we're looking up by user_id or email
	if userID, ok := d.GetOk("user_id"); ok {
		// Look up by user ID
		userID := userID.(string)
		tflog.Debug(ctx, fmt.Sprintf("Retrieving user with ID: %s", userID))

		// Call the API to get the user using the provider's API key
		user, exists, err = c.GetUser(userID)
		if err != nil {
			return diag.Errorf("error retrieving user by ID: %s", err)
		}

		if !exists {
			return diag.Errorf("user with ID %s not found", userID)
		}
	} else if email, ok := d.GetOk("email"); ok {
		// Look up by email
		email := email.(string)
		tflog.Debug(ctx, fmt.Sprintf("Retrieving user with email: %s", email))

		// Call the API to find the user by email using the provider's API key
		user, exists, err = c.FindUserByEmail(email)
		if err != nil {
			return diag.Errorf("error retrieving user by email: %s", err)
		}

		if !exists {
			return diag.Errorf("user with email %s not found", email)
		}
	} else {
		// This should never happen due to schema validation
		return diag.Errorf("either user_id or email must be provided")
	}

	// Set the resource ID
	d.SetId(user.ID)

	// Set the user_id field in case lookup was by email
	if err := d.Set("user_id", user.ID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_id: %s", err))
	}

	// Set the email field in case lookup was by user_id
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email: %s", err))
	}

	// Set the computed values
	if err := d.Set("name", user.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %s", err))
	}

	if err := d.Set("role", user.Role); err != nil {
		return diag.FromErr(fmt.Errorf("error setting role: %s", err))
	}

	if err := d.Set("added_at", user.AddedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting added_at: %s", err))
	}

	return nil
}
