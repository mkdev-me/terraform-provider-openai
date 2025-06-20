# OpenAI Service Account Module

This Terraform module manages service accounts within OpenAI projects. Service accounts are non-human users that can be used for automation and integration purposes.

## Features

- Create service accounts in OpenAI projects
- Read information about existing service accounts
- Manage service account lifecycle with Terraform

## Important Notes

### API Keys

OpenAI does not support programmatic creation of project API keys. After creating a service account with this module, you must manually create API keys in the OpenAI dashboard.

### Required Permissions

This module requires specific API key permissions depending on the mode:

- For **resource mode**: An admin API key with `api.organization.projects.service_accounts.write` scope
- For **data source mode**: An admin API key with `api.organization.projects.service_accounts.read` scope

If you don't have these permissions, the module will gracefully degrade and provide empty values.

## Advantages of Service Accounts

Unlike regular users, service accounts:
- Continue working even when team members leave the organization
- Provide dedicated identities for automated systems
- Enable clear audit trails of system activities
- Allow for fine-grained permission control

## Usage

This module can operate in two modes:
1. **Resource mode** (default): Creates and manages service accounts
2. **Data source mode**: Reads information about existing service accounts

### Creating a New Service Account

```hcl
module "service_account" {
  source     = "path/to/modules/service_account"
  project_id = "proj_abc123"
  name       = "My Service Account"
  
  # Optional: Custom API key (requires admin permissions with write scope)
  api_key    = var.openai_admin_key
}

# Outputs
output "service_account_id" {
  value = module.service_account.service_account_id
}

# Note: After creating the service account, you need to manually 
# create API keys in the OpenAI dashboard
```

### Reading an Existing Service Account (Data Source Mode)

```hcl
module "existing_service_account" {
  source            = "path/to/modules/service_account"
  project_id        = "proj_abc123"
  service_account_id = "service_xyz789"
  use_data_source   = true
  
  # Required: API key with api.organization.projects.service_accounts.read scope
  api_key           = var.openai_admin_key
}

# Outputs
output "service_account_name" {
  value = module.existing_service_account.name
}

output "service_account_role" {
  value = module.existing_service_account.role
}

output "created_at" {
  value = module.existing_service_account.created_at
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 0.13.0 |
| openai | ~> 1.0.0 |

## Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| project_id | The ID of the project where the service account will be created or read | `string` | n/a | yes |
| name | The name of the service account (required for creation mode) | `string` | `null` | yes in the reousource |
| service_account_id | The ID of an existing service account to retrieve (required for data source mode) | `string` | `null` | no |
| use_data_source | Whether to use a data source to read an existing service account | `bool` | `false` | no |
| api_key | Custom API key to use for this resource | `string` | `null` | no |

**Note**: The `name` parameter is required when creating service accounts (when `use_data_source = false`). The module will not attempt to create a service account if the name is not provided.

## Outputs

| Name | Description |
|------|-------------|
| id | The composite ID of the service account (project_id:service_account_id) |
| service_account_id | The ID of the service account |
| name | The name of the service account |
| created_at | The timestamp when the service account was created |
| role | The role of the service account |
| project_id | The project ID where the service account was created |
| api_key_id | The ID of the API key for the service account (if available) |
| api_key_value | The value of the API key (only available when creating a new service account) |

## Handling Permission Errors

This module is designed to gracefully handle permission errors. If you don't have the required permissions:

1. In data source mode, the module will return empty values rather than failing
2. Default values will be used for outputs
3. Meaningful error messages will be shown during plan/apply but won't prevent other operations

## Best Practices

1. **Name service accounts meaningfully**: Use descriptive names that indicate the purpose of the service account.

2. **Rotate API keys regularly**: Create new API keys periodically and update your systems to use the new keys.

3. **Use least privilege principle**: Give service accounts only the permissions they need to fulfill their function.

4. **Monitor service account activity**: Set up monitoring to detect unusual patterns or potential security issues.

5. **Document service accounts**: Maintain documentation about what each service account is used for and which systems depend on it.

## Limitations

- Service accounts cannot be modified after creation. To change properties, you must delete and recreate the service account.
- API keys are only shown once upon creation. Make sure to capture them when they're first created.
- Service accounts can only be created with an admin API key that has the necessary permissions
- API keys must be created manually in the OpenAI dashboard - programmatic creation is not supported

## Related Resources

- [Project Service Account Resource](../../docs/resources/project_service_account.md)
- [Project API Key Data Source](../../docs/data-sources/project_api_key.md) 