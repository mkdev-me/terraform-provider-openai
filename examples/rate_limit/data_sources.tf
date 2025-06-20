# Example usage of openai_rate_limit and openai_rate_limits data sources
# NOTE: These data sources require an admin API key with api.management.read scope
# If you don't have these permissions, you will see warnings but the terraform run
# will still complete successfully due to error handling improvements.

# Define API key variable for data sources
variable "openai_api_key" {
  type        = string
  description = "OpenAI Admin API key (sk-admin-...)"
  sensitive   = true
  default     = null # Will be taken from OPENAI_API_KEY env var
}

# Optional: If you want to try the data sources and have the right permissions
variable "try_data_sources" {
  description = "Whether to try using data sources (requires admin API key with api.management.read permissions)"
  type        = bool
  default     = false
}


# Retrieve all rate limits for a project - using count to make it optional
data "openai_rate_limits" "all_limits" {
  count      = var.try_data_sources ? 1 : 0
  project_id = var.project_id
}

# Retrieve rate limit for the DALL-E 3 model specifically
data "openai_rate_limit" "dalle3_limit" {
  count      = var.try_data_sources ? 1 : 0
  project_id = var.project_id
  model      = "dall-e-3"
}

# Output example: get DALL-E 3 limits using the data source
output "dalle3_limits_from_datasource" {
  value = var.try_data_sources ? {
    model                   = data.openai_rate_limit.dalle3_limit[0].model
    max_requests_per_minute = data.openai_rate_limit.dalle3_limit[0].max_requests_per_minute
    max_images_per_minute   = data.openai_rate_limit.dalle3_limit[0].max_images_per_minute
    max_requests_per_1_day  = data.openai_rate_limit.dalle3_limit[0].max_requests_per_1_day
  } : null
  description = "Rate limits for DALL-E 3 from data source (only shown if try_data_sources = true)"
}

# Added output to show all rate limits from the data source
output "all_rate_limits_from_datasource" {
  value = var.try_data_sources ? {
    count  = length(data.openai_rate_limits.all_limits[0].rate_limits)
    models = [for limit in data.openai_rate_limits.all_limits[0].rate_limits : limit.model]
  } : null
  description = "All rate limits from data source (only shown if try_data_sources = true)"
}

/*
IMPROVED ERROR HANDLING

Both data sources (openai_rate_limit and openai_rate_limits) have been updated to be more resilient
to authentication and permission issues. Previously, if your API key didn't have the required
permissions, Terraform would fail the entire run. Now:

1. Authentication errors (Invalid authorization, missing scopes, etc.) are handled gracefully
   and converted to warnings rather than errors
   
2. When errors occur, the data sources will return empty or placeholder values instead of failing
   
3. This allows you to use these data sources alongside other resources even if you don't have
   the proper admin permissions
   
4. All types of errors (404, permissions, authentication) are now handled with more detailed
   logging and diagnostics
*/

