---
page_title: "OpenAI: openai_organization_user"
subcategory: "Organization Users"
description: |-
  Retrieves information about a specific user in an OpenAI organization.
---

# Data Source: openai_organization_user

Retrieves information about a specific user in an OpenAI organization. This data source allows you to get details about a user, including their email address, name, role, and when they were added to the organization.

## Example Usage

### Lookup by User ID

```hcl
data "openai_organization_user" "example" {
  user_id = "user_abc123"
}

output "user_details" {
  value = {
    id       = data.openai_organization_user.example.id
    email    = data.openai_organization_user.example.email
    name     = data.openai_organization_user.example.name
    role     = data.openai_organization_user.example.role
    added_at = data.openai_organization_user.example.added_at
  }
}
```

### Lookup by Email Address

```hcl
data "openai_organization_user" "by_email" {
  email = "user@example.com"
}

output "user_id_from_email" {
  value = data.openai_organization_user.by_email.id
}
```

## Argument Reference

* `user_id` - (Optional) The ID of the user to retrieve. Either `user_id` or `email` must be provided.
* `email` - (Optional) The email address of the user to retrieve. Either `user_id` or `email` must be provided.
* `api_key` - (Optional) Custom API key to use for this data source. If not provided, the provider's default API key will be used. **Note:** The API key must have organization management permissions.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the user.
* `email` - The email address of the user.
* `name` - The name of the user.
* `role` - The role of the user in the organization. Can be "owner", "member", or "reader".
* `added_at` - The Unix timestamp when the user was added to the organization.

## Permission Requirements

To use this data source, your API key must have organization management permissions (specifically, the `api.management.read` scope). Typically, this requires an owner-level API key or a key specifically granted these permissions.

## Related Resources

* [`openai_organization_users`](organization_users.html) - Data source for retrieving multiple users in the organization.

variable "openai_admin_api_key" {
  description = "OpenAI Admin API Key with organization management permissions"
  type        = string
  sensitive   = true
} 