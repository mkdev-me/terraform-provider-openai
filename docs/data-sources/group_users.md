---
page_title: "openai_group_users Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Use this data source to retrieve a list of all users in an OpenAI organization group.
---

# openai_group_users (Data Source)

Use this data source to retrieve a list of all users in an OpenAI organization group.

This data source requires an **Admin API Key** for authentication.

## Example Usage

```terraform
# List all users in a group
data "openai_group_users" "engineering" {
  group_id = "group-abc123"
}

# Output total user count
output "total_users" {
  value = data.openai_group_users.engineering.user_count
}

# Output all user IDs
output "all_user_ids" {
  value = data.openai_group_users.engineering.user_ids
}

# Output detailed user information
output "user_details" {
  value = [
    for user in data.openai_group_users.engineering.users : {
      id       = user.user_id
      name     = user.user_name
      email    = user.email
      role     = user.role
      added_at = user.added_at
    }
  ]
}
```

## Schema

### Required

- `group_id` (String) The ID of the group to retrieve users from.

### Read-Only

- `id` (String) The ID of this resource.
- `users` (List of Object) List of users in the group. (see [below for nested schema](#nestedatt--users))
- `user_ids` (List of String) List of user IDs in the group.
- `user_count` (Number) Number of users in the group.

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `user_id` (String) The ID of the user.
- `user_name` (String) The name of the user.
- `email` (String) The email of the user.
- `role` (String) The user's organization role.
- `added_at` (Number) Unix timestamp when the user was added to the organization.

## Notes

- This data source requires an Admin API Key (`admin_key` in provider configuration or `OPENAI_ADMIN_KEY` environment variable).
- The data source automatically handles pagination, returning all users in the group regardless of the number.
