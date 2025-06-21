# Managing User Roles in OpenAI with Terraform

This directory contains examples of how to manage OpenAI users and their roles using Terraform. It includes examples for both user management resources and data sources.

## Example Files

This example contains three separate Terraform configuration files:

1. `main.tf` - Demonstrates how to use the user management resources
2. `data_sources.tf` - Demonstrates how to use the user data sources
3. `email_lookup.tf` - Demonstrates how to use email addresses instead of user IDs with data sources

## Resources

1. `openai_project_user` - Manages users within specific projects

## Data Sources

1. `openai_organization_user` - Retrieves information about a specific user in your organization
2. `openai_organization_users` - Retrieves information about all users in your organization
3. `openai_project_user` - Retrieves information about a user within a specific project
4. `openai_project_users` - Retrieves information about all users within a specific project

## User Management Workflow

Important: Users must be created through the OpenAI dashboard first before they can be managed via Terraform.

```OpenAI Dashboard (Create User) → Get User ID → Terraform (Manage Roles)
```

## Prerequisites

- An OpenAI account with administrator privileges
- An OpenAI Admin API key with appropriate permissions
- Terraform >= 0.13.0

## Finding User IDs

There are two ways to find user IDs:

### Method 1: Using the OpenAI API

Run the following command to list all users in your organization:

```bash
curl https://api.openai.com/v1/organization/users \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "Content-Type: application/json"
```

This will return all users in your organization along with their IDs (format: `user-XXXXXXXXXXXXXXXXXXXX`).

### Method 2: Using the OpenAI Dashboard

1. Go to the OpenAI dashboard
2. Navigate to Settings > Members
3. Use browser developer tools to inspect the network requests
4. Look for the user ID in the API responses

## Setup

1. Create a `terraform.tfvars` file with your credentials:

```hcl
openai_admin_key = "sk-xxxx"  # Your OpenAI Admin API key
```

2. Update the example files with real user IDs:

```hcl
# In main.tf - Replace with actual user IDs from your organization
data "openai_organization_user" "user" {
  user_id = "user-yatSd6LuWvgeoqZbd89xzPlJ"  # Replace with real user ID
}

# In data_sources.tf - Replace with actual user IDs from your organization
resource "openai_project_user" "data_source_user" {
  project_id = openai_project.data_source_example.id
  user_id    = "user-abc123xyz"  # Replace with real user ID
  role       = "member"
}
```

## Usage

1. Initialize Terraform:
```bash
terraform init
```

2. Apply the configuration:
```bash
terraform apply -var="openai_admin_key=$OPENAI_ADMIN_KEY"
```

## Example 1: User Management Resources (main.tf)

This example demonstrates how to use the `openai_project_user` resource to manage user permissions in projects.

### Resource: openai_project_user

This resource manages user roles within a specific project.

**Example:**
```hcl
resource "openai_project" "example" {
  name = "Example Project for User Management"
}

resource "openai_project_user" "project_user" {
  project_id = openai_project.example.id
  user_id    = data.openai_organization_user.user.id
  role       = "member"  # Can be "owner" or "member"
  
  depends_on = [openai_project.example]
}
```

### Data Source: openai_organization_user

This data source retrieves information about a specific user in your organization.

**Example:**
```hcl
data "openai_organization_user" "user" {
  user_id = "user-yatSd6LuWvgeoqZbd89xzPlJ"  # Must be a valid user ID
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

## Example 2: User Data Sources (data_sources.tf)

This example demonstrates how to use various user-related data sources to retrieve information.

### Data Source: openai_organization_user

This data source retrieves information about a specific user in your organization.

**Example using user_id:**
```hcl
data "openai_organization_user" "user_info" {
  user_id = "user-yatSd6LuWvgeoqZbd89xzPlJ"  # Must be a valid user ID
}

output "data_source_user_email" {
  value = data.openai_organization_user.user_info.email
}

output "data_source_user_role" {
  value = data.openai_organization_user.user_info.role
}
```

**Example using email:**
```hcl
data "openai_organization_user" "by_email" {
  email = "user@example.com"  # Must be a valid email address in your organization
}

output "user_id_from_email" {
  value = data.openai_organization_user.by_email.user_id
}
```

### Data Source: openai_organization_users

This data source retrieves information about all users in your organization.

**Example:**
```hcl
data "openai_organization_users" "all_org_users" {
  limit = 50  # Number of users to return (1-100)
}

output "organization_owner_count" {
  value = length([
    for user in data.openai_organization_users.all_org_users.users : 
    user if user.role == "owner"
  ])
}

output "organization_users_count" {
  value = length(data.openai_organization_users.all_org_users.users)
}
```

### Data Source: openai_project_user

This data source retrieves information about a user within a specific project.

**Example using user_id:**
```hcl
data "openai_project_user" "project_user_info" {
  project_id = openai_project.data_source_example.id
  user_id    = "user-yatSd6LuWvgeoqZbd89xzPlJ"  # Must be a valid user ID
  
  depends_on = [openai_project_user.data_source_user]  # Make sure the user is in the project first
}

output "data_source_project_user_role" {
  value = data.openai_project_user.project_user_info.role
}
```

**Example using email:**
```hcl
data "openai_project_user" "by_email" {
  project_id = openai_project.data_source_example.id
  email      = "user@example.com"  # Must be a valid email in your project
}

output "project_user_id_from_email" {
  value = data.openai_project_user.by_email.user_id
}
```

### Data Source: openai_project_users

This data source retrieves information about all users within a specific project.

**Important: Never output the entire data source directly**

```hcl
# INCORRECT - Will cause sensitive value errors
output "all_data" {
  value = data.openai_project_users.all_project_users  # DON'T DO THIS
}

# CORRECT - Access specific fields instead
output "users" {
  value = data.openai_project_users.all_project_users.users
}

output "user_ids" {
  value = data.openai_project_users.all_project_users.user_ids
}
```

This limitation exists because the data source contains internal fields that are marked as sensitive (like API keys) which cannot be exposed in outputs. Even though you might only want the user data, Terraform will detect the sensitive fields when you reference the entire data source.

Each user in the list contains the following attributes:
- `id` - The unique identifier for the user
- `email` - The email address of the user
- `role` - The role of the user in the project (owner or member)
- `added_at` - Unix timestamp when the user was added to the project

## Example 3: Email Lookup (email_lookup.tf)

This example demonstrates how to look up users by email instead of user IDs. This is particularly useful when you know users' email addresses but not their OpenAI user IDs.

### Using Email with Organization User Data Source

**Example:**
```hcl
data "openai_organization_user" "by_email" {
  email   = "user@example.com"  # Replace with a real email from your organization
}

output "org_user_by_email_id" {
  value = data.openai_organization_user.by_email.user_id
}
```

### Using Email with Project User Data Source

**Example:**
```hcl
data "openai_project_user" "by_email" {
  project_id = openai_project.data_source_example.id
  email      = "user@example.com"  # Replace with a real email from your project
}

output "project_user_by_email_id" {
  value = data.openai_project_user.by_email.user_id
}
```

### Practical Example: Find a User by Email and Use Their ID

```hcl
# First, find the user by email
data "openai_organization_user" "find_by_email" {
  email = "teammate@example.com"
}

# Then use their ID in a resource
resource "openai_project_user" "add_teammate" {
  project_id = openai_project.my_project.id
  user_id    = data.openai_organization_user.find_by_email.user_id
  role       = "member"
}
```

## Importing Resources

You can import existing resources into Terraform state so they can be managed using Terraform.

### Importing Project Users

To import an existing project user:

```bash
terraform import openai_project_user.example "project_id:user_id"
```

Example:

```bash
terraform import -var="openai_admin_key=$OPENAI_ADMIN_API" \
  openai_project_user.data_source_user "proj_8nlneYhywC7UZHXdesluczOo:user-yatSd6LuWvgeoqZbd89xzPlJ"
```

This command will import the specified user in the project into your Terraform state, and you can then manage it with Terraform. The provider will make an API call to fetch the current state of the user in the project, including their email, role, and when they were added.

## Related Resources

* [Organization Users Example](../organization_users) - Example for retrieving and managing organization users
* [Invite Example](../invite) - Example for inviting users to your organization

## Role Management and API Integration

The OpenAI provider handles role management in projects with the following behavior:

1. **Configuration is Always the Source of Truth**: Terraform uses the role specified in your configuration as the source of truth. If the API returns a different role during a `terraform plan` or `terraform apply`, Terraform will detect this as a difference and attempt to update the role to match your configuration.

2. **API Calls for Role Updates**: When Terraform detects that a role needs to be updated, it makes a direct API call to OpenAI using:
   ```
   POST https://api.openai.com/v1/organization/projects/{project_id}/users/{user_id}
   ```
   with a request body containing the role from your configuration:
   ```json
   { "role": "member" }
   ```

3. **Handling API Role Drift**: If you change a user's role outside of Terraform (e.g., in the OpenAI dashboard), the provider will:
   - Log a warning about the role mismatch
   - Maintain your Terraform configuration as the source of truth
   - Attempt to update the role to match your configuration on the next apply

4. **Importing Existing State**: When importing users with existing roles:
   ```bash
   terraform import openai_project_user.example "proj_abc123:user-xyz789"
   ```

   Make sure your configuration matches what you want (not necessarily what exists in the API):
   ```hcl
   resource "openai_project_user" "example" {
     project_id = "proj_abc123"
     user_id    = "user-xyz789"
     role       = "member"  # This will be applied even if the current role is "owner"
   }
   