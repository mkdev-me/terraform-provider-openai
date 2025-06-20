---
page_title: "OpenAI: openai_rate_limits"
subcategory: "Rate Limits"
description: |-
  Retrieves information about all rate limits for an OpenAI project.
---

# Data Source: openai_rate_limits

Use this data source to retrieve information about all rate limits for a specific OpenAI project.

## Example Usage

```hcl
data "openai_rate_limits" "all_limits" {
  project_id = "proj_abc123456789"
  
  # Optional: use a project-specific API key
  api_key    = var.project_api_key
}

output "all_rate_limits" {
  value = data.openai_rate_limits.all_limits.rate_limits
}

output "gpt4_limits" {
  value = [
    for limit in data.openai_rate_limits.all_limits.rate_limits:
    limit if limit.model == "gpt-4"
  ][0]
}
```

## Error Handling

This data source includes improved error handling for authentication and permission issues:

- Authentication errors (like "Invalid authorization" or missing scopes) will result in warnings rather than failing the entire Terraform run
- When such errors occur, the data source will return empty values and set a placeholder ID
- 404 errors (project not found) are also handled gracefully with informative warnings
- This allows you to use this data source alongside other resources even without the proper admin permissions

## Rate Limit Lifecycle Notes

While this data source only reads rate limit information, it's worth noting that when using the corresponding `openai_rate_limit` resource:

- Rate limits cannot be truly deleted via the OpenAI API
- When a rate limit resource is destroyed, it's reset to the model's default values
- The provider has been improved to handle this reset process reliably during `terraform destroy` operations
- Type conversion issues in the deletion process have been fixed to ensure consistent behavior without errors

## Argument Reference

* `project_id` - (Required) The ID of the project to retrieve rate limits for (format: `proj_abc123456789`).
* `api_key` - (Optional) A project-specific API key to use for authentication. If not provided, the provider's default API key will be used.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `rate_limits` - A list of rate limits for the project. Each rate limit contains the following attributes:
  * `model` - The model the rate limit applies to.
  * `rate_limit_id` - The ID of the rate limit.
  * `max_requests_per_minute` - Maximum number of API requests allowed per minute.
  * `max_tokens_per_minute` - Maximum number of tokens that can be processed per minute.
  * `max_images_per_minute` - Maximum number of images that can be generated per minute.
  * `batch_1_day_max_input_tokens` - Maximum number of input tokens allowed in batch operations per day.
  * `max_audio_megabytes_per_1_minute` - Maximum number of audio megabytes that can be processed per minute.
  * `max_requests_per_1_day` - Maximum number of API requests allowed per day.

## Permissions Required

This data source requires an OpenAI API key with appropriate permissions:
* Organization admin key with `api.management.read` scope or
* Project-specific API key with read access to the project 