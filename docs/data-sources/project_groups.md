---
page_title: "openai_project_groups Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Use this data source to retrieve a list of all groups in a specific OpenAI project.
---

# openai_project_groups (Data Source)

Use this data source to retrieve a list of all groups in a specific OpenAI project.

This data source requires an **Admin API Key** for authentication.

## Example Usage

```terraform
# List all groups in a specific project
data "openai_project_groups" "production_groups" {
  project_id = "proj-abc123"
}

# Output total project group count
output "total_project_groups" {
  value = data.openai_project_groups.production_groups.group_count
}

# Output all group IDs
output "all_group_ids" {
  value = data.openai_project_groups.production_groups.group_ids
}

# Output groups with owner role
output "owner_group_ids" {
  value = data.openai_project_groups.production_groups.owner_ids
}

# Output groups with member role
output "member_group_ids" {
  value = data.openai_project_groups.production_groups.member_ids
}

# Output detailed group information
output "group_details" {
  value = [
    for group in data.openai_project_groups.production_groups.groups : {
      id         = group.group_id
      name       = group.group_name
      role       = group.role
      created_at = group.created_at
    }
  ]
}
```

## Schema

### Required

- `project_id` (String) The ID of the project to retrieve groups from.

### Read-Only

- `id` (String) The ID of this resource.
- `groups` (List of Object) List of groups in the project. (see [below for nested schema](#nestedatt--groups))
- `group_ids` (List of String) List of group IDs in the project.
- `group_count` (Number) Number of groups in the project.
- `owner_ids` (List of String) List of group IDs with owner role.
- `member_ids` (List of String) List of group IDs with member role.

<a id="nestedatt--groups"></a>
### Nested Schema for `groups`

Read-Only:

- `group_id` (String) The ID of the group.
- `group_name` (String) The display name of the group.
- `role` (String) The role of the group (e.g., 'owner' or 'member').
- `created_at` (Number) Timestamp when the group was added to the project.

## Notes

- This data source requires an Admin API Key (`admin_key` in provider configuration or `OPENAI_ADMIN_KEY` environment variable).
- The data source automatically handles pagination, returning all groups in the project regardless of the number.
