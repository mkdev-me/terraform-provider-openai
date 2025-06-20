---
page_title: "OpenAI: openai_projects Data Source"
subcategory: ""
description: |-
  Retrieves a list of all OpenAI projects in an organization.
---

# openai_projects Data Source

Retrieves a list of all OpenAI projects in an organization. This data source is useful for auditing projects, creating dynamic configurations based on existing projects, and generating reports on project usage and configuration.

## Example Usage

```hcl
data "openai_projects" "all" {
  # No specific parameters required
  
  # Optional: Use a custom API key (must have permission to read project details)
  api_key = var.openai_admin_key
}

# Count projects
output "project_count" {
  value = length(data.openai_projects.all.projects)
}

# List project names
output "project_names" {
  value = [for p in data.openai_projects.all.projects : p.name]
}

# Find the default project
output "default_project" {
  value = [for p in data.openai_projects.all.projects : p if p.is_default][0]
}
```

## Argument Reference

* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A unique identifier for this data source invocation.
* `projects` - A list of projects in the organization. Each project contains:
  * `id` - The ID of the project.
  * `name` - The name of the project.
  * `status` - The status of the project. Can be "active" or "archived".
  * `created_at` - The timestamp (in Unix time) when the project was created.
  * `usage_limits` - Usage limits for the project, including:
    * `max_monthly_dollars` - Maximum monthly spend in dollars.
    * `max_parallel_requests` - Maximum number of parallel requests allowed.
    * `max_tokens` - Maximum number of tokens per request.

## Permission Requirements

To use this data source, your API key must have permissions to read project details. Typically, this requires an admin API key with organization-level permissions.

## Use Cases

This data source is particularly useful for:

* **Auditing**: Keep track of all projects in your organization
* **Dynamic Configuration**: Create resources in specific projects based on project attributes
* **Reporting**: Generate reports on project usage and configuration
* **Automation**: Automate tasks across all projects in your organization

## Technical Notes

* This data source is read-only and doesn't make any changes to your OpenAI resources
* The response includes basic project information; for detailed information about a specific project, use the `openai_project` data source
* The list may include both active and archived projects 