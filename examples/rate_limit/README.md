# OpenAI Rate Limit Example

This example demonstrates how to:
1. Create and manage rate limits for OpenAI models
2. Retrieve rate limit information using data sources

## Important Note

This example creates a real OpenAI project but **does not** set actual rate limits. Instead, it:

1. Creates a real OpenAI project using the `openai_project` resource
2. Uses the `openai_chat_completion` resource to document your intended rate limits
3. Provides human-readable documentation about your rate limit configuration

You will need to manually set the actual rate limits in the OpenAI dashboard after applying this configuration.

## Important Note About Rate Limit Deletion

When working with the OpenAI API, it's important to understand that rate limits cannot be truly deleted. Instead, when you remove a rate limit resource from your Terraform configuration, the provider will reset the rate limit to its default value.

The provider has been improved to handle this reset process reliably during `terraform destroy` operations. The deletion functionality in the provider has been fixed to properly handle type conversions and client retrieval, ensuring consistent behavior without panic errors.

## How This Example Works

Instead of using the module which may have compatibility issues, this example:

1. Uses a direct approach with `openai_rate_limit` to set rate limits for various models
2. Shows how to set different types of limits for different models:
   - `max_requests_per_minute` and `max_tokens_per_minute` for GPT models
   - `max_images_per_minute` and `max_requests_per_1_day` for DALL-E
   - `max_audio_megabytes_per_1_minute` for Whisper
   - `batch_1_day_max_input_tokens` for batch operations with embeddings
3. Demonstrates use of project-specific API keys for managing rate limits
4. Safely resets rate limits to default values when destroyed

## Data Sources Example

The `data_sources.tf` file demonstrates how to use the rate limit data sources:

1. `openai_rate_limit` - Retrieve information about a specific model's rate limit
2. `openai_rate_limits` - Retrieve information about all rate limits for a project
3. Using the `rate_limit` module in read-only mode

### Important Note About Permissions

The data sources require an OpenAI admin API key with the `api.management.read` scope. Regular project API keys typically don't have these permissions.

If you want to try using the data sources:

1. Set the `try_data_sources` variable to `true`:
```bash
terraform apply -var="try_data_sources=true"
```

2. Make sure you're using an admin key with the right permissions:
```bash
export TF_VAR_openai_admin_key="sk-admin-..."
```

If you don't have the right permissions, the data sources will fail with a permissions error, but the example will still work for setting rate limits through the resources.

This approach allows you to:
- Monitor your existing rate limits (with admin permissions)
- Create reports of configured limits across models
- Make decisions in your Terraform code based on current limits

## Prerequisites

- An OpenAI API key with admin permissions
- For data sources: API key needs the `api.management.read` scope
- For rate limit creation: API key needs permissions to modify rate limits

## Usage

### Setting up variables

```bash
# Export your OpenAI Admin API key
export OPENAI_ADMIN_API="sk-admin-yourkey"
```

### Create or modify rate limits

```bash
# Apply the configuration with default values
terraform apply -var="openai_api_key=$OPENAI_ADMIN_API"

# To also test the data sources, set the try_data_sources variable to true
terraform apply -var="openai_api_key=$OPENAI_ADMIN_API" -var="try_data_sources=true"
```

## Example components

### Resource creation

The `main.tf` file demonstrates how to create a resource to set rate limits for a specific model (DALL-E 3):

```hcl
resource "openai_rate_limit" "dalle3_limits" {
  project_id = var.project_id
  model      = "dall-e-3"
  
  max_requests_per_minute = 50
  max_images_per_minute   = 10
  max_requests_per_1_day  = 600
}
```

### Data sources

The `data_sources.tf` file demonstrates how to use both the singular and plural data sources:

```hcl
# Retrieve all rate limits for a project
data "openai_rate_limits" "all_limits" {
  count      = var.try_data_sources ? 1 : 0
  project_id = var.project_id
}

# Retrieve rate limit for a specific model
data "openai_rate_limit" "dalle3_limit" {
  count      = var.try_data_sources ? 1 : 0
  project_id = var.project_id
  model      = "dall-e-3"
}
```

## Important notes

1. **Authentication**: This example uses the same API key for both creating resources and retrieving data. In production, you might use different keys with appropriate permissions.

2. **Error handling**: The provider has been updated to handle authentication errors gracefully. If your API key doesn't have the necessary permissions for data sources, it will continue with warnings rather than failing.

3. **Conditional usage**: The `try_data_sources` variable conditionally attempts to use the data sources. This allows the example to work even without the required permissions.

4. **Resource deletion**: When you destroy the resources using `terraform destroy`, the rate limits will be reset to their default values, not completely removed. This is a limitation of the OpenAI API, but the provider handles it gracefully.

## Output examples

When you run the example with data sources enabled:

```bash
terraform apply -var="openai_api_key=$OPENAI_ADMIN_API" -var="try_data_sources=true"
```

You'll see outputs like:

```
Outputs:

dalle3_limits_from_datasource = {
  "max_images_per_minute" = 10
  "max_requests_per_1_day" = 600
  "max_requests_per_minute" = 50
  "model" = "dall-e-3"
}
dalle3_rate_limit_id = "rl-dall-e-3"
project_id = "proj_JGhw44csZsbtjw2yxuyPjMZN"
```

## Setting Actual Rate Limits

After applying the Terraform configuration:

1. Log in to the [OpenAI platform](https://platform.openai.com/)
2. Navigate to Settings > Organization > Rate limits
3. Select your newly created "Rate Limit Test Project"
4. Manually set the rate limits to match what's documented in your Terraform configuration

## Importing Existing Rate Limits

You can import existing rate limits into your Terraform state to manage them with Terraform:

```bash
terraform import -var="openai_admin_key=$OPENAI_ADMIN_KEY" openai_rate_limit.gpt4_limits proj_abc123xyz:rl-gpt-4-xxxxxxxx
```

For this to work properly:

1. The import ID must include both the project ID and rate limit ID in the format:
   `project_id:rate_limit_id`

2. You must use an admin API key with sufficient permissions to read rate limits.

3. Your configuration file should have a resource defined for the rate limit you're importing, but you don't need to pre-fill the project_id or other fields.

Example import command:

```bash
terraform import -var="openai_admin_key=$OPENAI_ADMIN_KEY" \
  openai_rate_limit.gpt4_limits proj_JGhw44csZsbtjw2yxuyPjMZN:rl-gpt-4-xxxxxxxx
```

After importing, you can verify the imported state with:

```bash
terraform state show openai_rate_limit.gpt4_limits
```

This makes it easy to bring existing rate limits under Terraform management, especially useful when:
- Taking over management of existing OpenAI projects
- Documenting current rate limits before making changes
- Building infrastructure as code for existing environments

## Clean Up

To remove resources created by this example (resetting rate limits to defaults):

```bash
terraform destroy
```

Note that this will archive the OpenAI project rather than completely delete it, and rate limits will be reset to their default values, not completely removed. 