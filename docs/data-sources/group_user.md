---
page_title: "openai_group_user Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Use this data source to retrieve information about a specific user in an OpenAI organization group.
---

# openai_group_user (Data Source)

Use this data source to retrieve information about a specific user in an OpenAI organization group.

This data source requires an **Admin API Key** for authentication.

## Example Usage

```terraform
# Look up a user in a group by user ID
data "openai_group_user" "by_id" {
  group_id = "group-abc123"
  user_id  = "user-xyz789"
}

# Look up a user in a group by email
data "openai_group_user" "by_email" {
  group_id = "group-abc123"
  email    = "engineer@example.com"
}

# Output user details
output "user_name" {
  value = data.openai_group_user.by_email.user_name
}

output "user_role" {
  value = data.openai_group_user.by_email.role
}

output "user_added_at" {
  value = data.openai_group_user.by_email.added_at
}
```

## Schema

### Required

- `group_id` (String) The ID of the group.

### Optional

- `user_id` (String) The ID of the user. Either user_id or email must be provided.
- `email` (String) The email of the user to search for. Either user_id or email must be provided.

### Read-Only

- `id` (String) The ID of this resource (group_id:user_id).
- `user_name` (String) The name of the user.
- `role` (String) The user's organization role.
- `added_at` (Number) Unix timestamp when the user was added to the organization.

## Notes

- This data source requires an Admin API Key (`admin_key` in provider configuration or `OPENAI_ADMIN_KEY` environment variable).
- Either `user_id` or `email` must be provided to look up the user.
- Email lookups are case-insensitive.
