# OpenAI Organization Users Module

This module provides a standardized way to retrieve information about users in your OpenAI organization.

## Important: Admin Permissions Required

This module requires an API key with admin permissions, specifically the `api.management.read` scope.

## Features

- Retrieve detailed information about a specific user
- List all users in your organization with filtering options
- Support for pagination through large user lists
- Filter users by email addresses
- Customizable output formats

## Usage

### Single User Mode

```hcl
module "organization_user" {
  source = "../../modules/organization_users"
  
  # Get a single user
  user_id = "user-abc123"  # Replace with actual user ID
  
  # Optional custom API key with admin permissions
  api_key = var.openai_admin_api_key
}

output "user_details" {
  value = {
    id    = module.organization_user.user.id
    email = module.organization_user.user.email
    name  = module.organization_user.user.name
    role  = module.organization_user.user.role
  }
}
```

### List Users Mode

```hcl
module "organization_users" {
  source = "../../modules/organization_users"
  
  # List mode enabled
  list_mode = true
  
  # Optional: Filtering and pagination
  limit  = 50
  emails = ["user@example.com"]
  after  = "user-xyz789"  # For pagination
  
  # Optional custom API key with admin permissions
  api_key = var.openai_admin_api_key
}

output "owner_users" {
  value = module.organization_users.owners
}

output "member_users" {
  value = module.organization_users.members
}

output "reader_users" {
  value = module.organization_users.readers
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| list_mode | Whether to operate in list mode or single user mode | `bool` | `false` | no |
| user_id | The ID of the user to retrieve (required in single user mode) | `string` | `null` | no |
| after | A cursor for pagination (list mode only) | `string` | `null` | no |
| limit | The number of users to return (list mode only) | `number` | `20` | no |
| emails | List of email addresses to filter by (list mode only) | `list(string)` | `[]` | no |
| api_key | Custom API key to use for this module | `string` | `null` | no |

## Outputs

### Single User Mode

| Name | Description | Type |
|------|-------------|------|
| user | Complete user object | `object` |
| user_id | The ID of the user | `string` |
| email | The email address of the user | `string` |
| name | The name of the user | `string` |
| role | The role of the user in the organization | `string` |
| added_at | The Unix timestamp when the user was added | `number` |

### List Mode

| Name | Description | Type |
|------|-------------|------|
| all_users | List of all users with complete details | `list(object)` |
| user_count | Total number of users in the result | `number` |
| owners | List of users with the 'owner' role | `list(object)` |
| members | List of users with the 'member' role | `list(object)` |
| readers | List of users with the 'reader' role | `list(object)` |
| first_id | ID of the first user in the result | `string` |
| last_id | ID of the last user in the result | `string` |
| has_more | Whether there are more users available | `bool` |

## Related Resources

Check out these related modules:

- [project_user](../project_user) - Manage users within OpenAI projects
- [invite](../invite) - Send invitations to users to join your organization

## Examples

See the [organization_users example](../../examples/organization_users) for a complete working example. 