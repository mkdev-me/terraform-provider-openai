---
page_title: "OpenAI: openai_project_service_accounts Data Source"
subcategory: ""
description: |-
  Retrieves a list of all service accounts in an OpenAI project.
---

# openai_project_service_accounts Data Source

Retrieves a list of all service accounts in an OpenAI project. Service accounts are bot users that are not associated with a real user and do not have the limitation of being removed when a user leaves an organization.

## Example Usage

```hcl
data "openai_project_service_accounts" "all" {
  project_id = "proj_abc123"
  
  # Optional: Use a custom API key (must be an admin key with api.organization.projects.service_accounts.read scope)
  api_key = var.openai_admin_key
}

# Access the list of service accounts
output "service_account_names" {
  value = [for sa in data.openai_project_service_accounts.all.service_accounts : sa.name]
}

# Count service accounts
output "service_account_count" {
  value = length(data.openai_project_service_accounts.all.service_accounts)
}
```

## Argument Reference

* `project_id` - (Required) The ID of the project from which to retrieve service accounts.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used. **Note:** This must be an organization admin API key with the `api.organization.projects.service_accounts.read` scope.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A unique ID for this data source in the format `{project_id}-service-accounts-{timestamp}`.
* `service_accounts` - A list of service accounts in the project. Each service account has the following attributes:
  * `id` - The ID of the service account.
  * `name` - The name of the service account.
  * `created_at` - The timestamp (in Unix time) when the service account was created.
  * `role` - The role of the service account.
  * `api_key_id` - The ID of the API key associated with the service account.

## Permission Requirements

To use this data source, you must have an admin API key with the `api.organization.projects.service_accounts.read` scope. Regular project API keys do not have sufficient permissions to read service account details.

## Import

Service account data sources cannot be imported. 