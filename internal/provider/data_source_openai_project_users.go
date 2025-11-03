package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIProjectUsers returns a schema.Resource that represents a data source for OpenAI project users.
// This data source allows users to retrieve a list of all users in a specific project.
func dataSourceOpenAIProjectUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIProjectUsersRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project to retrieve users from",
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
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The email address of the user",
						},
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role of the user in the project (owner or member)",
						},
						"added_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Timestamp when the user was added to the project",
						},
					},
				},
				Description: "List of users in the project",
			},
			"user_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of user IDs in the project (non-sensitive)",
			},
			"user_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of users in the project (non-sensitive)",
			},
			"owner_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of user IDs with owner role in the project (non-sensitive)",
			},
			"member_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of user IDs with member role in the project (non-sensitive)",
			},
		},
	}
}

// dataSourceOpenAIProjectUsersRead handles the read operation for the OpenAI project users data source.
// It retrieves a list of all users in a specific project from the OpenAI API.
func dataSourceOpenAIProjectUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		return diag.FromErr(fmt.Errorf("project_id is required"))
	}

	// Set the ID to the project_id
	d.SetId(projectID)

	// Automatic pagination - fetch all users with default batch size
	const batchSize = 100
	tflog.Debug(ctx, fmt.Sprintf("Fetching all users in project %s with batch size: %d", projectID, batchSize))

	var allUsers []map[string]interface{}
	var after string
	hasMore := true
	pageCount := 0

	// Paginate through all results
	for hasMore {
		pageCount++
		tflog.Debug(ctx, fmt.Sprintf("Fetching page %d for project %s (after: %s)", pageCount, projectID, after))

		usersList, err := c.ListProjectUsers(projectID, after, batchSize)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error listing users (page %d): %v", pageCount, err))
			return diag.Errorf("Error listing users in project (page %d): %s", pageCount, err)
		}

		// Transform the users from this page into a format appropriate for the schema
		for _, user := range usersList.Data {
			userData := map[string]interface{}{
				"id":       user.ID,
				"email":    user.Email,
				"role":     user.Role,
				"added_at": user.AddedAt,
			}
			allUsers = append(allUsers, userData)
		}

		// Check if there are more pages
		hasMore = usersList.HasMore
		if hasMore && usersList.LastID != "" {
			after = usersList.LastID
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Fetched %d total users for project %s across %d pages", len(allUsers), projectID, pageCount))

	// Use allUsers instead of users
	users := allUsers

	if err := d.Set("users", users); err != nil {
		return diag.FromErr(fmt.Errorf("error setting users: %s", err))
	}

	// Extract user IDs
	userIDs := make([]string, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user["id"].(string))
	}

	if err := d.Set("user_ids", userIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_ids: %s", err))
	}

	// Set user count for easy access
	if err := d.Set("user_count", len(users)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_count: %s", err))
	}

	// Extract owner and member IDs
	ownerIDs := make([]string, 0)
	memberIDs := make([]string, 0)
	for _, user := range users {
		role := user["role"].(string)
		if role == "owner" {
			ownerIDs = append(ownerIDs, user["id"].(string))
		} else if role == "member" {
			memberIDs = append(memberIDs, user["id"].(string))
		}
	}

	if err := d.Set("owner_ids", ownerIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting owner_ids: %s", err))
	}

	if err := d.Set("member_ids", memberIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting member_ids: %s", err))
	}

	return nil
}
