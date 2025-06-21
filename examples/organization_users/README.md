# OpenAI Organization Users Example

This example demonstrates how to use the `openai_organization_user` and `openai_organization_users` data sources to retrieve information about users in your OpenAI organization.

## Important: Admin Permissions Required

This example requires an API key with admin permissions, specifically the `api.management.read` scope. Regular API keys created through the OpenAI dashboard typically don't have these permissions.

## Prerequisites

- Terraform 0.13+
- OpenAI API key with admin permissions
- Your organization must have at least one user

## Configuration Files

This example includes:

- `main.tf`: Main configuration file with provider setup, data sources, and outputs
- `variables.tf`: Variable definitions for the admin API key
- `email_lookup.tf`: Demonstrates how to look up users by email instead of user ID

## Usage

1. Set up your API key:

```bash
export TF_VAR_openai_admin_api_key="your-admin-api-key"
```

2. Initialize Terraform:

```bash
terraform init
```

3. Apply the configuration:

```bash
terraform apply
```

4. Review the outputs showing user information:

```bash
terraform output
```

## Looking Up Users By Email

As of version 0.X.X, the provider now supports looking up users by email address instead of requiring user IDs:

```hcl
# Look up a user by email address
data "openai_organization_user" "by_email" {
  email   = "user@example.com"  # Replace with a real email from your organization
}

# The data source will return the user's ID, which can then be used in other resources
output "user_id" {
  value = data.openai_organization_user.by_email.user_id
}
```

This is particularly useful when:
- You know users' email addresses but not their OpenAI user IDs
- You want to automate processes using email addresses that are more human-readable
- You're building workflows where emails are the primary identifier for users

## Permission Issues

If you encounter permission errors such as:

```
error listing organization users: error listing users: API error: You have insufficient permissions for this operation. Missing scopes: api.management.read.
```

You need to:
1. Use an owner-level API key, or
2. Use a key specifically granted the `api.management.read` scope

## Example Output

```
Outputs:

specific_user = {
  "email" = "user@example.com"
  "id" = "user-abc123"
  "role" = "owner"
}
organization_user_count = 5
organization_users = [
  {
    "email" = "user1@example.com"
    "id" = "user-abc123"
    "role" = "owner"
  },
  {
    "email" = "user2@example.com"
    "id" = "user-def456"
    "role" = "member"
  },
  ...
]
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `openai_admin_api_key` | OpenAI Admin API Key with organization management permissions | (required) |
| `enable_organization_users` | Whether to enable organization user functionality | `true` |

## Resources Created

This example creates:
- One `openai_project` resource called "Organization Users Demo"
- Project user associations if the organization user functionality is enabled

## Related Examples

- [Project User Example](../project_user) - Managing users within projects
- [Invite Example](../invite) - Inviting users to your organization 