---
page_title: "OpenAI: openai_project Data Source"
subcategory: ""
description: |-
  Retrieves information about an OpenAI project.
---

# openai_project Data Source

Retrieves information about an OpenAI project by its ID. Projects in OpenAI are used to organize resources and users.

## Example Usage

```hcl
data "openai_project" "example" {
  project_id = "proj_abc123"
  
  # Optional: Use a custom API key (must have permission to read project details)
  api_key = var.openai_api_key
}

# Access project details
output "project_name" {
  value = data.openai_project.example.name
}

output "project_status" {
  value = data.openai_project.example.status
}
```

## Argument Reference

* `project_id` - (Required) The ID of the project to retrieve information for.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the project.
* `name` - The name of the project.
* `description` - The description of the project.
* `organization_id` - The ID of the organization the project belongs to.
* `created_at` - The timestamp (in Unix time) when the project was created.
* `archived_at` - The timestamp (in Unix time) when the project was archived, if applicable.
* `status` - The status of the project. Can be "active" or "archived".
* `is_default` - Whether this is the default project for the organization.
* `billing_mode` - The billing mode of the project.

## Permission Requirements

To use this data source, your API key must have permissions to read project details. If using a custom API key, it must have the appropriate permissions.

## Import

Project data sources cannot be imported. 