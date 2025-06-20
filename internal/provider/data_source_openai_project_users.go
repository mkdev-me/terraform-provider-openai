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
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		return diag.FromErr(fmt.Errorf("project_id is required"))
	}

	// Set the ID to the project_id
	d.SetId(projectID)

	// Get custom API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
		tflog.Debug(ctx, "Using custom API key for listing project users")
	}

	// List all users in the project
	tflog.Debug(ctx, fmt.Sprintf("Listing users in project %s", projectID))
	usersList, err := c.ListProjectUsers(projectID, apiKey)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error listing users: %v", err))
		return diag.Errorf("Error listing users in project: %s", err)
	}

	// Transform the users into a format appropriate for the schema
	users := make([]map[string]interface{}, 0, len(usersList.Data))
	for _, user := range usersList.Data {
		userData := map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"role":     user.Role,
			"added_at": user.AddedAt,
		}
		users = append(users, userData)
	}

	if err := d.Set("users", users); err != nil {
		return diag.FromErr(fmt.Errorf("error setting users: %s", err))
	}

	// Extract user IDs
	userIDs := make([]string, 0, len(usersList.Data))
	for _, user := range usersList.Data {
		userIDs = append(userIDs, user.ID)
	}

	if err := d.Set("user_ids", userIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_ids: %s", err))
	}

	// Set user count for easy access
	if err := d.Set("user_count", len(usersList.Data)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_count: %s", err))
	}

	// Extract owner and member IDs
	ownerIDs := make([]string, 0)
	memberIDs := make([]string, 0)
	for _, user := range usersList.Data {
		if user.Role == "owner" {
			ownerIDs = append(ownerIDs, user.ID)
		} else if user.Role == "member" {
			memberIDs = append(memberIDs, user.ID)
		}
	}

	if err := d.Set("owner_ids", ownerIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting owner_ids: %s", err))
	}

	if err := d.Set("member_ids", memberIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting member_ids: %s", err))
	}

	// Explicitly set the api_key to empty in the state
	if err := d.Set("api_key", ""); err != nil {
		return diag.FromErr(fmt.Errorf("failed to reset api_key: %v", err))
	}

	return nil
}
