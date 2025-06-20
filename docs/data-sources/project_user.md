---
page_title: "OpenAI: openai_project_user Data Source"
subcategory: ""
description: |-
  Retrieves information about a user in an OpenAI project.
---

# openai_project_user Data Source

Retrieves information about a specific user in an OpenAI project, including their role and when they were added to the project.

## Example Usage

### Lookup by User ID

```hcl
data "openai_project_user" "example" {
  project_id = "proj_abc123"
  user_id    = "user_abc123"
  
  # Optional: Use a custom API key (must have permission to read project user details)
  api_key = var.openai_api_key
}

# Access user details
output "user_role" {
  value = data.openai_project_user.example.role
}

output "user_added_at" {
  value = data.openai_project_user.example.added_at
}
```

### Lookup by Email Address

```hcl
data "openai_project_user" "by_email" {
  project_id = "proj_abc123"
  email      = "user@example.com"
}

# Access user details including the user ID
output "user_id" {
  value = data.openai_project_user.by_email.user_id
}

output "user_role" {
  value = data.openai_project_user.by_email.role
}
```

## Argument Reference

* `project_id` - (Required) The ID of the project to retrieve the user from.
* `user_id` - (Optional) The ID of the user to retrieve. Either `user_id` or `email` must be provided.
* `email` - (Optional) The email address of the user to retrieve. Either `user_id` or `email` must be provided.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A composite ID uniquely identifying the project user in the format `{project_id}:{user_id}`.
* `user_id` - The ID of the user (useful when looking up by email).
* `email` - The email address of the user.
* `role` - The role of the user in the project. Can be "owner" or "member".
* `added_at` - The timestamp (in Unix time) when the user was added to the project.

## Permission Requirements

To use this data source, your API key must have permissions to read user details in the specified project. If using a custom API key, it must have the appropriate permissions.

## Import

Project user data sources cannot be imported. 