---
page_title: "OpenAI: openai_project_api_key Data Source"
subcategory: "Project API Keys"
description: |-
  Retrieves information about a specific API key within an OpenAI project.
---

# Data Source: openai_project_api_key

Retrieves information about a specific API key within an OpenAI project.

-> **Note:** This data source requires an OpenAI admin API key with appropriate permissions.

## Example Usage

```hcl
data "openai_project_api_key" "example" {
  project_id = "proj_abc123xyz"
  api_key_id = "key_abc123xyz"
}

output "key_name" {
  value = data.openai_project_api_key.example.name
}

output "key_created_at" {
  value = data.openai_project_api_key.example.created_at
}
```

## Argument Reference

* `project_id` - (Required) The ID of the project the API key belongs to (format: `proj_abc123xyz`).
* `api_key_id` - (Required) The ID of the API key to retrieve (format: `key_abc123xyz`).

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A composite ID in the format `project_id:api_key_id`.
* `name` - The name of the API key.
* `created_at` - The timestamp when the API key was created (RFC3339 format).
* `last_used_at` - The timestamp when the API key was last used (RFC3339 format), if available.

## Important Notes

* This data source does not provide access to the actual API key value for security reasons.
* Only metadata such as name, creation time, and last used time are available.
* API keys must be created manually in the OpenAI dashboard.
* The OpenAI API does not support programmatic creation of project API keys.

## Permissions Required

This data source requires an OpenAI admin API key with the following permissions:

* Organization role of Admin or Owner
* `api.management.read` scope 