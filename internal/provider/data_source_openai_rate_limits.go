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

// dataSourceOpenAIRateLimits defines the schema and read operation for the OpenAI rate limits data source.
// This data source allows retrieving information about all rate limits for a specific OpenAI project.
func dataSourceOpenAIRateLimits() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIRateLimitsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project (format: proj_abc123)",
			},
			"rate_limits": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The OpenAI-assigned ID for this rate limit",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The model this rate limit applies to",
						},
						"max_requests_per_minute": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of API requests allowed per minute",
						},
						"max_tokens_per_minute": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of tokens that can be processed per minute",
						},
						"max_images_per_minute": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of images that can be generated per minute",
						},
						"batch_1_day_max_input_tokens": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of input tokens allowed in batch operations per day",
						},
						"max_audio_megabytes_per_1_minute": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of audio megabytes that can be processed per minute",
						},
						"max_requests_per_1_day": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of API requests allowed per day",
						},
					},
				},
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Project-specific API key to use for authentication. If not provided, the provider's default API key will be used.",
			},
		},
	}
}

// dataSourceOpenAIRateLimitsRead fetches information about all rate limits for a specific project from OpenAI.
func dataSourceOpenAIRateLimitsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Log information about field naming convention differences
	tflog.Info(ctx, "[IMPORTANT] OpenAI API uses fields with '_per_1_minute' (with _1_), while Terraform uses '_per_minute'")

	c, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %w", err))
	}

	projectID := d.Get("project_id").(string)
	tflog.Debug(ctx, fmt.Sprintf("Reading all rate limits for project %s", projectID))

	// Use custom API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
		tflog.Debug(ctx, "Using resource-specific API key")
	}

	// Get rate limits
	rateLimits, err := c.ListRateLimitsWithKey(projectID, apiKey)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "403") {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Permission error reading rate limits",
					Detail:   "Make sure you are using an admin API key (sk-admin-...) with proper permissions to read rate limits.",
				},
			}
		}

		return diag.FromErr(fmt.Errorf("error retrieving rate limits: %w", err))
	}

	// Set ID to the project ID with timestamp to ensure uniqueness
	d.SetId(fmt.Sprintf("%s:%d", projectID, time.Now().Unix()))

	// Log the number of rate limits found
	tflog.Debug(ctx, fmt.Sprintf("Found %d rate limits for project %s", len(rateLimits.Data), projectID))

	// Create a list to store the flatten rate limits
	rateLimitsJSON := make([]interface{}, 0, len(rateLimits.Data))

	// Verify model filter if provided
	var modelFilter string
	if v, ok := d.GetOk("model"); ok {
		modelFilter = v.(string)
		tflog.Debug(ctx, fmt.Sprintf("Filtering rate limits by model: %s", modelFilter))
	}

	// Format each rate limit for Terraform state
	for _, rateLimit := range rateLimits.Data {
		// Check model filter
		if modelFilter != "" && rateLimit.Model != modelFilter {
			continue
		}

		// Create map for this rate limit
		rateLimitMap := map[string]interface{}{
			"id":                               rateLimit.ID,
			"model":                            rateLimit.Model,
			"max_requests_per_minute":          rateLimit.MaxRequestsPer1Minute,
			"max_tokens_per_minute":            rateLimit.MaxTokensPer1Minute,
			"max_images_per_minute":            rateLimit.MaxImagesPer1Minute,
			"batch_1_day_max_input_tokens":     rateLimit.Batch1DayMaxInputTokens,
			"max_audio_megabytes_per_1_minute": rateLimit.MaxAudioMegabytesPer1Minute,
			"max_requests_per_1_day":           rateLimit.MaxRequestsPer1Day,
		}

		// Log each rate limit for debugging
		tflog.Debug(ctx, fmt.Sprintf("Rate limit: ID=%s, Model=%s, MaxReq=%d, MaxTokens=%d, MaxImg=%d",
			rateLimit.ID, rateLimit.Model, rateLimit.MaxRequestsPer1Minute,
			rateLimit.MaxTokensPer1Minute, rateLimit.MaxImagesPer1Minute))

		rateLimitsJSON = append(rateLimitsJSON, rateLimitMap)
	}

	// Set the result
	if err := d.Set("rate_limits", rateLimitsJSON); err != nil {
		return diag.FromErr(fmt.Errorf("error setting rate_limits: %w", err))
	}

	return diag.Diagnostics{}
}
