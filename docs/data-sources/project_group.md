---
page_title: "openai_project_group Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Use this data source to retrieve information about a specific group in an OpenAI project.
---

# openai_project_group (Data Source)

Use this data source to retrieve information about a specific group in an OpenAI project.

This data source requires an **Admin API Key** for authentication.

## Example Usage

```terraform
# Look up a specific group in a project by group ID
data "openai_project_group" "engineering" {
  project_id = "proj-abc123"
  group_id   = "group-xyz789"
}

# Look up a group by name (useful when you don't know the group ID)
data "openai_project_group" "support_team" {
  project_id = "proj-abc123"
  group_name = "Support Team"
}

# Output the group's role and when it was added
output "engineering_group_role" {
  value = data.openai_project_group.engineering.role
}

output "engineering_group_added_at" {
  value = data.openai_project_group.engineering.created_at
}

output "support_team_group_id" {
  value = data.openai_project_group.support_team.group_id
}
```

## Schema

### Required

- `project_id` (String) The ID of the project to retrieve the group from.

### Optional

- `group_id` (String) The ID of the group to retrieve.
- `group_name` (String) The name of the group to search for. Used if group_id is not provided.

~> **Note:** Either `group_id` or `group_name` must be provided.

### Read-Only

- `id` (String) The ID of the resource (composite of project_id:group_id).
- `role` (String) The role of the group in the project (e.g., 'owner' or 'member').
- `created_at` (Number) Timestamp when the group was added to the project.

## Notes

- This data source requires an Admin API Key (`admin_key` in provider configuration or `OPENAI_ADMIN_KEY` environment variable).
- When looking up by `group_name`, the search is case-insensitive.
