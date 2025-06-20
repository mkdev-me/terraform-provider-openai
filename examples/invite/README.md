# OpenAI Invitation Resource

This example demonstrates how to use the `openai_invite` resource to invite users to your OpenAI organization and assign them to projects.

## Important: Understanding OpenAI API Limitations

The OpenAI API has specific limitations regarding invitations and project assignments:

1. **Invitations don't support project assignments**: The OpenAI API doesn't allow specifying project assignments during invitation creation.
2. **Existing users can't be re-invited**: If a user already exists in the organization, attempting to invite them returns an error.
3. **Project assignments require separate API calls**: Users must be assigned to projects using the project user API after they exist in the organization.

## How the Terraform Provider Handles These Limitations

The `openai_invite` resource has been updated to handle these limitations gracefully:

### For New Users (Not in Organization)
1. Creates an invitation for the user with the specified organization role
2. Waits for the user to appear in the organization (up to 30 seconds)
3. Automatically assigns the user to specified projects

### For Existing Users (Already in Organization)
1. Detects the "user already exists" error
2. Skips the invitation step
3. Proceeds directly to project assignments
4. Updates project roles if they differ from the configuration

## Usage Examples

### Example 1: Basic Invitation with Project Assignments

```hcl
resource "openai_invite" "user" {
  email = "user@example.com"
  role  = "reader"  # Organization-level role

  # These projects will be assigned after the user joins
  projects {
    id   = "proj_xxx"
    role = "owner"  # Project-level role
  }

  projects {
    id   = "proj_yyy"
    role = "member"
  }
}
```

### Example 2: Manual Control (Alternative Approach)

If you need more control over the process, you can use separate resources:

```hcl
# Step 1: Create invitation
resource "openai_invite" "user" {
  email = "user@example.com"
  role  = "reader"
  # No projects specified
}

# Step 2: Get user information
data "openai_organization_user" "user" {
  email = openai_invite.user.email
  depends_on = [openai_invite.user]
}

# Step 3: Assign to projects
resource "openai_project_user" "assignment" {
  project_id = "proj_xxx"
  user_id    = data.openai_organization_user.user.id
  role       = "member"
}
```

## Attributes Reference

### Arguments

* `email` - (Required) The email address of the user to invite
* `role` - (Required) The organization-level role: `owner` or `reader`
* `projects` - (Optional) List of projects to assign the user to
  * `id` - (Required) The project ID
  * `role` - (Required) The project-level role: `owner` or `member`
* `api_key` - (Optional) Custom API key for this resource

### Computed Attributes

* `invite_id` - The invitation ID (or synthetic ID for existing users)
* `status` - The invitation status
* `created_at` - When the invitation was created
* `expires_at` - When the invitation expires
* `user_id` - The user's ID (available after they exist in the organization)

## Important Notes

1. **Project assignments may fail**: If the user hasn't accepted the invitation yet, project assignments will be attempted but may fail. The resource will still be created successfully.

2. **Synthetic IDs for existing users**: When inviting an existing user, the provider creates a synthetic ID (format: `existing-user-{email}`) to track the resource in Terraform state.

3. **Idempotency**: The resource is idempotent - running it multiple times for the same user will not cause errors.

4. **Role updates**: If a user already exists in a project with a different role, the provider will attempt to update their role to match the configuration.

## Troubleshooting

### User already exists error
This is handled automatically by the provider. The invitation step is skipped and the provider proceeds with project assignments.

### Project assignment failures
Check the Terraform logs for warnings about project assignment failures. Common causes:
- User hasn't accepted the invitation yet
- Insufficient permissions
- Invalid project ID

### Organization owners
Organization owners automatically have owner access to all projects and their project roles cannot be changed to "member". 