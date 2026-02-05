---
page_title: "openai_group_user Resource - terraform-provider-openai"
subcategory: ""
description: |-
  Manages a user's membership in an OpenAI organization group.
---

# openai_group_user (Resource)

Manages a user's membership in an OpenAI organization group. Use this resource to add users to groups.

This resource requires an **Admin API Key** for authentication.

## Example Usage

```terraform
# Add a user to a group
resource "openai_group_user" "engineer" {
  group_id = openai_group.engineering.id
  user_id  = "user-abc123"
}

# Reference an existing group and add multiple users
resource "openai_group_user" "support_agent_1" {
  group_id = "group-xyz789"
  user_id  = "user-def456"
}

resource "openai_group_user" "support_agent_2" {
  group_id = "group-xyz789"
  user_id  = "user-ghi789"
}

# Output user details
output "engineer_email" {
  value = openai_group_user.engineer.email
}

output "engineer_name" {
  value = openai_group_user.engineer.user_name
}
```

## Schema

### Required

- `group_id` (String) The ID of the group.
- `user_id` (String) The ID of the user to add to the group.

### Read-Only

- `id` (String) The identifier of the group user (group_id:user_id).
- `user_name` (String) The name of the user.
- `email` (String) The email of the user.
- `role` (String) The user's organization role.
- `added_at` (Number) Unix timestamp when the user was added to the organization.

## Import

Group user memberships can be imported using the format `group_id:user_id`:

```shell
terraform import openai_group_user.example group-abc123:user-xyz789
```

## Notes

- This resource requires an Admin API Key (`admin_key` in provider configuration or `OPENAI_ADMIN_KEY` environment variable).
- Users must already exist in the organization before they can be added to a group.
- Removing this resource will remove the user from the group, not delete the user from the organization.
