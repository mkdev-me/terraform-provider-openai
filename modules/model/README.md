# OpenAI Model Module

This module provides a simple way to work with OpenAI models in your Terraform configurations.

## Usage

```hcl
module "openai_model" {
  source = "../../modules/model"
  
  model_id = "gpt-4"
  
  # Optional: Use a project-specific API key instead of the provider's default
  # api_key = var.project_api_key
}

output "model_details" {
  value = module.openai_model.model_details
}
```

## API Key Configuration

This module supports API key configuration in multiple ways:

1. **Provider-level API key** - Set via the provider block or `OPENAI_API_KEY` environment variable
2. **Module-level API key** - Pass an `api_key` variable to this module

### Example with explicit API key

```hcl
provider "openai" {
  # Default provider API key
  api_key = var.openai_api_key
}

module "default_model" {
  source  = "../../modules/model"
  model_id = "gpt-4"
  # Uses the provider's default API key
}

module "project_model" {
  source  = "../../modules/model"
  model_id = "gpt-4"
  api_key = var.project_api_key  # Override provider API key for this module
}
```

### Troubleshooting API Keys

If you encounter errors related to API keys, try:

1. Ensuring your API key is correctly set and has the necessary permissions
2. Explicitly setting the API key at the provider or module level instead of using environment variables
3. Creating a terraform.tfvars file with your API keys:
   ```
   openai_api_key = "sk-..."
   project_api_key = "sk-proj-..."
   ```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| model_id | The ID of the model to retrieve | `string` | n/a | yes |
| api_key | Optional project-specific API key | `string` | `""` | no |

## Outputs

| Name | Description |
|------|-------------|
| model_details | Complete details about the model |
| model_id | ID of the model | 