---
page_title: "OpenAI: openai_rate_limit Resource"
subcategory: "Rate Limits"
description: |-
  Manages rate limits for an OpenAI project and model combination.
---

# Resource: openai_rate_limit

This resource allows you to manage rate limits for OpenAI models within a project. Rate limits help control API usage and costs by setting caps on requests, tokens, images, and batch operations.

## Example Usage

```hcl
resource "openai_rate_limit" "gpt4_limits" {
  project_id              = "proj_abc123456789"
  model                   = "gpt-4"
  max_requests_per_minute = 100
  max_tokens_per_minute   = 10000
  
  # Optional parameters for specific model types
  max_images_per_minute        = 50
  batch_1_day_max_input_tokens = 1000000
  
  # If using a project-specific API key
  api_key = var.openai_project_api_key
}
```

## Argument Reference

* `project_id` - (Required) The ID of the project this rate limit applies to (format: `proj_abc123456789`).
* `model` - (Required) The model this rate limit applies to (e.g., `gpt-4`, `gpt-3.5-turbo`, `dall-e-3`). This identifies which model's usage will be limited.
* `max_requests_per_minute` - (Optional) Maximum number of API requests allowed per minute for this model in this project.
* `max_tokens_per_minute` - (Optional) Maximum number of tokens that can be processed per minute for this model in this project.
* `max_images_per_minute` - (Optional) Maximum number of images that can be generated per minute. Only relevant for image models like DALL-E.
* `batch_1_day_max_input_tokens` - (Optional) Maximum number of input tokens allowed in batch operations per day. Only relevant for certain models that support batch operations.
* `api_key` - (Optional) A project-specific API key to use when managing this rate limit. If not provided, the provider's default API key will be used.

-> **Note:** At least one limit parameter must be specified.

## Additional Parameters Available in OpenAI API

The OpenAI API supports these additional rate limit parameters which can be configured:

* `max_audio_megabytes_per_1_minute` - (Optional) The maximum audio megabytes per minute. Only relevant for audio models like Whisper.
* `max_requests_per_1_day` - (Optional) The maximum requests per day. Provides a daily cap in addition to the per-minute rate.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A unique identifier for this rate limit in Terraform.
* `rate_limit_id` - The OpenAI-assigned ID for this rate limit (format: `rl-modelname`).

## Resource Lifecycle

The OpenAI API doesn't actually allow for true deletion of rate limits. When a rate limit resource is removed from your Terraform configuration:

1. The provider will attempt to reset the rate limit to the default values it captured when the rate limit was initially created.
2. The resource will then be removed from the Terraform state.

This ensures that removing a rate limit from your Terraform configuration doesn't leave custom rate limits in place on the OpenAI side.

## Import

Rate limits can be imported using the project ID and model name combination, separated by a colon:

```
terraform import openai_rate_limit.example proj_abc123456789:gpt-4
```

## Requirements

* An OpenAI API key with appropriate permissions to modify rate limits
* Project API keys can only set rate limits for their own project
* Organization admin keys can set rate limits for any project

## Notes

* Rate limits are applied per model and project
* Changes to rate limits can take a few minutes to propagate through OpenAI's systems
* Setting very low rate limits might affect application performance
* Setting `null` or removing a limit parameter will reset it to the default (unlimited) 