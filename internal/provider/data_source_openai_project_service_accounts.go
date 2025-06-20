package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIProjectServiceAccounts returns the schema and operations for listing all OpenAI Project Service Accounts
func dataSourceOpenAIProjectServiceAccounts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIProjectServiceAccountsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project from which to retrieve service accounts",
			},
			"service_accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the service account",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the service account",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp (in Unix time) when the service account was created",
						},
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role of the service account",
						},
						"api_key_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the API key associated with the service account",
						},
					},
				},
				Description: "List of service accounts in the project",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

// dataSourceOpenAIProjectServiceAccountsRead reads all service accounts in a project
func dataSourceOpenAIProjectServiceAccountsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, err := GetOpenAIClientWithAdminKey(m)
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics

	// Get input values
	projectID := d.Get("project_id").(string)

	// Get all service accounts for the project
	tflog.Debug(ctx, "Listing OpenAI project service accounts", map[string]interface{}{
		"project_id": projectID,
	})

	// Note: We're assuming the client method returns a slice of ServiceAccount objects
	// If this doesn't match the actual implementation, adjust accordingly
	serviceAccounts, err := c.ListProjectServiceAccounts(projectID)
	if err != nil {
		// Handle permission errors gracefully
		if strings.Contains(err.Error(), "insufficient permissions") || strings.Contains(err.Error(), "Missing scopes") {
			tflog.Info(ctx, "Permission error listing service accounts", map[string]interface{}{
				"error": err.Error(),
			})
			return diag.FromErr(fmt.Errorf("cannot retrieve service accounts due to permission issues: %w", err))
		}

		// Check if the error is a 404, which means the project is not found
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return diag.FromErr(fmt.Errorf("project with ID %s not found", projectID))
		}

		return diag.FromErr(fmt.Errorf("error listing project service accounts: %w", err))
	}

	// Generate a unique ID for this data source
	d.SetId(fmt.Sprintf("%s-service-accounts-%d", projectID, time.Now().Unix()))

	// Format the service accounts into a list of maps for Terraform
	// Assuming serviceAccounts is a struct with a Data field that contains the slice
	accountsList := make([]map[string]interface{}, 0)
	if serviceAccounts != nil && serviceAccounts.Data != nil {
		for _, account := range serviceAccounts.Data {
			accountMap := map[string]interface{}{
				"id":         account.ID,
				"name":       account.Name,
				"created_at": account.CreatedAt,
				"role":       account.Role,
			}

			// Add API key ID if available
			if account.APIKey != nil {
				accountMap["api_key_id"] = account.APIKey.ID
			}

			accountsList = append(accountsList, accountMap)
		}
	}

	if err := d.Set("service_accounts", accountsList); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set service_accounts: %v", err))
	}

	return diags
}
