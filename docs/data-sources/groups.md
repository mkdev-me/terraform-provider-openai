---
page_title: "openai_groups Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Use this data source to retrieve a list of all groups in your OpenAI organization.
---

# openai_groups (Data Source)

Use this data source to retrieve a list of all groups in your OpenAI organization.

This data source requires an **Admin API Key** for authentication.

## Example Usage

```terraform
# List all groups in the organization
data "openai_groups" "all" {}

# Output total group count
output "total_groups" {
  value = data.openai_groups.all.group_count
}

# Output all group IDs
output "all_group_ids" {
  value = data.openai_groups.all.group_ids
}

# Output detailed group information
output "group_details" {
  value = [
    for group in data.openai_groups.all.groups : {
      id            = group.id
      name          = group.name
      is_scim       = group.is_scim_managed
      created_at    = group.created_at
    }
  ]
}
```

## Schema

### Read-Only

- `id` (String) The ID of this resource.
- `groups` (List of Object) List of groups in the organization. (see [below for nested schema](#nestedatt--groups))
- `group_ids` (List of String) List of group IDs in the organization.
- `group_count` (Number) Number of groups in the organization.

<a id="nestedatt--groups"></a>
### Nested Schema for `groups`

Read-Only:

- `id` (String) The ID of the group.
- `name` (String) The display name of the group.
- `created_at` (Number) Unix timestamp (in seconds) when the group was created.
- `is_scim_managed` (Boolean) Whether the group is managed through SCIM.

## Notes

- This data source requires an Admin API Key (`admin_key` in provider configuration or `OPENAI_ADMIN_KEY` environment variable).
- The data source automatically handles pagination, returning all groups regardless of the number.
