---
page_title: "openai_group Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Use this data source to retrieve information about a specific group in your OpenAI organization.
---

# openai_group (Data Source)

Use this data source to retrieve information about a specific group in your OpenAI organization.

This data source requires an **Admin API Key** for authentication.

## Example Usage

```terraform
# Look up a group by ID
data "openai_group" "by_id" {
  id = "group-abc123"
}

# Look up a group by name
data "openai_group" "by_name" {
  name = "Engineering Team"
}

# Output the group details
output "group_id" {
  value = data.openai_group.by_name.id
}

output "group_created_at" {
  value = data.openai_group.by_name.created_at
}

output "is_scim_managed" {
  value = data.openai_group.by_name.is_scim_managed
}
```

## Schema

### Optional

- `id` (String) The ID of the group to retrieve. Either id or name must be provided.
- `name` (String) The name of the group to search for. Either id or name must be provided.

### Read-Only

- `created_at` (Number) Unix timestamp (in seconds) when the group was created.
- `is_scim_managed` (Boolean) Whether the group is managed through SCIM and controlled by your identity provider.

## Notes

- This data source requires an Admin API Key (`admin_key` in provider configuration or `OPENAI_ADMIN_KEY` environment variable).
- Either `id` or `name` must be provided to look up the group.
- Name lookups are case-insensitive.
