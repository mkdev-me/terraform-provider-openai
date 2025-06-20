---
page_title: "OpenAI: openai_organization_users Data Source"
subcategory: "Organization Users"
description: |-
  Retrieves information about users in an OpenAI organization.
---

# Data Source: openai_organization_users

Retrieves information about users in an OpenAI organization. This data source allows you to list users, optionally filtering by email address, and paginate through results.

## Example Usage

```hcl
data "openai_organization_users" "all" {
  limit = 10
}

output "organization_users" {
  value = data.openai_organization_users.all.users
}

output "pagination_info" {
  value = {
    first_id = data.openai_organization_users.all.first_id
    last_id  = data.openai_organization_users.all.last_id
    has_more = data.openai_organization_users.all.has_more
  }
}

# Filter users by email
data "openai_organization_users" "filtered" {
  emails = ["user@example.com"]
}

# Pagination example
data "openai_organization_users" "page_two" {
  after = data.openai_organization_users.all.last_id
  limit = 10
}

# List all users in the organization
data "openai_organization_users" "all_users" {
  # Optional: limit the number of users returned (default: 20, max: 100)
  limit = 50
}

# Output the user IDs (useful for finding users who have accepted invitations)
output "user_ids" {
  value = {
    for user in data.openai_organization_users.all_users.users : user.email => user.id
  }
}
```

## Argument Reference

* `api_key` - (Optional) Custom API key to use for this data source. If not provided, the provider's default API key will be used. **Note:** The API key must have organization management permissions.
* `user_id` - (Optional) If provided, returns a list containing only this specific user.
* `after` - (Optional) The ID of the last object returned in a previous API call. Used for pagination.
* `limit` - (Optional) The number of users to return. Default is 20.
* `emails` - (Optional) List of email addresses to filter users by.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `users` - A list of user objects, each containing:
  * `id` - The ID of the user.
  * `email` - The email address of the user.
  * `name` - The name of the user.
  * `role` - The role of the user in the organization. Can be "owner", "member", or "reader".
  * `added_at` - The Unix timestamp when the user was added to the organization.
* `first_id` - The ID of the first user in the returned list. Useful for pagination.
* `last_id` - The ID of the last user in the returned list. Useful for pagination.
* `has_more` - Boolean indicating whether there are more users available beyond this page.

## Permission Requirements

To use this data source, your API key must have organization management permissions (specifically, the `api.management.read` scope). Typically, this requires an owner-level API key or a key specifically granted these permissions.

## Related Resources

* [`openai_organization_user`](organization_user.html) - Data source for retrieving a specific user in the organization.

## Invitation Workflow

This data source is essential for the invite-accept-assign workflow:

1. Send an invitation using the `openai_invite` resource
2. Wait for the user to accept the invitation
3. Use `openai_organization_users` to check if the user is now in the organization and get their user ID
4. Add the user to projects using the `openai_project_user` resource

See the [invite example](/examples/invite/) for a complete workflow demonstration.

## Known Issues

1. When checking for users who have recently accepted invitations, you may need to run `terraform apply` multiple times, as it can take some time for the API to reflect the user's acceptance.
2. If filtering by email, ensure the emails are exact matches to what's in your OpenAI organization.

## Pagination

If you have more than `limit` users in your organization, you can paginate through the results by using the `after` parameter with the `last_id` from the previous response. 