# Project User Resource

The `openai_project_user` resource allows you to manage users within OpenAI projects. This includes adding users to projects, changing their roles, and removing them from projects.

## Example Usage

```hcl
# Add a user to a project with the "member" role
resource "openai_project_user" "example_member" {
  project_id = openai_project.example.id
  user_id    = "user-abc123xyz"
  role       = "member"
}

# Add a user to a project with the "owner" role
resource "openai_project_user" "example_owner" {
  project_id = openai_project.example.id
  user_id    = "user-def456uvw"
  role       = "owner"
  
  # Optionally use a project-specific API key
  api_key    = var.project_api_key
}
```

## Argument Reference

* `project_id` - (Required) The ID of the OpenAI project to which the user will be added.
* `user_id` - (Required) The ID of the user to add to the project. Users must already be members of the organization.
* `role` - (Required) The role to assign to the user in the project. Valid values are `"owner"` or `"member"`. This can be updated after creation.
* `api_key` - (Optional) A custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - A unique identifier for this resource in the format `{project_id}:{user_id}`.
* `email` - The email address of the user.
* `added_at` - The timestamp (in Unix time) when the user was added to the project.

## Import

Project users can be imported using a combination of the project ID and user ID, separated by a colon:

```
$ terraform import openai_project_user.example project_id:user_id
```

For example:

```
$ terraform import openai_project_user.example proj_abc123xyz:user_def456uvw
```

When importing, the provider will make an API call to fetch the current state of the user in the project, including their email, role, and when they were added to the project.

Note: Import requires an API key with sufficient permissions to read project user information.

## Organization User Management

Users must already be members of the organization before they can be added to a project. This section outlines how to manage organization users via the OpenAI API, which is a prerequisite for project user management.

### Listing Organization Users

To list users in your organization:

```bash
curl https://api.openai.com/v1/organization/users?after=user_abc&limit=20 \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "Content-Type: application/json"
```

### Modifying Organization User Roles

To change a user's role at the organization level:

```bash
curl -X POST https://api.openai.com/v1/organization/users/user_abc \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
      "role": "owner"
  }'
```

### Removing Users from the Organization

To delete a user from your organization:

```bash
curl -X DELETE https://api.openai.com/v1/organization/users/user_abc \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "Content-Type: application/json"
```

Note: Organization user management requires an admin API key with appropriate permissions.

## Special Notes and Limitations

### Updating User Roles

Users' roles can be changed between "owner" and "member" by updating the `role` attribute, provided they are not organization owners. The update will be applied via the OpenAI API.

```hcl
resource "openai_project_user" "example" {
  project_id = openai_project.example.id
  user_id    = "user-abc123xyz"
  role       = "owner"  # Can be changed to "member" later
}
```

API endpoint used:
```
POST https://api.openai.com/v1/organization/projects/{project_id}/users/{user_id}
```

### Removing Users from Projects

When a user is removed from a project (by removing the resource from your configuration), the provider will attempt to remove them via the OpenAI API.

API endpoint used:
```
DELETE https://api.openai.com/v1/organization/projects/{project_id}/users/{user_id}
```

### Organization Owner Restrictions

Users who are owners of the organization cannot:
1. Be removed from projects - The API will return an error with code `user_organization_owner`
2. Have their roles changed - The API will return an error with code `user_organization_owner`

In both cases, the provider will log a warning and continue without failing the operation. For organization owners, role changes will be ignored, and when attempting to remove them, they will only be removed from Terraform state (not from the actual project).

### Required Permissions

To manage project users, you must use an API key with appropriate permissions:
- For adding, updating, and removing users: The API key must have the `api.management.write` scope.
- For reading user details: The API key must have the `api.management.read` scope.

Using a project-specific API key may result in permission errors for certain operations.

## Important Notes

- **Organization Membership Requirements**: Users must already be members of the organization before they can be added to a project. The OpenAI API does not support adding users to an organization programmatically. See the [Organization Users Management Limitations](../ORGANIZATION_USERS.md) for more details.

- **User ID Discovery**: You can get the IDs of users in your organization through the OpenAI dashboard or by using the OpenAI API to list organization users (through a manual API call, not via Terraform).

- **Permissions**: To add users to projects, you need an admin API key with the appropriate permissions. 