# OpenAI Project User Module

This Terraform module provides a way to manage users within OpenAI projects. It supports both adding individual users to projects and retrieving information about all users in a project.

## Important Notes

- **This module requires that users already exist in your OpenAI organization**
- Users must be invited through the OpenAI dashboard before they can be managed via Terraform
- The API uses `POST` method for updating user roles (not `PATCH` as previously documented)
- User IDs must be obtained using the API before they can be referenced in Terraform

## Finding User IDs

There are two ways to find user IDs:

### Method 1: Using the OpenAI API

Run the following command to list all users in your organization:

```bash
curl https://api.openai.com/v1/organization/users \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "Content-Type: application/json"
```

Look for the user's email and note their ID (format: `user-XXXXXXXXXXXXXXXXXXXX`).

### Method 2: Using the OpenAI Dashboard

1. Go to the OpenAI dashboard
2. Navigate to Settings > Members
3. Use browser developer tools to inspect the network requests
4. Look for the user ID in the API responses

## Usage

### Adding a User to a Project

```hcl
module "project_user" {
  source     = "../../modules/project_user"
  project_id = "proj_abc123xyz"  # OpenAI project ID
  user_id    = "user_abc123xyz"  # OpenAI user ID
  role       = "member"          # Either "member" or "owner"
  
  # Optional: Custom admin API key
  openai_admin_key = var.openai_admin_key
}

output "user_email" {
  value = module.project_user.email
}
```

### Retrieving All Users in a Project

```hcl
module "project_users" {
  source     = "../../modules/project_user"
  project_id = "proj_abc123xyz"
  list_mode  = true  # Enable list mode to get all project users
  
  # Optional: Custom admin API key
  openai_admin_key = var.openai_admin_key
}

output "user_count" {
  value = module.project_users.user_count
}

output "all_project_users" {
  value = module.project_users.all_users
}

output "project_owners" {
  value = module.project_users.project_owners
}
```

## Input Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| `project_id` | The ID of the project where the user will be added | `string` | n/a | yes |
| `user_id` | The ID of the user to add to the project (format: user-abc123) | `string` | `null` | no (required when `list_mode` is false) |
| `role` | The role to assign to the user (owner or member) | `string` | `"member"` | no |
| `openai_admin_key` | Custom API key for authentication (must have admin permissions) | `string` | `null` | no |
| `list_mode` | When true, retrieves all users in a project instead of managing a single user | `bool` | `false` | no |

## Output Values

### Single User Mode (list_mode = false)

| Name | Description |
|------|-------------|
| `id` | The unique identifier for the project user |
| `email` | The email address of the user |
| `added_at` | The timestamp when the user was added to the project |
| `role` | The role assigned to the user |

### List Mode (list_mode = true)

| Name | Description |
|------|-------------|
| `all_users` | List of all users in the project |
| `user_count` | Number of users in the project |
| `project_owners` | List of users with owner role in the project |
| `project_members` | List of users with member role in the project |

## Examples

### Add a User as a Project Owner

```hcl
module "project_owner" {
  source     = "../../modules/project_user"
  project_id = openai_project.my_project.id
  user_id    = "user-abc123xyz"
  role       = "owner"
}
```

### List All Project Users and Get Statistics

```hcl
module "project_users" {
  source     = "../../modules/project_user"
  project_id = openai_project.my_project.id
  list_mode  = true
}

locals {
  owner_count = length(module.project_users.project_owners)
  member_count = length(module.project_users.project_members)
  total_users = module.project_users.user_count
}

output "project_user_stats" {
  value = {
    total_users = local.total_users
    owners = local.owner_count
    members = local.member_count
    owner_percentage = "${(local.owner_count / local.total_users) * 100}%"
  }
}
```

## Requirements

* Terraform >= 0.13
* OpenAI Provider v1.0.0 or higher
* An OpenAI API key with admin permissions

## Limitations

* Adding a user to a project requires that the user already exists in your OpenAI organization
* The user_id must be in the format `user-abc123xyz` (the full OpenAI user ID)
* For list mode, the API key must have admin permissions for the organization or project

## Resource Management Workflow

1. First, ensure users have been invited through the OpenAI dashboard
2. Obtain user IDs using one of the methods described above
3. Create an OpenAI project 
4. Use this module to add users to the project with appropriate roles
5. Create appropriate dependencies between resources

Example workflow:

```hcl
# Create a project
resource "openai_project" "example" {
  name        = "Example Project"
  description = "A project created via Terraform"
}

# Add a user as a member
module "project_member" {
  source     = "../../modules/project_user"
  project_id = openai_project.example.id
  user_id    = "user-yatSd6LuWvgeoqZbd89xzPlJ"
  role       = "member"
  
  depends_on = [openai_project.example]
}

# Add a user as an owner
module "project_owner" {
  source     = "../../modules/project_user"
  project_id = openai_project.example.id
  user_id    = "user-AnotherUserIdHere"
  role       = "owner"
  
  depends_on = [openai_project.example]
}
```

## Importing Existing Project Users

You can import existing project users into Terraform state to manage them with this module:

```bash
terraform import module.project_user.openai_project_user.user "proj_abc123:user_def456"
```

Where:
- `module.project_user` is the module instance name
- `proj_abc123` is your project ID 
- `user_def456` is the user ID

## Working with Organization Users

To retrieve information about users at the organization level (not just project-specific), use the `openai_organization_user` data source:

```hcl
data "openai_organization_user" "user" {
  user_id = "user-abc123xyz"  # Must be a valid user ID
  api_key = var.openai_admin_key  # Must have api.management.read scope
}

output "user_details" {
  value = {
    id    = data.openai_organization_user.user.id
    email = data.openai_organization_user.user.email
    name  = data.openai_organization_user.user.name
    role  = data.openai_organization_user.user.role
  }
}
```

For listing all users in your organization, use the `openai_organization_users` data source:

```hcl
data "openai_organization_users" "all" {
  limit = 50  # Number of users to return (1-100)
  api_key = var.openai_admin_key  # Must have api.management.read scope
}

output "organization_users_count" {
  value = length(data.openai_organization_users.all.users)
}
```

See the [organization_users module](../organization_users) for a more comprehensive interface to organization user management.

## Technical Details and Troubleshooting

### API Method Change

The OpenAI API requires a `POST` method (not `PATCH`) for user role updates. This has been fixed in the latest version of the provider.

### Common Errors

- **404 Not Found**: Ensure the user ID exists and is correctly formatted
- **403 Forbidden**: Check your API key permissions
- **400 Bad Request**: Verify that the role is either "owner" or "member"

### Missing Users

If a user doesn't appear in the API response, make sure:
1. They have accepted their invitation
2. You're using an admin API key with sufficient permissions
3. The user is part of your organization

## License

This module is licensed under the terms of the license file included in the parent repository. 