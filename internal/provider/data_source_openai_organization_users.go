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
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "API key for authentication. If not provided, the provider's default API key will be used.",
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
			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of a specific user to retrieve. If provided, other filter parameters are ignored.",
			},
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A cursor for use in pagination. 'after' is an object ID that defines your place in the list.",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "A limit on the number of objects to be returned. Limit can range between 1 and 100, and the default is 20.",
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
						"added_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The Unix timestamp when the user was added to the organization",
						},
					},
				},
				Description: "List of users in the organization",
			},
			"first_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the first user in the list",
			},
			"last_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the last user in the list",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more users that can be retrieved by paginating",
			},
		},
	}
}

// dataSourceOpenAIOrganizationUsersRead handles the read operation for the OpenAI organization users data source.
// It retrieves a list of all users in the organization or a specific user from the OpenAI API.
func dataSourceOpenAIOrganizationUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Check if a specific user ID was provided
	if userID, ok := d.GetOk("user_id"); ok {
		// If user_id is provided, fetch only that specific user
		id := userID.(string)
		tflog.Debug(ctx, fmt.Sprintf("Retrieving specific user with ID: %s", id))

		user, exists, err := c.GetUser(id, apiKey)
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
				"id":       user.ID,
				"object":   user.Object,
				"email":    user.Email,
				"name":     user.Name,
				"role":     user.Role,
				"added_at": user.AddedAt,
			},
		}

		if err := d.Set("users", users); err != nil {
			return diag.FromErr(fmt.Errorf("error setting users: %s", err))
		}

		// For a single user, has_more is always false
		if err := d.Set("has_more", false); err != nil {
			return diag.FromErr(fmt.Errorf("error setting has_more: %s", err))
		}

		// Set first_id and last_id to the user's ID
		if err := d.Set("first_id", user.ID); err != nil {
			return diag.FromErr(fmt.Errorf("error setting first_id: %s", err))
		}

		if err := d.Set("last_id", user.ID); err != nil {
			return diag.FromErr(fmt.Errorf("error setting last_id: %s", err))
		}

		// Explicitly set the api_key to empty in the state
		if err := d.Set("api_key", ""); err != nil {
			return diag.FromErr(fmt.Errorf("failed to reset api_key: %v", err))
		}

		return nil
	}

	// If no specific user ID, list users with filters
	after := d.Get("after").(string)
	limit := d.Get("limit").(int)

	// Extract the emails filter if provided
	var emails []string
	if rawEmails, ok := d.GetOk("emails"); ok {
		for _, email := range rawEmails.([]interface{}) {
			emails = append(emails, email.(string))
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Listing organization users with limit: %d", limit))

	// Call the API
	resp, err := c.ListUsers(after, limit, emails, apiKey)
	if err != nil {
		return diag.Errorf("error listing organization users: %s", err)
	}

	// Set a unique ID based on the query parameters
	d.SetId(fmt.Sprintf("organization-users-%s-%d", after, limit))

	// Prepare the users list
	users := make([]map[string]interface{}, 0, len(resp.Data))
	for _, user := range resp.Data {
		u := map[string]interface{}{
			"id":       user.ID,
			"object":   user.Object,
			"email":    user.Email,
			"name":     user.Name,
			"role":     user.Role,
			"added_at": user.AddedAt,
		}
		users = append(users, u)
	}

	// Set the computed values
	if err := d.Set("users", users); err != nil {
		return diag.FromErr(fmt.Errorf("error setting users: %s", err))
	}

	if err := d.Set("has_more", resp.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %s", err))
	}

	if err := d.Set("first_id", resp.FirstID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting first_id: %s", err))
	}

	if err := d.Set("last_id", resp.LastID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting last_id: %s", err))
	}

	// Explicitly set the api_key to empty in the state
	if err := d.Set("api_key", ""); err != nil {
		return diag.FromErr(fmt.Errorf("failed to reset api_key: %v", err))
	}

	return nil
}
