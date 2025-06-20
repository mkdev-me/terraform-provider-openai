---
page_title: "OpenAI: openai_project_users Data Source"
subcategory: "Project Users"
description: |-
  Retrieves a list of all users in a specific OpenAI project.
---

# Data Source: openai_project_users

Retrieves a list of all users in a specific OpenAI project, including their roles, email addresses, and when they were added to the project.

-> **Note:** This data source requires an OpenAI admin API key with appropriate permissions.

## Example Usage

```hcl
data "openai_project_users" "all_users" {
  project_id = "proj_abc123xyz"
}

output "user_count" {
  value = length(data.openai_project_users.all_users.users)
}

output "owner_emails" {
  value = [for user in data.openai_project_users.all_users.users : user.email if user.role == "owner"]
}
```

## Argument Reference

* `project_id` - (Required) The ID of the project to retrieve users from (format: `proj_abc123xyz`).
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The project ID.
* `users` - A list of users in the project. Each user contains:
  * `id` - The ID of the user (format: `user_abc123xyz`).
  * `email` - The email address of the user.
  * `role` - The role of the user in the project. Can be "owner" or "member".
  * `added_at` - The timestamp (in Unix time) when the user was added to the project.

## Use Cases

This data source is particularly useful for:

* Auditing user access to projects
* Creating dynamic configurations based on project membership
* Determining how many owners/members a project has
* Generating reports on project access
* Automating user management workflows

## Example: Finding Project Owners

```hcl
data "openai_project_users" "all_users" {
  project_id = "proj_abc123xyz"
}

locals {
  project_owners = [
    for user in data.openai_project_users.all_users.users:
    user if user.role == "owner"
  ]
}

output "owner_count" {
  value = length(local.project_owners)
}

output "owner_details" {
  value = local.project_owners
}
```

## Permissions Required

This data source requires an OpenAI API key with the following permissions:

* Project role of Admin or Owner
* If using an organization-level key, it must have admin privileges 