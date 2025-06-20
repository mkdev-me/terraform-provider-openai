output "id" {
  description = "Terraform resource ID for the rate limit"
  value       = local.has_data_source ? try(data.openai_rate_limit.rate_limit[0].id, "unknown") : (local.has_resource ? openai_rate_limit.rate_limit[0].id : "unknown")
}

output "rate_limit_id" {
  description = "The ID of the rate limit"
  value       = local.rate_limit_id
}

output "project_id" {
  description = "Project ID associated with this rate limit"
  value       = var.project_id
}

output "model" {
  description = "The model the rate limit is for"
  value       = local.model
}

output "max_requests_per_minute" {
  description = "The maximum requests per minute"
  value       = local.requests_per_min
}

output "max_tokens_per_minute" {
  description = "The maximum tokens per minute"
  value       = local.tokens_per_min
}

output "max_images_per_minute" {
  description = "The maximum images per minute"
  value       = local.images_per_min
}

output "batch_1_day_max_input_tokens" {
  description = "The maximum input tokens for batch processing in a day"
  value       = local.batch_tokens
}

output "max_audio_megabytes_per_1_minute" {
  description = "The maximum audio megabytes per minute"
  value       = local.audio_megabytes
}

output "max_requests_per_1_day" {
  description = "The maximum requests per day"
  value       = local.requests_per_day
}

output "created_at" {
  description = "The date and time the rate limit was created"
  value       = local.created_at
}

output "all_rate_limits" {
  description = "All rate limits (only populated when list_mode = true)"
  value       = local.all_rate_limits
} 