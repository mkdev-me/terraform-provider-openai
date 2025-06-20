package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/fjcorp/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceOpenAIRateLimit defines the schema and CRUD operations for OpenAI rate limits.
// This resource allows users to manage rate limits for OpenAI projects through Terraform,
// including creation, reading, updating, and deletion of rate limits for specific models.
// Rate limits help control API usage and costs by setting caps on requests, tokens, and images.
func resourceOpenAIRateLimit() *schema.Resource {
	return &schema.Resource{
		Description: "Manages rate limits for an OpenAI model in a project. Note that rate limits cannot be truly deleted via the API, so this resource will reset rate limits to defaults when removed. This resource requires an admin API key with the api.management.read scope for full functionality, but will gracefully handle permission errors to allow operations to continue.",

		CreateContext: resourceOpenAIRateLimitCreate,
		ReadContext:   resourceOpenAIRateLimitRead,
		UpdateContext: resourceOpenAIRateLimitUpdate,
		DeleteContext: resourceOpenAIRateLimitDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIRateLimitImport,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the project to set rate limits for.",
			},
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The OpenAI model name to set rate limits for.",
			},
			"max_requests_per_minute": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of requests per minute.",
			},
			"max_tokens_per_minute": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of tokens per minute.",
			},
			"max_images_per_minute": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of images per minute.",
			},
			"batch_1_day_max_input_tokens": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of input tokens per day for batch processing.",
			},
			"max_audio_megabytes_per_1_minute": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum audio megabytes per minute.",
			},
			"max_requests_per_1_day": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of requests per day.",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "API key to use for this specific resource instead of the provider-level API key.",
			},
			"rate_limit_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the rate limit.",
			},
			"ignore_rate_limit_warning": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set to true to acknowledge that OpenAI rate limits cannot be truly deleted and will be reset to defaults on removal.",
			},
		},
	}
}

// resourceOpenAIRateLimitCreate creates a new rate limit for a model in an OpenAI project.
// It sets up limits for requests, tokens, and/or images per minute
// and calls the OpenAI API to apply these limits to the specified project and model.
func resourceOpenAIRateLimitCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Setup the client
	c, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %w", err))
	}

	// Get the model and project ID from the schema
	model := d.Get("model").(string)
	projectID := d.Get("project_id").(string)

	tflog.Info(ctx, fmt.Sprintf("Creating rate limit for model %s in project %s", model, projectID))

	// Generate a consistent ID based on project and model
	projectSuffix := ""
	if len(projectID) > 8 {
		projectSuffix = projectID[len(projectID)-8:]
	} else {
		projectSuffix = projectID
	}
	rateLimitID := fmt.Sprintf("rl-%s-%s", model, projectSuffix)

	// Set the ID in our state right away
	d.SetId(rateLimitID)
	d.Set("rate_limit_id", rateLimitID)

	// Get the API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
		tflog.Debug(ctx, "Using resource-specific API key")
	}

	// Prepare parameters for updating the rate limit
	var maxRequestsPerMinute, maxTokensPerMinute, maxImagesPerMinute,
		batch1DayMaxInputTokens, maxAudioMegabytesPer1Minute, maxRequestsPer1Day *int

	// Get values from configuration
	if v, ok := d.GetOk("max_requests_per_minute"); ok {
		requestsPerMin := v.(int)
		maxRequestsPerMinute = &requestsPerMin
		tflog.Debug(ctx, fmt.Sprintf("Setting max_requests_per_minute to %d", requestsPerMin))
	}
	if v, ok := d.GetOk("max_tokens_per_minute"); ok {
		tokensPerMin := v.(int)
		maxTokensPerMinute = &tokensPerMin
		tflog.Debug(ctx, fmt.Sprintf("Setting max_tokens_per_minute to %d", tokensPerMin))
	}
	if v, ok := d.GetOk("max_images_per_minute"); ok {
		imagesPerMin := v.(int)
		maxImagesPerMinute = &imagesPerMin
		tflog.Debug(ctx, fmt.Sprintf("Setting max_images_per_minute to %d", imagesPerMin))
	}
	if v, ok := d.GetOk("batch_1_day_max_input_tokens"); ok {
		batchTokens := v.(int)
		batch1DayMaxInputTokens = &batchTokens
		tflog.Debug(ctx, fmt.Sprintf("Setting batch_1_day_max_input_tokens to %d", batchTokens))
	}
	if v, ok := d.GetOk("max_audio_megabytes_per_1_minute"); ok {
		audioMB := v.(int)
		maxAudioMegabytesPer1Minute = &audioMB
		tflog.Debug(ctx, fmt.Sprintf("Setting max_audio_megabytes_per_1_minute to %d", audioMB))
	}
	if v, ok := d.GetOk("max_requests_per_1_day"); ok {
		reqPerDay := v.(int)
		maxRequestsPer1Day = &reqPerDay
		tflog.Debug(ctx, fmt.Sprintf("Setting max_requests_per_1_day to %d", reqPerDay))
	}

	// Update the rate limit (this is effectively the create operation)
	tflog.Debug(ctx, fmt.Sprintf("Calling API to create/update rate limit for model %s", model))

	_, updateErr := c.UpdateRateLimitWithKey(
		projectID,
		model,
		maxRequestsPerMinute,
		maxTokensPerMinute,
		maxImagesPerMinute,
		batch1DayMaxInputTokens,
		maxAudioMegabytesPer1Minute,
		maxRequestsPer1Day,
		apiKey,
	)

	if updateErr != nil {
		tflog.Error(ctx, fmt.Sprintf("Error updating rate limit via API: %v", updateErr))

		// Special handling for permission errors
		if strings.Contains(updateErr.Error(), "permission") ||
			strings.Contains(updateErr.Error(), "403") ||
			strings.Contains(updateErr.Error(), "insufficient permissions") {
			// For permission errors, add a warning but preserve the resource in state with config values
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Permission error creating rate limit",
				Detail:   fmt.Sprintf("API error: %s. The resource will be created in Terraform state, but the actual settings in OpenAI may not match.", updateErr),
			})

			// Keep the resource in state with configuration values
			return diags
		}

		// For other errors, try to read existing values
		rateLimit, readErr := c.GetRateLimitWithKey(projectID, model, apiKey)
		if readErr != nil {
			// If both update and read fail, return the update error
			return diag.FromErr(fmt.Errorf("failed to create rate limit: %w", updateErr))
		}

		// If read succeeds, use those values and add a warning
		if rateLimit != nil {
			// Update state with existing values
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

			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Error updating rate limit",
				Detail:   fmt.Sprintf("API error: %s. Using existing rate limit values from API.", updateErr),
			})

			return diags
		}

		// If both fail, return the update error
		return diag.FromErr(fmt.Errorf("failed to create rate limit: %w", updateErr))
	}

	// Creation was successful, now read back the values to ensure state consistency
	rateLimit, readErr := c.GetRateLimitWithKey(projectID, model, apiKey)
	if readErr != nil {
		// If we can't read, use config values and add a warning
		tflog.Warn(ctx, fmt.Sprintf("Created rate limit but couldn't read it back: %v", readErr))
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Created rate limit but couldn't read it back",
			Detail:   fmt.Sprintf("Error reading back created rate limit: %s. Using configuration values.", readErr),
		})

		return diags
	}

	// Successfully read the rate limit back, update state
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

	tflog.Info(ctx, fmt.Sprintf("Successfully created rate limit for model %s with ID %s", model, rateLimit.ID))
	return diags
}

// resourceOpenAIRateLimitRead retrieves the current state of a rate limit from the OpenAI API.
// It reads the rate limit details including the requests, tokens, and images per minute values,
// and updates the Terraform state accordingly.
func resourceOpenAIRateLimitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %w", err))
	}

	// The ID in Terraform may have a suffix, but we save it to maintain consistency
	rateLimitID := d.Id()
	projectID := d.Get("project_id").(string)
	model := d.Get("model").(string)

	// Enhanced debug logging
	fmt.Printf("\n\n[TF-READ-DEBUG] ========== TERRAFORM READ RATE LIMIT DEBUG ==========\n")
	fmt.Printf("[TF-READ-DEBUG] Resource ID from Terraform state (d.Id()): %s\n", rateLimitID)
	fmt.Printf("[TF-READ-DEBUG] Project ID: %s\n", projectID)
	fmt.Printf("[TF-READ-DEBUG] Model from config: %s\n", model)

	if rateLimitID == "" {
		d.SetId("")
		fmt.Printf("[TF-READ-DEBUG] Rate limit ID is empty, clearing resource from state\n")
		fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
		return diag.Diagnostics{}
	}

	// Use the project-specific API key if provided
	apiKey := ""
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
		fmt.Printf("[TF-READ-DEBUG] Using project-specific API key for reading rate limit\n")
	} else {
		fmt.Printf("[TF-READ-DEBUG] Using default API key from provider configuration\n")
	}

	// IMPORTANT: First try with just the model name, as the OpenAI API actually uses model names
	// not the IDs with project suffixes that we generate for Terraform state management
	fmt.Printf("[TF-READ-DEBUG] First trying GetRateLimitWithKey with model name: %s\n", model)
	rateLimit, err := c.GetRateLimitWithKey(projectID, model, apiKey)

	// If we can't find it with the model name, try with the full ID as fallback
	if err != nil {
		fmt.Printf("[TF-READ-DEBUG] Failed to get rate limit with model name: %v\n", err)
		fmt.Printf("[TF-READ-DEBUG] Now trying with full rate limit ID: %s\n", rateLimitID)

		rateLimit, err = c.GetRateLimitWithKey(projectID, rateLimitID, apiKey)
		if err != nil {
			// Log the complete error for debugging
			fmt.Printf("[TF-READ-DEBUG] Also failed with full ID: %v\n", err)
			fmt.Printf("[TF-READ-DEBUG] Error type: %T\n", err)

			// Handle various error cases
			if responseHasStatusCode(err, 404) || strings.Contains(err.Error(), "not found") {
				fmt.Printf("[TF-READ-DEBUG] Resource not found, removing from Terraform state\n")
				d.SetId("")
				fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
				return nil
			} else if strings.Contains(err.Error(), "No project found") {
				fmt.Printf("[TF-READ-DEBUG] Project not found, removing from Terraform state: %s\n", projectID)
				d.SetId("")
				fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
				return nil
			} else if strings.Contains(err.Error(), "insufficient permissions") || strings.Contains(err.Error(), "do not have permission") {
				// Always create a warning, never an error for permission issues
				warning := diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Permission error reading rate limit",
					Detail:   fmt.Sprintf("Permission error: %s. The resource will remain in the Terraform state, but the actual values might differ from what's shown.", err),
				}

				// Just log the warning and continue
				tflog.Warn(ctx, fmt.Sprintf("Permission error when reading rate limit: %v. Continuing with Terraform operation.", err))
				fmt.Printf("[TF-READ-DEBUG] Permission error, returning warning\n")
				fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")

				// Always return only the warning, never an error for permission issues
				return diag.Diagnostics{warning}
			} else {
				// For any other error, return it to Terraform so it can handle drift properly
				fmt.Printf("[TF-READ-DEBUG] Unhandled error, returning to Terraform\n")
				fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
				return diag.Errorf("Error reading rate limit from OpenAI API: %s", err)
			}
		}
	}

	// If we got here, we successfully found the rate limit
	fmt.Printf("[TF-READ-DEBUG] Successfully read rate limit: ID=%s, Model=%s, MaxReq=%d, MaxTokens=%d\n",
		rateLimit.ID, rateLimit.Model, rateLimit.MaxRequestsPer1Minute, rateLimit.MaxTokensPer1Minute)

	if err := d.Set("model", rateLimit.Model); err != nil {
		fmt.Printf("[TF-READ-DEBUG] Error setting model: %v\n", err)
		fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
		return diag.FromErr(err)
	}
	if err := d.Set("max_requests_per_minute", rateLimit.MaxRequestsPer1Minute); err != nil {
		fmt.Printf("[TF-READ-DEBUG] Error setting max_requests_per_minute: %v\n", err)
		fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
		return diag.FromErr(err)
	}
	if err := d.Set("max_tokens_per_minute", rateLimit.MaxTokensPer1Minute); err != nil {
		fmt.Printf("[TF-READ-DEBUG] Error setting max_tokens_per_minute: %v\n", err)
		fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
		return diag.FromErr(err)
	}
	if err := d.Set("max_images_per_minute", rateLimit.MaxImagesPer1Minute); err != nil {
		fmt.Printf("[TF-READ-DEBUG] Error setting max_images_per_minute: %v\n", err)
		fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
		return diag.FromErr(err)
	}

	// For fields that may not be returned by the API for certain models,
	// only update the state if the API returned a non-zero value

	// If batch_1_day_max_input_tokens is 0 in the API response but non-zero in our state,
	// keep the existing state value
	oldBatch1DayMaxInputTokens := d.Get("batch_1_day_max_input_tokens").(int)
	if rateLimit.Batch1DayMaxInputTokens > 0 || oldBatch1DayMaxInputTokens == 0 {
		if err := d.Set("batch_1_day_max_input_tokens", rateLimit.Batch1DayMaxInputTokens); err != nil {
			fmt.Printf("[TF-READ-DEBUG] Error setting batch_1_day_max_input_tokens: %v\n", err)
			fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
			return diag.FromErr(err)
		}
	}

	// If max_audio_megabytes_per_1_minute is 0 in the API response but non-zero in our state,
	// keep the existing state value
	oldMaxAudioMegabytesPer1Minute := d.Get("max_audio_megabytes_per_1_minute").(int)
	if rateLimit.MaxAudioMegabytesPer1Minute > 0 || oldMaxAudioMegabytesPer1Minute == 0 {
		if err := d.Set("max_audio_megabytes_per_1_minute", rateLimit.MaxAudioMegabytesPer1Minute); err != nil {
			fmt.Printf("[TF-READ-DEBUG] Error setting max_audio_megabytes_per_1_minute: %v\n", err)
			fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
			return diag.FromErr(err)
		}
	}

	// If max_requests_per_1_day is 0 in the API response but non-zero in our state,
	// keep the existing state value
	oldMaxRequestsPer1Day := d.Get("max_requests_per_1_day").(int)
	if rateLimit.MaxRequestsPer1Day > 0 || oldMaxRequestsPer1Day == 0 {
		if err := d.Set("max_requests_per_1_day", rateLimit.MaxRequestsPer1Day); err != nil {
			fmt.Printf("[TF-READ-DEBUG] Error setting max_requests_per_1_day: %v\n", err)
			fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
			return diag.FromErr(err)
		}
	} else {
		fmt.Printf("[TF-READ-DEBUG] Preserving max_requests_per_1_day value %d in state (API returned 0)\n",
			oldMaxRequestsPer1Day)
	}

	if err := d.Set("rate_limit_id", rateLimit.ID); err != nil {
		fmt.Printf("[TF-READ-DEBUG] Error setting rate_limit_id: %v\n", err)
		fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")
		return diag.FromErr(fmt.Errorf("failed to set rate_limit_id: %v", err))
	}

	// Keep the current resource ID
	d.SetId(rateLimit.ID)
	fmt.Printf("[TF-READ-DEBUG] Set resource ID to: %s\n", rateLimit.ID)

	// Log the values for debugging
	fmt.Printf("[TF-READ-DEBUG] Rate limit values being set in Terraform state:\n")
	fmt.Printf("  - max_requests_per_minute: %d\n", rateLimit.MaxRequestsPer1Minute)
	fmt.Printf("  - max_tokens_per_minute: %d\n", rateLimit.MaxTokensPer1Minute)
	fmt.Printf("  - max_images_per_minute: %d\n", rateLimit.MaxImagesPer1Minute)
	fmt.Printf("  - batch_1_day_max_input_tokens: %d (preserved: %t)\n",
		rateLimit.Batch1DayMaxInputTokens, rateLimit.Batch1DayMaxInputTokens == 0 && oldBatch1DayMaxInputTokens > 0)
	fmt.Printf("  - max_audio_megabytes_per_1_minute: %d (preserved: %t)\n",
		rateLimit.MaxAudioMegabytesPer1Minute, rateLimit.MaxAudioMegabytesPer1Minute == 0 && oldMaxAudioMegabytesPer1Minute > 0)
	fmt.Printf("  - max_requests_per_1_day: %d (preserved: %t)\n",
		rateLimit.MaxRequestsPer1Day, rateLimit.MaxRequestsPer1Day == 0 && oldMaxRequestsPer1Day > 0)
	fmt.Printf("[TF-READ-DEBUG] ========== END TERRAFORM READ RATE LIMIT DEBUG ==========\n\n")

	return diag.Diagnostics{}
}

// resourceOpenAIRateLimitUpdate updates an OpenAI rate limit resource.
func resourceOpenAIRateLimitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the API client
	c, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %w", err))
	}

	// Get the API key for this resource
	var apiKey string
	if v, ok := d.GetOk("api_key"); ok {
		apiKey = v.(string)
	}

	// Get resource parameters
	projectID := d.Get("project_id").(string)
	model := d.Get("model").(string)
	rateLimitID := d.Id() // Use the ID directly

	tflog.Info(ctx, fmt.Sprintf("[IMPORTANT] OpenAI API uses fields with '_per_1_minute' (with _1_), while Terraform uses '_per_minute'"))
	tflog.Info(ctx, fmt.Sprintf("[IMPORTANT] Current rate limits from API for model %s will be shown in debug logs", model))

	// Read current rate limits from API for comparison
	currentRateLimit, err := c.GetRateLimitWithKey(projectID, model, apiKey)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf("Current API rate limits for %s: MaxReq=%d, MaxTokens=%d, MaxImages=%d",
			model, currentRateLimit.MaxRequestsPer1Minute, currentRateLimit.MaxTokensPer1Minute, currentRateLimit.MaxImagesPer1Minute))
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Could not read current rate limits before update: %v", err))
	}

	// Prepare parameters for updating the rate limit
	var maxRequestsPerMinute, maxTokensPerMinute, maxImagesPerMinute,
		batch1DayMaxInputTokens, maxAudioMegabytesPer1Minute, maxRequestsPer1Day *int

	// Use all parameters in the configuration, not just changed ones
	if v, ok := d.GetOk("max_requests_per_minute"); ok {
		requestsPerMin := v.(int)
		maxRequestsPerMinute = &requestsPerMin
		tflog.Info(ctx, fmt.Sprintf("[IMPORTANT] Setting max_requests_per_1_minute in API call to: %d (from max_requests_per_minute in config)", requestsPerMin))

		// Add comparison with existing value if available
		if currentRateLimit != nil && currentRateLimit.MaxRequestsPer1Minute != requestsPerMin {
			tflog.Info(ctx, fmt.Sprintf("[CHANGE] Updating max_requests_per_minute from %d to %d",
				currentRateLimit.MaxRequestsPer1Minute, requestsPerMin))
		}
	}

	if v, ok := d.GetOk("max_tokens_per_minute"); ok {
		tokensPerMin := v.(int)
		maxTokensPerMinute = &tokensPerMin
		tflog.Info(ctx, fmt.Sprintf("[IMPORTANT] Setting max_tokens_per_1_minute in API call to: %d (from max_tokens_per_minute in config)", tokensPerMin))

		// Add comparison with existing value if available
		if currentRateLimit != nil && currentRateLimit.MaxTokensPer1Minute != tokensPerMin {
			tflog.Info(ctx, fmt.Sprintf("[CHANGE] Updating max_tokens_per_minute from %d to %d",
				currentRateLimit.MaxTokensPer1Minute, tokensPerMin))
		}
	}

	if v, ok := d.GetOk("max_images_per_minute"); ok {
		imagesPerMin := v.(int)
		maxImagesPerMinute = &imagesPerMin
		tflog.Info(ctx, fmt.Sprintf("[IMPORTANT] Setting max_images_per_1_minute in API call to: %d (from max_images_per_minute in config)", imagesPerMin))

		// Add comparison with existing value if available
		if currentRateLimit != nil && currentRateLimit.MaxImagesPer1Minute != imagesPerMin {
			tflog.Info(ctx, fmt.Sprintf("[CHANGE] Updating max_images_per_minute from %d to %d",
				currentRateLimit.MaxImagesPer1Minute, imagesPerMin))
		}
	}

	if v, ok := d.GetOk("batch_1_day_max_input_tokens"); ok {
		batchTokens := v.(int)
		batch1DayMaxInputTokens = &batchTokens
		tflog.Debug(ctx, fmt.Sprintf("Setting batch_1_day_max_input_tokens to %d", batchTokens))
	}

	if v, ok := d.GetOk("max_audio_megabytes_per_1_minute"); ok {
		audioMB := v.(int)
		maxAudioMegabytesPer1Minute = &audioMB
		tflog.Debug(ctx, fmt.Sprintf("Setting max_audio_megabytes_per_1_minute to %d", audioMB))
	}

	if v, ok := d.GetOk("max_requests_per_1_day"); ok {
		reqPerDay := v.(int)
		maxRequestsPer1Day = &reqPerDay
		tflog.Debug(ctx, fmt.Sprintf("Setting max_requests_per_1_day to %d", reqPerDay))
	}

	// Update the rate limit using the API
	tflog.Debug(ctx, fmt.Sprintf("Updating rate limit for project %s, model %s with ID %s",
		projectID, model, rateLimitID))

	// Use model when passing to API functions, not the rate limit ID
	rateLimit, err := c.UpdateRateLimitWithKey(
		projectID,
		model,
		maxRequestsPerMinute,
		maxTokensPerMinute,
		maxImagesPerMinute,
		batch1DayMaxInputTokens,
		maxAudioMegabytesPer1Minute,
		maxRequestsPer1Day,
		apiKey,
	)

	if err != nil {
		// Log significant details for debugging
		tflog.Error(ctx, fmt.Sprintf("Failed to update rate limit: %v", err))

		// Check if error is due to organization limits
		if strings.Contains(err.Error(), "cannot exceed the organization rate limit") {
			tflog.Warn(ctx, fmt.Sprintf("Attempted to set rate limits higher than organization allows: %s", err.Error()))

			// Read the current state to get the actual values imposed by the API
			readDiags := resourceOpenAIRateLimitRead(ctx, d, meta)
			if readDiags.HasError() {
				return diag.FromErr(fmt.Errorf("error updating rate limit (organization limit exceeded) and failed to read current state: %w", err))
			}

			// Add a warning diagnostic but don't fail the apply
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Rate limit values adjusted to organization limits",
					Detail:   fmt.Sprintf("Attempted values exceed organization limits: %s. State has been updated with actual values.", err.Error()),
				},
			}
		} else if strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "insufficient permissions") {
			// Handle permission errors
			tflog.Error(ctx, "Permission error updating rate limit. Make sure you're using an admin API key with proper permissions.")

			// Instead of failing with an error, create a warning and continue
			warning := diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Permission error updating rate limit",
				Detail:   "OpenAI API returned a permission error. Your API key might not have permission to modify rate limits. The resource has been created in Terraform state, but the actual rate limit settings may differ from what you specified.",
			}

			// Create a placeholder rateLimit with the values from the configuration
			// so we can proceed with setting the state rather than failing
			rateLimit := &client.RateLimit{
				ID:    rateLimitID,
				Model: model,
			}

			// Set values from configuration to use in state
			if maxRequestsPerMinute != nil {
				rateLimit.MaxRequestsPer1Minute = *maxRequestsPerMinute
			}
			if maxTokensPerMinute != nil {
				rateLimit.MaxTokensPer1Minute = *maxTokensPerMinute
			}
			if maxImagesPerMinute != nil {
				rateLimit.MaxImagesPer1Minute = *maxImagesPerMinute
			}
			if batch1DayMaxInputTokens != nil {
				rateLimit.Batch1DayMaxInputTokens = *batch1DayMaxInputTokens
			}
			if maxAudioMegabytesPer1Minute != nil {
				rateLimit.MaxAudioMegabytesPer1Minute = *maxAudioMegabytesPer1Minute
			}
			if maxRequestsPer1Day != nil {
				rateLimit.MaxRequestsPer1Day = *maxRequestsPer1Day
			}

			// Set the values in Terraform state
			setRateLimitState(d, rateLimit)

			// Return only the warning, not an error
			return diag.Diagnostics{warning}
		}

		return diag.FromErr(fmt.Errorf("error updating rate limit: %w", err))
	}

	// Read the rate limit again to verify changes were applied
	updatedRateLimit, err := c.GetRateLimitWithKey(projectID, model, apiKey)
	if err == nil {
		tflog.Info(ctx, fmt.Sprintf("Rate limit after update API call for %s: MaxReq=%d, MaxTokens=%d, MaxImages=%d",
			model, updatedRateLimit.MaxRequestsPer1Minute, updatedRateLimit.MaxTokensPer1Minute, updatedRateLimit.MaxImagesPer1Minute))

		// Verify if the update worked
		if maxRequestsPerMinute != nil && updatedRateLimit.MaxRequestsPer1Minute != *maxRequestsPerMinute {
			tflog.Warn(ctx, fmt.Sprintf("[WARNING] API didn't update max_requests_per_minute to requested value! Requested: %d, Actual: %d",
				*maxRequestsPerMinute, updatedRateLimit.MaxRequestsPer1Minute))

			// Create a warning diagnostic for the user
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Rate limit not updated as expected",
					Detail: fmt.Sprintf("OpenAI API did not update the rate limit as requested. This could be due to organization-level restrictions. Requested: %d, Actual: %d",
						*maxRequestsPerMinute, updatedRateLimit.MaxRequestsPer1Minute),
				},
			}
		}
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Could not verify rate limit update: %v", err))
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully updated rate limit for model %s in project %s",
		model, projectID))

	// Update the state with the response data
	d.SetId(rateLimit.ID)

	// Log the values we received from the API
	tflog.Debug(ctx, fmt.Sprintf("Rate limit values from API response:\n"+
		"  - ID: %s\n"+
		"  - Model: %s\n"+
		"  - max_requests_per_minute: %d\n"+
		"  - max_tokens_per_minute: %d\n"+
		"  - max_images_per_minute: %d\n"+
		"  - batch_1_day_max_input_tokens: %d\n"+
		"  - max_audio_megabytes_per_1_minute: %d\n"+
		"  - max_requests_per_1_day: %d",
		rateLimit.ID, rateLimit.Model,
		rateLimit.MaxRequestsPer1Minute, rateLimit.MaxTokensPer1Minute,
		rateLimit.MaxImagesPer1Minute, rateLimit.Batch1DayMaxInputTokens,
		rateLimit.MaxAudioMegabytesPer1Minute, rateLimit.MaxRequestsPer1Day))

	// Use our helper function to set all values from the response back to state
	setRateLimitState(d, rateLimit)

	// FIXING: Always perform a read after update to ensure state consistency
	readDiags := resourceOpenAIRateLimitRead(ctx, d, meta)
	if readDiags.HasError() {
		tflog.Warn(ctx, fmt.Sprintf("Resource was updated but state couldn't be refreshed: %v", readDiags))
		// Merge the warnings but don't return errors as the resource was updated
		for _, diagnostic := range readDiags {
			if diagnostic.Severity == diag.Warning {
				diags = append(diags, diagnostic)
			}
		}
	}

	return diags
}

// Helper function to set rate limit values in state
func setRateLimitState(d *schema.ResourceData, rateLimit *client.RateLimit) {
	// Set all values from the rate limit object to state
	d.Set("model", rateLimit.Model)
	d.Set("max_requests_per_minute", rateLimit.MaxRequestsPer1Minute)
	d.Set("max_tokens_per_minute", rateLimit.MaxTokensPer1Minute)
	d.Set("max_images_per_minute", rateLimit.MaxImagesPer1Minute)
	d.Set("batch_1_day_max_input_tokens", rateLimit.Batch1DayMaxInputTokens)
	d.Set("max_audio_megabytes_per_1_minute", rateLimit.MaxAudioMegabytesPer1Minute)
	d.Set("max_requests_per_1_day", rateLimit.MaxRequestsPer1Day)

	// Keep the current ID
	if rateLimit.ID != "" {
		d.SetId(rateLimit.ID)
	}
}

// Helper function to determine API key type
func getAPIKeyType(apiKey string) string {
	if strings.HasPrefix(apiKey, "sk-proj-") {
		return "project API key"
	} else if strings.HasPrefix(apiKey, "sk-") {
		return "organization API key"
	}
	return "unknown key type"
}

// resourceOpenAIRateLimitDelete handles deletion of a rate limit by resetting it to default values.
// Since OpenAI API doesn't support true deletion of rate limits, this function resets the rate limit
// to the default values based on the comprehensive model defaults we have compiled.
func resourceOpenAIRateLimitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "START: Deleting OpenAI rate limit resource")
	fmt.Printf("\n\n[TF-DELETE-DEBUG] ========== TERRAFORM DELETE RATE LIMIT DEBUG ==========\n")
	fmt.Printf("[TF-DELETE-DEBUG] Resource ID: %s\n", d.Id())

	// Get the client from the provider configuration
	// Fix: Use GetOpenAIClient helper function instead of direct type assertion
	client, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting OpenAI client: %w", err))
	}
	fmt.Printf("[TF-DELETE-DEBUG] Client API URL: %s\n", client.APIURL)

	var diags diag.Diagnostics

	// Extract the model and project ID from the resource state
	model := d.Get("model").(string)
	projectID := d.Get("project_id").(string)
	fmt.Printf("[TF-DELETE-DEBUG] Model from state: %s\n", model)
	fmt.Printf("[TF-DELETE-DEBUG] Project ID from state: %s\n", projectID)

	// Get the rate limit ID directly from the resource ID
	resourceID := d.Id()
	fmt.Printf("[TF-DELETE-DEBUG] Resource ID (full): %s\n", resourceID)

	// Get the API key if provided
	apiKey := d.Get("api_key").(string)
	if apiKey == "" {
		apiKey = client.APIKey
		fmt.Printf("[TF-DELETE-DEBUG] No custom API key provided, using client default\n")
	} else {
		maskedKey := "*****"
		if len(apiKey) > 5 {
			maskedKey = apiKey[:5] + "*****"
		}
		fmt.Printf("[TF-DELETE-DEBUG] Using custom API key (masked): %s\n", maskedKey)
	}

	// Try to get the existing rate limit before deleting
	fmt.Printf("[TF-DELETE-DEBUG] Checking if rate limit exists before deletion\n")
	existingRateLimit, getRLErr := client.GetRateLimitWithKey(projectID, model, apiKey)
	if getRLErr != nil {
		fmt.Printf("[TF-DELETE-DEBUG] Error getting existing rate limit: %v\n", getRLErr)
		// Try with the full ID as fallback
		existingRateLimit, getRLErr = client.GetRateLimitWithKey(projectID, resourceID, apiKey)
		if getRLErr != nil {
			fmt.Printf("[TF-DELETE-DEBUG] Also failed with full ID: %v\n", getRLErr)
		}
	}

	if existingRateLimit != nil {
		fmt.Printf("[TF-DELETE-DEBUG] Found existing rate limit: ID=%s, Model=%s, MaxReq=%d, MaxTokens=%d\n",
			existingRateLimit.ID, existingRateLimit.Model, existingRateLimit.MaxRequestsPer1Minute, existingRateLimit.MaxTokensPer1Minute)
	} else {
		fmt.Printf("[TF-DELETE-DEBUG] Rate limit not found or nil response\n")
	}

	// Attempt to delete the rate limit - Use the model name since that's what the API expects
	fmt.Printf("[TF-DELETE-DEBUG] Calling DeleteRateLimitWithKey with: projectID=%s, model=%s\n",
		projectID, model)

	err = client.DeleteRateLimitWithKey(projectID, model, apiKey)
	if err != nil {
		fmt.Printf("[TF-DELETE-DEBUG] Error deleting rate limit: %v\n", err)
		fmt.Printf("[TF-DELETE-DEBUG] Error type: %T\n", err)
		fmt.Printf("[TF-DELETE-DEBUG] Error contains 'permission': %v\n", strings.Contains(err.Error(), "permission"))
		fmt.Printf("[TF-DELETE-DEBUG] Error contains '403': %v\n", strings.Contains(err.Error(), "403"))

		// Try with the full ID as fallback
		fmt.Printf("[TF-DELETE-DEBUG] Trying again with full resource ID: %s\n", resourceID)
		err = client.DeleteRateLimitWithKey(projectID, resourceID, apiKey)
		if err != nil {
			fmt.Printf("[TF-DELETE-DEBUG] Also failed with full ID: %v\n", err)

			if strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "403") {
				// Special handling for permission errors
				fmt.Printf("[TF-DELETE-DEBUG] Got permission/403 error, checking if we should handle differently\n")

				// Check if the error is specifically a permission error for deletion
				if strings.Contains(err.Error(), "permission to delete") {
					// Get the rate limit again to see if it still exists with original values
					fmt.Printf("[TF-DELETE-DEBUG] Checking if rate limit still exists after deletion attempt\n")
					rateLimit, getRLErr := client.GetRateLimitWithKey(projectID, model, apiKey)
					if getRLErr == nil && rateLimit != nil {
						// The rate limit still exists, but we can consider it "deleted" from Terraform's perspective
						fmt.Printf("[TF-DELETE-DEBUG] Rate limit still exists but marking as deleted in Terraform state\n")
						fmt.Printf("[TF-DELETE-DEBUG] ========== END TERRAFORM DELETE RATE LIMIT DEBUG ==========\n\n")
						return diags
					}
				}
			}

			// Convert the error to diagnostics
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error deleting OpenAI rate limit",
				Detail:   fmt.Sprintf("Error deleting OpenAI rate limit: %s", err),
			})
			fmt.Printf("[TF-DELETE-DEBUG] Added error diagnostic\n")
			fmt.Printf("[TF-DELETE-DEBUG] ========== END TERRAFORM DELETE RATE LIMIT DEBUG ==========\n\n")
			return diags
		}
	}

	fmt.Printf("[TF-DELETE-DEBUG] Rate limit deletion successful\n")

	// Verify if rate limit was actually deleted/reset by checking the API again
	fmt.Printf("[TF-DELETE-DEBUG] Verifying rate limit state after deletion\n")
	verifyRateLimit, verifyErr := client.GetRateLimitWithKey(projectID, model, apiKey)
	if verifyErr != nil {
		fmt.Printf("[TF-DELETE-DEBUG] Error verifying rate limit state: %v\n", verifyErr)
	} else if verifyRateLimit != nil {
		fmt.Printf("[TF-DELETE-DEBUG] Rate limit still exists after deletion (expected, since it's reset to defaults)\n")
		fmt.Printf("[TF-DELETE-DEBUG] Updated rate limit details: ID=%s, Model=%s, MaxReq=%d, MaxTokens=%d\n",
			verifyRateLimit.ID, verifyRateLimit.Model, verifyRateLimit.MaxRequestsPer1Minute, verifyRateLimit.MaxTokensPer1Minute)
	} else {
		fmt.Printf("[TF-DELETE-DEBUG] Rate limit not found after deletion (unexpected)\n")
	}

	fmt.Printf("[TF-DELETE-DEBUG] ========== END TERRAFORM DELETE RATE LIMIT DEBUG ==========\n\n")
	tflog.Info(ctx, "END: Deleted OpenAI rate limit resource")
	return diags
}

// Helper function to get min of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// responseHasStatusCode is a helper function to check if an error response
// contains a specific HTTP status code. This is useful for handling 404 Not Found
// errors when a resource doesn't exist.
func responseHasStatusCode(err error, code int) bool {
	if err == nil {
		return false
	}

	// Check for various formats of status code in error message
	errorString := fmt.Sprintf("%v", err)
	statusPatterns := []string{
		fmt.Sprintf("status: %d", code),
		fmt.Sprintf("status code: %d", code),
		fmt.Sprintf("status=%d", code),
		fmt.Sprintf("statusCode=%d", code),
		fmt.Sprintf("status_code=%d", code),
		fmt.Sprintf("status code %d", code),
	}

	for _, pattern := range statusPatterns {
		if strings.Contains(errorString, pattern) {
			return true
		}
	}

	return false
}

// resourceOpenAIRateLimitImport handles importing an existing rate limit into Terraform state.
// It parses the given import ID (rate limit ID) and fetches the rate limit information
// from the OpenAI API to populate the Terraform state.
func resourceOpenAIRateLimitImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Get the OpenAI client
	c, err := GetOpenAIClient(meta)
	if err != nil {
		return nil, fmt.Errorf("error getting OpenAI client: %w", err)
	}

	// The ID should be in the format "project_id:rate_limit_id"
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return nil, fmt.Errorf("invalid import ID, expected format: project_id:rate_limit_id (e.g., proj_abc123:rl-gpt-4-abc123)")
	}

	projectID := idParts[0]
	rateLimitID := idParts[1]

	tflog.Debug(ctx, fmt.Sprintf("Importing rate limit with ID: %s for project: %s", rateLimitID, projectID))

	// Parse the rate limit ID to extract model info
	// Rate limit IDs typically follow the pattern: rl-[model]-[project-suffix]
	var model string
	if strings.HasPrefix(rateLimitID, "rl-") {
		parts := strings.Split(rateLimitID, "-")
		if len(parts) < 3 || parts[0] != "rl" {
			return nil, fmt.Errorf("invalid rate limit ID format, expected rl-model-projectsuffix, got: %s", rateLimitID)
		}

		// Extract the model name which could be multi-part (e.g., gpt-3.5-turbo)
		// The model is everything from the second part to the second-last part
		modelParts := parts[1 : len(parts)-1]
		model = strings.Join(modelParts, "-")
	} else {
		// If the ID doesn't start with "rl-", assume it's already a model name
		model = rateLimitID
	}

	tflog.Debug(ctx, fmt.Sprintf("Extracted model: %s from rate limit ID: %s", model, rateLimitID))

	// Fetch the rate limit data from the API
	// The GetRateLimitWithKey function will handle the URL formatting correctly
	// Use the full rate limit ID for exact matching
	rateLimit, err := c.GetRateLimitWithKey(projectID, rateLimitID, "")
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error retrieving rate limit from API: %v", err))
		return nil, fmt.Errorf("error retrieving rate limit from API: %w", err)
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully retrieved rate limit: ID=%s, Model=%s",
		rateLimit.ID, rateLimit.Model))

	// Set all resource data fields from the API response
	d.SetId(rateLimit.ID)

	// Set the model in the resource data (use the one from API, not parsed)
	if err := d.Set("model", rateLimit.Model); err != nil {
		return nil, fmt.Errorf("error setting model: %w", err)
	}

	// Set the project ID in the resource data
	if err := d.Set("project_id", projectID); err != nil {
		return nil, fmt.Errorf("error setting project_id: %w", err)
	}

	// Set the rate limit ID (the one without the project suffix)
	if err := d.Set("rate_limit_id", strings.TrimSuffix(rateLimit.ID,
		"-"+strings.Split(rateLimit.ID, "-")[len(strings.Split(rateLimit.ID, "-"))-1])); err != nil {
		return nil, fmt.Errorf("error setting rate_limit_id: %w", err)
	}

	// Set all the other fields from the API response
	if err := d.Set("max_requests_per_minute", rateLimit.MaxRequestsPer1Minute); err != nil {
		return nil, fmt.Errorf("error setting max_requests_per_minute: %w", err)
	}

	if err := d.Set("max_tokens_per_minute", rateLimit.MaxTokensPer1Minute); err != nil {
		return nil, fmt.Errorf("error setting max_tokens_per_minute: %w", err)
	}

	if err := d.Set("max_images_per_minute", rateLimit.MaxImagesPer1Minute); err != nil {
		return nil, fmt.Errorf("error setting max_images_per_minute: %w", err)
	}

	if err := d.Set("batch_1_day_max_input_tokens", rateLimit.Batch1DayMaxInputTokens); err != nil {
		return nil, fmt.Errorf("error setting batch_1_day_max_input_tokens: %w", err)
	}

	if err := d.Set("max_audio_megabytes_per_1_minute", rateLimit.MaxAudioMegabytesPer1Minute); err != nil {
		return nil, fmt.Errorf("error setting max_audio_megabytes_per_1_minute: %w", err)
	}

	if err := d.Set("max_requests_per_1_day", rateLimit.MaxRequestsPer1Day); err != nil {
		return nil, fmt.Errorf("error setting max_requests_per_1_day: %w", err)
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully imported rate limit %s for model %s in project %s. Values: MaxReq=%d, MaxTokens=%d, MaxImg=%d",
		rateLimit.ID, rateLimit.Model, projectID,
		rateLimit.MaxRequestsPer1Minute, rateLimit.MaxTokensPer1Minute, rateLimit.MaxImagesPer1Minute))

	return []*schema.ResourceData{d}, nil
}
