---
page_title: "OpenAI: openai_invite Resource"
subcategory: ""
description: |-
  Manages invitations to an OpenAI organization.
---

# openai_invite Resource

The `openai_invite` resource allows you to create and manage invitations to your OpenAI organization. This resource can be used to invite new users and specify their roles at the organization level.

~> **IMPORTANT NOTE:** The OpenAI API does not correctly apply project assignments via invitations. To add users to projects, you must wait for them to accept the invitation and then assign them to projects using the `openai_project_user` resource.

## Example Usage

```hcl
# Invite a user to the organization
resource "openai_invite" "basic_invite" {
  email = "newuser@example.com"
  role  = "reader"
}

# Using a custom API key
resource "openai_invite" "custom_key_invite" {
  email   = "developer@example.com"
  role    = "reader"
  api_key = var.openai_admin_key  # Optional: Use a custom API key
}

# Check the status of an invitation
output "invite_status" {
  value = openai_invite.basic_invite.status
}
```

## Recommended Project Assignment Workflow

Since the OpenAI API does not correctly apply project assignments through invitations, we recommend the following workflow:

```hcl
# Step 1: Invite user to the organization
resource "openai_invite" "org_invite" {
  email = "newuser@example.com"
  role  = "reader"
}

# Step 2: Fetch user_id via data source (after the user accepts the invitation)
data "openai_organization_user" "new_user" {
  email = "newuser@example.com"
}

# Step 3: Grant project access via openai_project_user
resource "openai_project_user" "access_to_project" {
  project_id = openai_project.example.id
  user_id    = data.openai_organization_user.new_user.id
  role       = "member"
}
```

## Argument Reference

* `email` - (Required) The email address of the user to invite.
* `role` - (Required) The organization-level role to assign to the user. Valid values:
  * `owner` - Full administrative access to the organization.
  * `reader` - Standard member access to the organization.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used. **Note:** The API key must have organization management permissions.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the invitation.
* `status` - The current status of the invitation. Possible values:
  * `pending` - The invitation has been sent but not yet accepted.
  * `accepted` - The invitation has been accepted by the user.
  * `declined` - The invitation has been declined by the user.
  * `expired` - The invitation has expired.
* `expires_at` - The timestamp (in Unix time) when the invitation will expire.
* `created_at` - The timestamp (in Unix time) when the invitation was created.

## Permission Requirements

To create invitations, your API key must have organization management permissions. Typically, this requires an owner-level API key or an API key with the appropriate administrative scopes.

## Import

Invitations can be imported using the invite ID, e.g.,

```bash
terraform import openai_invite.example inv_abc123
``` 