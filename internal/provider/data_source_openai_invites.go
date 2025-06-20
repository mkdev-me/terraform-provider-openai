package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fjcorp/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIInvites returns a schema.Resource that represents a data source for OpenAI invites.
// This data source allows users to retrieve a list of all pending invitations in an OpenAI organization.
func dataSourceOpenAIInvites() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIInvitesRead,
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "API key for authentication. If not provided, the provider's default API key will be used.",
			},
			"invites": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the invitation",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The email address of the invited user",
						},
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role assigned to the invited user (owner or reader)",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the invitation",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the invitation was created",
						},
						"expires_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the invitation expires",
						},
						"projects": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the project",
									},
									"role": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The role assigned to the user within the project (owner or member)",
									},
								},
							},
							Description: "Projects assigned to the invited user",
						},
					},
				},
				Description: "List of pending invitations",
			},
		},
	}
}

// dataSourceOpenAIInvitesRead handles the read operation for the OpenAI invites data source.
// It retrieves a list of all pending invitations from the OpenAI API.
func dataSourceOpenAIInvitesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get custom API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Set a unique ID for the data source
	d.SetId(fmt.Sprintf("invites_%d", time.Now().Unix()))

	// Add retry logic for the API call with extended timeout
	var invitesResponse *client.ListInvitesResponse
	err = retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		var err error
		invitesResponse, err = c.ListInvites(apiKey)
		if err != nil {
			// Check specifically for timeout-related errors
			if strings.Contains(err.Error(), "timeout") ||
				strings.Contains(err.Error(), "deadline exceeded") ||
				strings.Contains(err.Error(), "504") {
				return retry.RetryableError(fmt.Errorf("API timeout when listing invitations (this is common for organizations with many invites), retrying: %s", err))
			}
			return retry.RetryableError(fmt.Errorf("error listing invitations, retrying: %s", err))
		}
		return nil
	})

	if err != nil {
		// Set empty invites list with helpful error message if timeout persists
		if strings.Contains(err.Error(), "timeout") ||
			strings.Contains(err.Error(), "deadline exceeded") ||
			strings.Contains(err.Error(), "504") {
			d.Set("invites", []map[string]interface{}{})
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "OpenAI API Timeout",
					Detail:   "The OpenAI API timed out while retrieving invitations. This is common for organizations with many pending invites. The data source will return an empty list. You can try again later or contact OpenAI support if this persists.",
				},
			}
		}
		return diag.FromErr(fmt.Errorf("error listing invitations after retries: %s", err))
	}

	if invitesResponse == nil {
		return diag.FromErr(fmt.Errorf("unexpected nil response when listing invitations"))
	}

	// Format the invites for Terraform
	formattedInvites := make([]map[string]interface{}, len(invitesResponse.Data))
	for i, invite := range invitesResponse.Data {
		// Format projects if present
		projects := make([]map[string]interface{}, len(invite.Projects))
		for j, project := range invite.Projects {
			projects[j] = map[string]interface{}{
				"id":   project.ID,
				"role": project.Role,
			}
		}

		// Format the invite
		formattedInvites[i] = map[string]interface{}{
			"id":         invite.ID,
			"email":      invite.Email,
			"role":       invite.Role,
			"status":     invite.Status,
			"created_at": time.Unix(int64(invite.CreatedAt), 0).Format(time.RFC3339),
			"expires_at": time.Unix(int64(invite.ExpiresAt), 0).Format(time.RFC3339),
			"projects":   projects,
		}
	}

	// Update the state with the formatted invites
	if err := d.Set("invites", formattedInvites); err != nil {
		return diag.FromErr(fmt.Errorf("error setting invites: %s", err))
	}

	return nil
}
