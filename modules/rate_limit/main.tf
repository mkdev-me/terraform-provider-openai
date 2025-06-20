# OpenAI Rate Limit Module

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "~> 1.0.0"
    }
  }
}

# If list_mode is true, use the openai_rate_limits data source to get all rate limits
# This will fail unless an admin API key with api.management.read scope is provided
data "openai_rate_limits" "all_limits" {
  count      = var.list_mode ? 1 : 0
  project_id = var.project_id
  api_key    = var.openai_admin_key
}

# If use_data_source is true but list_mode is false, use the openai_rate_limit data source
# This will fail unless an admin API key with api.management.read scope is provided
data "openai_rate_limit" "rate_limit" {
  count      = (!var.list_mode && var.use_data_source) ? 1 : 0
  project_id = var.project_id
  model      = var.model
  api_key    = var.openai_admin_key
}

# Only create the resource if both list_mode and use_data_source are false
resource "openai_rate_limit" "rate_limit" {
  count                            = (!var.list_mode && !var.use_data_source) ? 1 : 0
  project_id                       = var.project_id
  model                            = var.model
  max_requests_per_minute          = var.max_requests_per_minute
  max_tokens_per_minute            = var.max_tokens_per_minute
  max_images_per_minute            = var.max_images_per_minute
  batch_1_day_max_input_tokens     = var.batch_1_day_max_input_tokens
  max_audio_megabytes_per_1_minute = var.max_audio_megabytes_per_1_minute
  max_requests_per_1_day           = var.max_requests_per_1_day
  api_key                          = var.openai_admin_key
}

locals {
  # Safely check if list mode data source successfully retrieved data
  has_list_mode = var.list_mode && try(
    length(data.openai_rate_limits.all_limits) > 0 &&
    data.openai_rate_limits.all_limits[0].rate_limits != null,
    false
  )

  # Safely check if data source for single rate limit exists
  has_data_source = !var.list_mode && var.use_data_source && try(
    length(data.openai_rate_limit.rate_limit) > 0 &&
    data.openai_rate_limit.rate_limit[0].id != "",
    false
  )

  # Check if resource is available (count is 1)
  has_resource = !var.list_mode && !var.use_data_source && length(openai_rate_limit.rate_limit) > 0

  # Format the ID based on whether we're using the data source or resource
  rate_limit_id    = local.has_data_source ? try(data.openai_rate_limit.rate_limit[0].rate_limit_id, "unknown") : (local.has_resource ? openai_rate_limit.rate_limit[0].id : "unknown")
  model            = local.has_data_source ? try(data.openai_rate_limit.rate_limit[0].model, var.model) : (local.has_resource ? openai_rate_limit.rate_limit[0].model : var.model)
  requests_per_min = local.has_data_source ? try(data.openai_rate_limit.rate_limit[0].max_requests_per_minute, var.max_requests_per_minute) : (local.has_resource ? openai_rate_limit.rate_limit[0].max_requests_per_minute : var.max_requests_per_minute)
  tokens_per_min   = local.has_data_source ? try(data.openai_rate_limit.rate_limit[0].max_tokens_per_minute, var.max_tokens_per_minute) : (local.has_resource ? openai_rate_limit.rate_limit[0].max_tokens_per_minute : var.max_tokens_per_minute)
  images_per_min   = local.has_data_source ? try(data.openai_rate_limit.rate_limit[0].max_images_per_minute, var.max_images_per_minute) : (local.has_resource ? openai_rate_limit.rate_limit[0].max_images_per_minute : var.max_images_per_minute)
  batch_tokens     = local.has_data_source ? try(data.openai_rate_limit.rate_limit[0].batch_1_day_max_input_tokens, var.batch_1_day_max_input_tokens) : (local.has_resource ? openai_rate_limit.rate_limit[0].batch_1_day_max_input_tokens : var.batch_1_day_max_input_tokens)
  audio_megabytes  = local.has_data_source ? try(data.openai_rate_limit.rate_limit[0].max_audio_megabytes_per_1_minute, var.max_audio_megabytes_per_1_minute) : (local.has_resource ? openai_rate_limit.rate_limit[0].max_audio_megabytes_per_1_minute : var.max_audio_megabytes_per_1_minute)
  requests_per_day = local.has_data_source ? try(data.openai_rate_limit.rate_limit[0].max_requests_per_1_day, var.max_requests_per_1_day) : (local.has_resource ? openai_rate_limit.rate_limit[0].max_requests_per_1_day : var.max_requests_per_1_day)
  created_at       = timestamp()

  # Get all rate limits when in list mode
  all_rate_limits = local.has_list_mode ? try(data.openai_rate_limits.all_limits[0].rate_limits, []) : []
} 