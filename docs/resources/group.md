---
page_title: "openai_group Resource - terraform-provider-openai"
subcategory: ""
description: |-
  Manages an OpenAI organization group.
---

# openai_group (Resource)

Manages an OpenAI organization group. Groups are collections of users that can be assigned roles at the organization or project level.

This resource requires an **Admin API Key** for authentication.

## Example Usage

```terraform
# Create a group in the organization
resource "openai_group" "engineering" {
  name = "Engineering Team"
}

# Create another group
resource "openai_group" "support" {
  name = "Support Team"
}

# Output the group IDs
output "engineering_group_id" {
  value = openai_group.engineering.id
}

output "support_group_id" {
  value = openai_group.support.id
}

# Check if groups are SCIM managed
output "engineering_is_scim_managed" {
  value = openai_group.engineering.is_scim_managed
}
```

## Schema

### Required

- `name` (String) The display name of the group.

### Read-Only

- `id` (String) The identifier of the group.
- `created_at` (Number) Unix timestamp (in seconds) when the group was created.
- `is_scim_managed` (Boolean) Whether the group is managed through SCIM and controlled by your identity provider.

## Import

Groups can be imported using the group ID:

```shell
terraform import openai_group.example group-abc123
```

## Notes

- This resource requires an Admin API Key (`admin_key` in provider configuration or `OPENAI_ADMIN_KEY` environment variable).
- Groups managed through SCIM (`is_scim_managed = true`) may have restrictions on modifications.
- After creating a group, you can add users using the `openai_group_user` resource.
- Groups can be assigned to projects using the `openai_project_group` resource.
