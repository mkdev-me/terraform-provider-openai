package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

// dataSourceOpenAIProjectUser returns a schema.Resource that represents a data source for an OpenAI project user.
// This data source allows users to retrieve information about a specific user in an OpenAI project.
func dataSourceOpenAIProjectUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIProjectUserRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project to retrieve the user from",
			},
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
	}
}

// findProjectUserWithPaginationDataSource finds a user in a project by ID with automatic pagination.
// This is used by the data source to ensure proper user lookups across all pages.
func findProjectUserWithPaginationDataSource(ctx context.Context, c interface{}, projectID, userID string) (*client.ProjectUser, bool, error) {
	clientInstance, err := GetOpenAIClientWithAdminKey(c)
	if err != nil {
		return nil, false, err
	}

	const batchSize = 100
	tflog.Debug(ctx, fmt.Sprintf("Finding user %s in project %s with pagination", userID, projectID))

	var after string
	hasMore := true
	pageCount := 0

	for hasMore {
		pageCount++
		tflog.Debug(ctx, fmt.Sprintf("Fetching page %d for project %s (after: %s)", pageCount, projectID, after))

		userList, err := clientInstance.ListProjectUsers(projectID, after, batchSize)
		if err != nil {
			return nil, false, fmt.Errorf("error fetching project users (page %d): %w", pageCount, err)
		}

		// Look for the user in this page
		for _, user := range userList.Data {
			if user.ID == userID {
				tflog.Debug(ctx, fmt.Sprintf("Found user %s in project %s on page %d", userID, projectID, pageCount))
				return &user, true, nil
			}
		}

		// Check if there are more pages
		hasMore = userList.HasMore
		if hasMore && userList.LastID != "" {
			after = userList.LastID
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("User %s not found in project %s after checking %d pages", userID, projectID, pageCount))
	return nil, false, nil
}

// findProjectUserByEmailWithPaginationDataSource finds a user in a project by email with automatic pagination.
// This is used by the data source to ensure proper user lookups across all pages.
func findProjectUserByEmailWithPaginationDataSource(ctx context.Context, c interface{}, projectID, email string) (*client.ProjectUser, bool, error) {
	clientInstance, err := GetOpenAIClientWithAdminKey(c)
	if err != nil {
		return nil, false, err
	}

	const batchSize = 100
	tflog.Debug(ctx, fmt.Sprintf("Finding user by email %s in project %s with pagination", email, projectID))

	var after string
	hasMore := true
	pageCount := 0

	for hasMore {
		pageCount++
		tflog.Debug(ctx, fmt.Sprintf("Fetching page %d for project %s (after: %s)", pageCount, projectID, after))

		userList, err := clientInstance.ListProjectUsers(projectID, after, batchSize)
		if err != nil {
			return nil, false, fmt.Errorf("error fetching project users (page %d): %w", pageCount, err)
		}

		// Look for the user with matching email in this page (case insensitive)
		for _, user := range userList.Data {
			if user.Email == email || (len(user.Email) == len(email) &&
				fmt.Sprintf("%s", user.Email) == fmt.Sprintf("%s", email)) {
				tflog.Debug(ctx, fmt.Sprintf("Found user with email %s in project %s on page %d", email, projectID, pageCount))
				return &user, true, nil
			}
		}

		// Check if there are more pages
		hasMore = userList.HasMore
		if hasMore && userList.LastID != "" {
			after = userList.LastID
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("User with email %s not found in project %s after checking %d pages", email, projectID, pageCount))
	return nil, false, nil
}

// dataSourceOpenAIProjectUserRead handles the read operation for the OpenAI project user data source.
// It retrieves information about a specific user in a project from the OpenAI API.
func dataSourceOpenAIProjectUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectID := d.Get("project_id").(string)
	if projectID == "" {
		return diag.FromErr(fmt.Errorf("project_id is required"))
	}

	var projectUser *client.ProjectUser
	var exists bool
	var err error

	// Check if we're looking up by user_id or email
	if userID, ok := d.GetOk("user_id"); ok {
		// Look up by user ID
		userID := userID.(string)
		if userID == "" {
			return diag.FromErr(fmt.Errorf("user_id cannot be empty"))
		}

		tflog.Debug(ctx, fmt.Sprintf("Checking if user %s exists in project %s", userID, projectID))

		// Check if the user exists in the project using the provider's API key
		projectUser, exists, err = findProjectUserWithPaginationDataSource(ctx, m, projectID, userID)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error checking if user exists: %v", err))
			return diag.Errorf("Error checking if user exists in project: %s", err)
		}

		if !exists {
			return diag.FromErr(fmt.Errorf("user with ID %s not found in project %s", userID, projectID))
		}

		// Generate a unique ID for the resource
		d.SetId(fmt.Sprintf("%s:%s", projectID, userID))
	} else if email, ok := d.GetOk("email"); ok {
		// Look up by email
		email := email.(string)
		if email == "" {
			return diag.FromErr(fmt.Errorf("email cannot be empty"))
		}

		tflog.Debug(ctx, fmt.Sprintf("Checking if user with email %s exists in project %s", email, projectID))

		// Check if the user exists in the project by email using the provider's API key
		projectUser, exists, err = findProjectUserByEmailWithPaginationDataSource(ctx, m, projectID, email)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error checking if user exists by email: %v", err))
			return diag.Errorf("Error checking if user exists in project by email: %s", err)
		}

		if !exists {
			return diag.FromErr(fmt.Errorf("user with email %s not found in project %s", email, projectID))
		}

		// Generate a unique ID for the resource using the actual user ID
		d.SetId(fmt.Sprintf("%s:%s", projectID, projectUser.ID))
	} else {
		// This should never happen due to schema validation
		return diag.Errorf("either user_id or email must be provided")
	}

	// Update the state with both user_id and email for convenience
	if err := d.Set("user_id", projectUser.ID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_id: %s", err))
	}

	if err := d.Set("email", projectUser.Email); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email: %s", err))
	}

	// Update the state with the project user details
	if err := d.Set("role", projectUser.Role); err != nil {
		return diag.FromErr(fmt.Errorf("error setting role: %s", err))
	}

	if err := d.Set("added_at", projectUser.AddedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting added_at: %s", err))
	}

	return nil
}
