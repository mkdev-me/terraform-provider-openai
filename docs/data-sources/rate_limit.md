---
page_title: "OpenAI: openai_rate_limit"
subcategory: "Rate Limits"
description: |-
  Retrieves information about a specific rate limit for a model in an OpenAI project.
---

# Data Source: openai_rate_limit

Use this data source to retrieve information about a specific rate limit for a model in an OpenAI project.

## Example Usage

```hcl
data "openai_rate_limit" "gpt4_limits" {
  project_id = "proj_abc123456789"
  model      = "gpt-4"
  
  # Optional: use a project-specific API key
  api_key    = var.project_api_key
}

output "gpt4_max_requests" {
  value = data.openai_rate_limit.gpt4_limits.max_requests_per_minute
}

output "gpt4_max_tokens" {
  value = data.openai_rate_limit.gpt4_limits.max_tokens_per_minute
}
```

## Error Handling

This data source includes improved error handling for authentication and permission issues:

- Authentication errors (like "Invalid authorization" or missing scopes) will result in warnings rather than failing the entire Terraform run
- When such errors occur, the data source will return empty values and set a placeholder ID
- 404 errors (rate limit or model not found) are also handled gracefully with informative warnings
- This allows you to use this data source alongside other resources even without the proper admin permissions

For conditional usage, you can use Terraform's `count` parameter:

```hcl
variable "try_data_sources" {
  type    = bool
  default = false
}

data "openai_rate_limit" "gpt4_limits" {
  count      = var.try_data_sources ? 1 : 0
  project_id = "proj_abc123456789"
  model      = "gpt-4"
}

output "gpt4_max_requests" {
  value = var.try_data_sources ? data.openai_rate_limit.gpt4_limits[0].max_requests_per_minute : null
}
```

## Argument Reference

* `project_id` - (Required) The ID of the project to retrieve the rate limit for (format: `proj_abc123456789`).
* `model` - (Required) The model to retrieve the rate limit for (e.g., `gpt-4`, `gpt-3.5-turbo`).
* `api_key` - (Optional) A project-specific API key to use for authentication. If not provided, the provider's default API key will be used.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `max_requests_per_minute` - Maximum number of API requests allowed per minute.
* `max_tokens_per_minute` - Maximum number of tokens that can be processed per minute.
* `max_images_per_minute` - Maximum number of images that can be generated per minute.
* `batch_1_day_max_input_tokens` - Maximum number of input tokens allowed in batch operations per day.
* `max_audio_megabytes_per_1_minute` - Maximum number of audio megabytes that can be processed per minute.
* `max_requests_per_1_day` - Maximum number of API requests allowed per day.
* `rate_limit_id` - The OpenAI-assigned ID for this rate limit.

## Permissions Required

This data source requires an OpenAI API key with appropriate permissions:
* Organization admin key with `api.management.read` scope or
* Project-specific API key with read access to the project 