# OpenAI Project API Key Module

This module provides a way to retrieve information about existing OpenAI project API keys using Terraform. It includes functionality for retrieving information about a specific API key or listing all API keys for a project.

## Important Note

* This module only retrieves information about **existing** API keys
* OpenAI does not support programmatically creating project API keys via their API
* You must first create API keys manually in the OpenAI dashboard
* The module then allows you to reference these keys in your Terraform configurations

## Module Features

- Retrieve information about a specific API key within a project
- List all API keys for a project

## Usage

### Retrieving Information About a Specific API Key

```hcl
module "project_api_key" {
  source = "../../modules/project_api"

  project_id   = "proj_abc123xyz"   # The project ID the API key belongs to
  api_key_id   = "key_abc123xyz"    # The ID of the API key to look up
  retrieve_all = false              # Set to false to look up a single API key
  
  openai_admin_key = var.openai_admin_key
}

output "api_key_name" {
  value = module.project_api_key.name
}

output "api_key_created_at" {
  value = module.project_api_key.created_at
}
```

### Retrieving All API Keys for a Project

```hcl
module "all_project_keys" {
  source = "../../modules/project_api"

  project_id   = "proj_abc123xyz"  # The project ID to retrieve API keys for
  retrieve_all = true              # Set to true to retrieve all API keys
  
  openai_admin_key = var.openai_admin_key
}

output "all_api_keys" {
  value = module.all_project_keys.api_keys
}

output "api_key_count" {
  value = length(module.all_project_keys.api_keys)
}
```

## Input Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| `project_id` | The ID of the project to retrieve API keys for | `string` | n/a | yes |
| `api_key_id` | The ID of the API key to look up (required when retrieve_all is false) | `string` | `null` | no |
| `retrieve_all` | Whether to retrieve all API keys for the project | `bool` | `false` | no |
| `openai_admin_key` | OpenAI Admin API key | `string` | n/a | yes |
| `organization_id` | OpenAI Organization ID | `string` | `""` | no |

## Outputs

| Name | Description | Condition |
|------|-------------|-----------|
| `api_key_id` | The ID of the API key | Available when retrieve_all = false |
| `name` | The name of the API key | Available when retrieve_all = false |
| `created_at` | Timestamp when the API key was created | Available when retrieve_all = false |
| `last_used_at` | Timestamp when the API key was last used | Available when retrieve_all = false |
| `api_keys` | List of all API keys for the project | Available when retrieve_all = true |

## Examples

See the [examples/project_api](../../examples/project_api) directory for examples of how to use this module.

## Creating Project API Keys

Since OpenAI doesn't support programmatic creation of project API keys, you must create them manually:

1. Log in to the [OpenAI Platform](https://platform.openai.com/)
2. Navigate to your desired project
3. Go to the API Keys section
4. Create a new API key
5. Copy the API key ID (format: `key_abc123xyz`)
6. Use this module to retrieve information about the key

## Admin API Keys

For creating and managing admin API keys programmatically, see the [Admin API Key module](../admin_api_key/README.md). 