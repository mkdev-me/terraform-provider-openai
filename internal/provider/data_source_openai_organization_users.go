package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIOrganizationUsers returns a schema.Resource that represents a data source for OpenAI organization users.
// This data source allows users to retrieve a list of all users in their organization or a specific user.
func dataSourceOpenAIOrganizationUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIOrganizationUsersRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of a specific user to retrieve. If provided, other filter parameters are ignored.",
			},
			"emails": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Filter by the email address of users.",
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the user",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The object type, always 'organization.user'",
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
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role of the user in the organization (owner, member, or reader)",
						},
						"created": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The Unix timestamp when the user was added to the organization",
						},
					},
				},
				Description: "List of users in the organization",
			},
		},
	}
}

// dataSourceOpenAIOrganizationUsersRead handles the read operation for the OpenAI organization users data source.
// It retrieves a list of all users in the organization or a specific user from the OpenAI API.
func dataSourceOpenAIOrganizationUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if a specific user ID was provided
	if userID, ok := d.GetOk("user_id"); ok {
		// If user_id is provided, fetch only that specific user
		id := userID.(string)
		tflog.Debug(ctx, fmt.Sprintf("Retrieving specific user with ID: %s", id))

		user, exists, err := c.GetUser(id)
		if err != nil {
			return diag.Errorf("error retrieving user: %s", err)
		}

		if !exists {
			return diag.Errorf("user with ID %s not found", id)
		}

		// Set a unique ID based on the user ID
		d.SetId(fmt.Sprintf("organization-user-%s", id))

		// Create a list with just this one user
		users := []map[string]interface{}{
			{
				"id":      user.ID,
				"object":  user.Object,
				"email":   user.Email,
				"name":    user.Name,
				"role":    user.Role,
				"created": user.Created,
			},
		}

		if err := d.Set("users", users); err != nil {
			return diag.FromErr(fmt.Errorf("error setting users: %s", err))
		}

		return nil
	}

	// Extract the emails filter if provided
	var emails []string
	if rawEmails, ok := d.GetOk("emails"); ok {
		for _, email := range rawEmails.([]interface{}) {
			emails = append(emails, email.(string))
		}
	}

	// Automatic pagination - fetch all users with default batch size
	const batchSize = 100
	tflog.Debug(ctx, fmt.Sprintf("Fetching all organization users with batch size: %d", batchSize))

	var allUsers []map[string]interface{}
	var after string
	hasMore := true
	pageCount := 0

	// Paginate through all results
	for hasMore {
		pageCount++
		tflog.Debug(ctx, fmt.Sprintf("Fetching page %d with after: %s", pageCount, after))

		resp, err := c.ListUsers(after, batchSize, emails)
		if err != nil {
			return diag.Errorf("error listing organization users (page %d): %s", pageCount, err)
		}

		// Add users from this page to the collection
		for _, orgUser := range resp.Data {
			// Convert to flattened user for consistent field access
			user := orgUser.ToUser()
			u := map[string]interface{}{
				"id":      user.ID,
				"object":  user.Object,
				"email":   user.Email,
				"name":    user.Name,
				"role":    user.Role,
				"created": user.Created,
			}
			allUsers = append(allUsers, u)
		}

		// Check if there are more pages
		hasMore = resp.HasMore
		if hasMore && resp.LastID != "" {
			after = resp.LastID
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Fetched %d total users across %d pages", len(allUsers), pageCount))

	// Set a unique ID for the data source
	d.SetId(fmt.Sprintf("organization-users-all-%d", len(allUsers)))

	// Set the computed values
	if err := d.Set("users", allUsers); err != nil {
		return diag.FromErr(fmt.Errorf("error setting users: %s", err))
	}

	return nil
}
