---
page_title: "OpenAI: openai_project_service_account Data Source"
subcategory: ""
description: |-
  Retrieves information about a specific service account in an OpenAI project.
---

# openai_project_service_account Data Source

Retrieves information about a specific service account in an OpenAI project. Service accounts are bot users that are not associated with a real user and do not have the limitation of being removed when a user leaves an organization.

## Example Usage

```hcl
data "openai_project_service_account" "example" {
  project_id         = "proj_abc123"
  service_account_id = "user-abc123"
  
  # Optional: Use a custom API key (must be an admin key with api.organization.projects.service_accounts.read scope)
  api_key = var.openai_admin_key
}

# Access the service account details
output "service_account_name" {
  value = data.openai_project_service_account.example.name
}
```

### With Resource Dependency

When used together with a service account resource, establish proper dependencies:

```hcl
# Create a service account
resource "openai_project_service_account" "example" {
  count      = var.create_service_account ? 1 : 0
  project_id = "proj_abc123"
  name       = "Example Account"
}

# Data source with conditional execution and explicit dependency
data "openai_project_service_account" "example" {
  count              = var.use_data_sources && var.create_service_account ? 1 : 0
  project_id         = "proj_abc123"
  service_account_id = var.create_service_account ? openai_project_service_account.example[0].service_account_id : "placeholder_id"
  api_key            = var.openai_admin_key
  
  depends_on = [
    openai_project_service_account.example
  ]
}

# Safe output with fallback
output "service_account_role" {
  value = var.use_data_sources && var.create_service_account ? 
    try(data.openai_project_service_account.example[0].role, "Permission error") : 
    "Data source disabled"
}
```

This approach ensures that:
- The data source is only evaluated when the service account exists
- The dependency is properly established to avoid race conditions
- Permission errors are handled gracefully

## Argument Reference

* `project_id` - (Required) The ID of the project to which the service account belongs.
* `service_account_id` - (Required) The ID of the service account to retrieve.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used. **Note:** This must be an organization admin API key with the `api.organization.projects.service_accounts.read` scope.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A composite ID uniquely identifying the service account in the format `{project_id}:{service_account_id}`.
* `name` - The name of the service account.
* `created_at` - The timestamp (in Unix time) when the service account was created.
* `role` - The role of the service account.
* `api_key_id` - The ID of the API key associated with the service account.

## Permission Requirements

To use this data source, you must have an admin API key with the `api.organization.projects.service_accounts.read` scope. Regular project API keys do not have sufficient permissions to read service account details.

## Import

Service account data sources cannot be imported. 