---
page_title: "OpenAI: openai_invite Data Source"
subcategory: ""
description: |-
  Retrieves information about a specific invitation to an OpenAI organization.
---

# openai_invite Data Source

The `openai_invite` data source retrieves details about a specific invitation to join an OpenAI organization.

## Example Usage

```hcl
resource "openai_invite" "example" {
  email = "user@example.com"
  role  = "reader"
}

data "openai_invite" "invite_details" {
  invite_id = openai_invite.example.id
  
  depends_on = [openai_invite.example]
}

output "invite_status" {
  value = data.openai_invite.invite_details.status
}
```

## Invitation Workflow Challenges

Working with OpenAI invitations presents several challenges:

1. **Project Assignment Not Supported**: The OpenAI API does not support adding users to projects through invitations. You must invite users to the organization first, then add them to projects after they accept the invitation.

2. **Deletion of Accepted Invitations**: The OpenAI API does not allow deleting invitations that have already been accepted. Attempting to delete an accepted invitation will result in an error.

3. **Invitation Status Lag**: After a user accepts an invitation, it may take some time for the API to reflect the change in status.

## Recommended Workflow

To handle these challenges, follow this workflow:

1. Send an invitation to a user (organization-level only)
2. Wait for the user to accept the invitation
3. Use the `openai_organization_users` data source to check if the user appears in the organization
4. Add the user to projects using the `openai_project_user` resource
5. If using Terraform to manage the lifecycle, handle accepted invitations appropriately:
   - Either remove them from Terraform state
   - Or use the modified client that treats "already accepted" errors as success during deletion

See the [invite example](/examples/invite/) for a complete implementation of this workflow.

## Argument Reference

* `invite_id` - (Required) The ID of the invitation to retrieve.

## Attributes Reference

* `id` - The ID of the invitation.
* `email` - The email address the invitation was sent to.
* `role` - The role assigned to the invited user.
* `status` - The status of the invitation (pending, accepted, etc.).
* `created_at` - The timestamp when the invitation was created.
* `expires_at` - The timestamp when the invitation expires.

## Handling Invitation Deletion

When an invitation is accepted, you cannot delete it through the API. If you're using Terraform to manage invitations and encounter the "already accepted" error during deletion, you have two options:

1. **Remove from state**: `terraform state rm openai_invite.example`
2. **Use the modified client**: If the provider has been updated with the patch that handles "already accepted" errors, it will gracefully handle this case.

## Permission Requirements

To use this data source, your API key must have permissions to read invitation details. This typically requires an admin API key with organization-level access.

## Import

Invite data sources cannot be imported. 