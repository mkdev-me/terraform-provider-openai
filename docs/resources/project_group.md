---
page_title: "openai_project_group Resource - terraform-provider-openai"
subcategory: ""
description: |-
  Manages a group's access to an OpenAI Project. Groups are collections of users that can be synced from an identity provider via SCIM.
---

# openai_project_group (Resource)

Manages a group's access to an OpenAI Project. Groups are collections of users that can be synced from an identity provider via SCIM.

This resource requires an **Admin API Key** for authentication.

## Example Usage

```terraform
# First, create a project
resource "openai_project" "development" {
  name = "Development Environment"
}

# Add a group to the project
resource "openai_project_group" "dev_team" {
  project_id = openai_project.development.id
  group_id   = "group-123abc"  # Replace with actual group ID from your IdP
  role       = "owner"
}

# Add another group with member role
resource "openai_project_group" "engineering" {
  project_id = openai_project.development.id
  group_id   = "group-456def"  # Replace with actual group ID from your IdP
  role       = "member"
}

# Output group details
output "dev_team_group_name" {
  value       = openai_project_group.dev_team.group_name
  description = "Name of the development team group"
}
```

## Schema

### Required

- `project_id` (String) The ID of the project.
- `group_id` (String) The ID of the group to add to the project.
- `role` (String) The role of the group in the project (e.g., 'owner' or 'member').

### Read-Only

- `id` (String) The identifier of the project group (project_id:group_id).
- `group_name` (String) The display name of the group.
- `created_at` (Number) The timestamp when the group was added to the project.

## Import

Project groups can be imported using the composite ID format `project_id:group_id`:

```shell
terraform import openai_project_group.example proj-abc123:group-xyz789
```

## Notes

- Groups are synced from your identity provider via SCIM. You cannot create groups through this provider; you can only manage which groups have access to which projects.
- The OpenAI API does not support updating the role of a group in a project. To change a group's role, Terraform will destroy and recreate the resource.
- All operations require an Admin API Key (`admin_key` in provider configuration or `OPENAI_ADMIN_KEY` environment variable).
