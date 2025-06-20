package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAIRateLimit defines the schema and read operation for the OpenAI rate limit data source.
// This data source allows retrieving information about rate limits for a specific model in an OpenAI project.
func dataSourceOpenAIRateLimit() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIRateLimitRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project (format: proj_abc123)",
			},
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The model to get rate limits for (e.g., 'gpt-4', 'gpt-3.5-turbo')",
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
			"rate_limit_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The OpenAI-assigned ID for this rate limit",
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

// dataSourceOpenAIRateLimitRead performs the API request to get the rate limits for a model.
func dataSourceOpenAIRateLimitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Log information about field naming convention differences
	tflog.Info(ctx, "[IMPORTANT] OpenAI API uses fields with '_per_1_minute' (with _1_), while Terraform uses '_per_minute'")

	c, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %w", err))
	}

	projectID := d.Get("project_id").(string)
	model := d.Get("model").(string)

	tflog.Debug(ctx, fmt.Sprintf("Reading rate limit for model %s in project %s", model, projectID))

	// Use custom API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
		tflog.Debug(ctx, "Using resource-specific API key")
	}

	// Get rate limit data
	rateLimit, err := c.GetRateLimitWithKey(projectID, model, apiKey)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "403") {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Permission error reading rate limit",
					Detail:   "Make sure you are using an admin API key (sk-admin-...) with proper permissions to read rate limits.",
				},
			}
		}

		return diag.FromErr(fmt.Errorf("error retrieving rate limit: %w", err))
	}

	// Set ID to combination of project ID and rate limit ID for uniqueness
	// Format: project_id:rate_limit_id
	d.SetId(fmt.Sprintf("%s:%s", projectID, rateLimit.ID))

	tflog.Debug(ctx, fmt.Sprintf("Retrieved rate limit: ID=%s, Model=%s, MaxReq=%d, MaxTokens=%d, MaxImg=%d",
		rateLimit.ID, rateLimit.Model, rateLimit.MaxRequestsPer1Minute, rateLimit.MaxTokensPer1Minute, rateLimit.MaxImagesPer1Minute))

	// Map API fields to Terraform schema fields
	if err := d.Set("model", rateLimit.Model); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("max_requests_per_minute", rateLimit.MaxRequestsPer1Minute); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("max_tokens_per_minute", rateLimit.MaxTokensPer1Minute); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("max_images_per_minute", rateLimit.MaxImagesPer1Minute); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("batch_1_day_max_input_tokens", rateLimit.Batch1DayMaxInputTokens); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("max_audio_megabytes_per_1_minute", rateLimit.MaxAudioMegabytesPer1Minute); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("max_requests_per_1_day", rateLimit.MaxRequestsPer1Day); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rate_limit_id", rateLimit.ID); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
