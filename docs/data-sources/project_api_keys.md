---
page_title: "OpenAI: openai_project_api_keys"
subcategory: "Project API Keys"
description: |-
  Retrieves a list of all API keys for a specific OpenAI project.
---

# Data Source: openai_project_api_keys

Retrieves a list of all API keys for a specific OpenAI project.

-> **Note:** This data source requires an OpenAI admin API key with appropriate permissions.

## Example Usage

```hcl
data "openai_project_api_keys" "all_keys" {
  project_id = "proj_abc123xyz"
}

output "key_count" {
  value = length(data.openai_project_api_keys.all_keys.api_keys)
}

output "key_names" {
  value = [for key in data.openai_project_api_keys.all_keys.api_keys : key.name]
}
```

## Argument Reference

* `project_id` - (Required) The ID of the project to retrieve API keys for (format: `proj_abc123xyz`).

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The project ID.
* `api_keys` - A list of API keys for the project. Each API key contains:
  * `id` - The ID of the API key (format: `key_abc123xyz`).
  * `name` - The name of the API key.
  * `created_at` - The timestamp when the API key was created (RFC3339 format).
  * `last_used_at` - The timestamp when the API key was last used (RFC3339 format), if available.

## Important Notes

* This data source does not provide access to the actual API key values for security reasons.
* Only metadata such as name, creation time, and last used time are available.
* API keys must be created manually in the OpenAI dashboard.
* The OpenAI API does not support programmatic creation of project API keys.

## Permissions Required

This data source requires an OpenAI admin API key with the following permissions:

* Organization role of Admin or Owner
* `api.management.read` scope 