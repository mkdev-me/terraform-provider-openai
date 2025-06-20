# OpenAI Rate Limit Terraform Module

This Terraform module enables you to manage rate limits for OpenAI projects and models. Rate limits help control API usage and costs by setting caps on requests and tokens per minute.

## Important Note About Rate Limit Deletion

When working with the OpenAI API, it's important to understand that rate limits cannot be truly deleted. Instead, when you remove a rate limit resource from your Terraform configuration, the provider will reset the rate limit to its default value.

The provider has been improved to handle this reset process reliably during `terraform destroy` operations. The deletion functionality in the provider has been fixed to properly handle type conversions and client retrieval, ensuring consistent behavior without panic errors.

## Usage

### Creating or Updating Rate Limits

```hcl
module "gpt4_rate_limit" {
  source = "../../modules/rate_limit"

  project_id              = "proj_abc123456789"  # Your OpenAI project ID
  model                   = "gpt-4"              # The model to limit
  max_requests_per_minute = 100                  # Limit to 100 requests per minute
  max_tokens_per_minute   = 10000                # Limit to 10K tokens per minute
  api_key                 = var.openai_api_key   # API key with sufficient permissions
}
```

### Reading Existing Rate Limits (Data Source Mode)

```hcl
module "gpt4_rate_limit_read" {
  source = "../../modules/rate_limit"

  project_id       = "proj_abc123456789"  # Your OpenAI project ID
  model            = "gpt-4"              # The model to retrieve
  use_data_source  = true                 # Use the data source instead of the resource
  api_key          = var.openai_api_key   # API key with sufficient permissions
}

output "current_token_limit" {
  value = module.gpt4_rate_limit_read.max_tokens_per_minute
}
```

### Retrieving All Rate Limits (List Mode)

```hcl
module "all_rate_limits" {
  source = "../../modules/rate_limit"

  project_id       = "proj_abc123456789"  # Your OpenAI project ID
  list_mode        = true                 # Retrieve all rate limits for the project
  api_key          = var.openai_api_key   # API key with sufficient permissions
  
  # These are still required by the module but ignored in list mode
  model = "unused-in-list-mode"
  max_requests_per_minute = 0
  max_tokens_per_minute = 0
}

output "all_model_limits" {
  value = module.all_rate_limits.all_rate_limits
}

# Access a specific rate limit from the list
output "gpt4_limit" {
  value = [for limit in module.all_rate_limits.all_rate_limits : limit if limit.model == "gpt-4"][0]
}
```

## Provider Compatibility

This module requires the `fjcorp/openai` provider version `~> 1.0.0` and uses the now fully implemented `openai_rate_limit` resource.

## Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| project_id | The ID of the project this rate limit applies to | string | n/a | yes |
| model | The model this rate limit applies to (e.g., 'gpt-4', 'gpt-3.5-turbo') | string | n/a | yes |
| use_data_source | Whether to use a data source to read existing rate limits instead of creating/updating them | bool | false | no |
| list_mode | Whether to retrieve all rate limits instead of working with a single rate limit | bool | false | no |
| max_requests_per_minute | Maximum number of API requests allowed per minute | number | null | no |
| max_tokens_per_minute | Maximum number of tokens that can be processed per minute | number | null | no |
| max_images_per_minute | Maximum number of images that can be generated per minute | number | null | no |
| batch_1_day_max_input_tokens | Maximum number of input tokens allowed in batch operations per day | number | null | no |
| max_audio_megabytes_per_1_minute | Maximum number of audio megabytes that can be processed per minute | number | null | no |
| max_requests_per_1_day | Maximum number of API requests allowed per day | number | null | no |
| api_key | API key for accessing the project | string | null | no |

## Outputs

| Name | Description |
|------|-------------|
| id | The unique identifier for this rate limit |
| rate_limit_id | The OpenAI rate limit ID (format: rl-XXX) |
| project_id | The ID of the project this rate limit applies to |
| model | The model this rate limit applies to |
| max_requests_per_minute | The configured maximum requests per minute |
| max_tokens_per_minute | The configured maximum tokens per minute |
| max_images_per_minute | The configured maximum images per minute |
| batch_1_day_max_input_tokens | The configured maximum batch input tokens per day |
| max_audio_megabytes_per_1_minute | The configured maximum audio megabytes per minute |
| max_requests_per_1_day | The configured maximum requests per day |
| all_rate_limits | List of all rate limits (only populated when list_mode = true) |

## Implementation Details

The module now uses the actual `openai_rate_limit` resource which directly applies rate limits to your OpenAI project through the API. This allows you to manage your rate limits directly through Terraform.

### Rate Limit Lifecycle

When you remove a rate limit resource from your Terraform configuration and run `terraform destroy`:
1. The rate limit is not completely deleted from the OpenAI API (this is not supported by the API)
2. Instead, the provider resets the rate limit to default values for that model
3. This behavior is now implemented reliably with proper error handling to avoid panic errors

### API Key Permissions

- For full functionality, your API key must have appropriate permissions to modify rate limits
- Project API keys can set rate limits for their own project only
- Organization admin keys have full permissions to set rate limits for any project 