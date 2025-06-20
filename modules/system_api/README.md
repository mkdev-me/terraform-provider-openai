# OpenAI Admin API Key Module

This Terraform module creates and manages OpenAI Admin API keys using the native Terraform OpenAI provider resources. It provides a mechanism for creating, tracking, and deleting OpenAI system-level API keys with specific permissions and expiration settings.

## Features

- Create OpenAI Admin API keys with customizable settings
- Support for permanent keys (no expiration) or temporary keys with expiration dates
- Restrict key permissions with specific scopes for granular access control
- Return key values as Terraform outputs (with sensitive flag for security)
- List all admin API keys in your organization

## Requirements

- OpenAI account with administrator privileges
- Admin API key with `api.management.write` permissions
- Terraform 0.13+

## Usage

### Basic Usage

```hcl
module "admin_key" {
  source = "../../modules/system_api"
  
  name = "terraform-admin-key"
}

output "admin_key_id" {
  value = module.admin_key.key_id
}

output "admin_key_value" {
  value     = module.admin_key.key_value
  sensitive = true
}
```

### With Expiration Date

```hcl
module "temporary_admin_key" {
  source = "../../modules/system_api"
  
  name       = "temporary-admin-key"
  expires_at = 1735689600  # Unix timestamp for 2025-01-01
}
```

### With Restricted Scopes

```hcl
module "restricted_admin_key" {
  source = "../../modules/system_api"
  
  name   = "read-only-admin-key"
  scopes = ["api.management.read"]
}
```

### Using a Custom Admin API Key

```hcl
module "admin_key_custom_auth" {
  source = "../../modules/system_api"
  
  name    = "admin-key-with-custom-auth"
  api_key = var.openai_admin_key  # Your existing admin API key with required permissions
}
```

### List All Admin API Keys

```hcl
module "admin_keys" {
  source = "../../modules/system_api"
  
  name      = "terraform-admin-key"  # Still required to create a key
  list_keys = true                   # Enable listing of all keys
  list_limit = 50                    # Retrieve up to 50 keys (default: 20)
}

output "all_admin_keys" {
  value = module.admin_keys.all_api_keys
}

output "has_more_keys" {
  value = module.admin_keys.has_more_keys
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| name | The name of the system API key | string | n/a | yes |
| expires_at | Unix timestamp for when the API key expires | number | null | no |
| scopes | The scopes this key is restricted to (e.g. api.management.read) | list(string) | [] | no |
| api_key | Custom API key to use for this operation | string | null | no |
| list_keys | Whether to list all admin API keys | bool | false | no |
| list_limit | Maximum number of API keys to retrieve when listing | number | 20 | no |
| list_after | API key ID to start listing from (for pagination) | string | null | no |

## Outputs

| Name | Description |
|------|-------------|
| key_id | The ID of the API key (format: key_XXXX) |
| key_value | The value of the API key (format: sk-admin-XXXX). Only returned at creation time and marked as sensitive. |
| created_at | Creation timestamp of the API key (ISO 8601 format) |
| name | The name of the API key as set during creation |
| expires_at | Expiration date of the API key (Unix timestamp, if specified) |
| all_api_keys | List of all admin API keys (only available if list_keys is true) |
| has_more_keys | Whether there are more API keys available beyond the limit |

## Scope Options

Admin API keys can be restricted to specific permission scopes:

- `api.management.read` - Read access to management operations
- `api.management.write` - Write access to management operations  
- `api.rate_limits.read` - Read access to rate limit configurations
- `api.rate_limits.write` - Write access to rate limit configurations
- `api.projects.read` - Read access to projects
- `api.projects.write` - Write access to projects

If no scopes are provided, the key will have full admin access.

## Security Best Practices

- Use the principle of least privilege when assigning scopes
- Set appropriate expiration dates for temporary access
- Store key values securely in encrypted storage
- Rotate keys regularly
- Never commit keys to version control
- Use unique, descriptive names for each key 