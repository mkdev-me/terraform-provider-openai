# OpenAI Invitation Workflow Module

This module provides a reusable implementation of the proper OpenAI invitation workflow that addresses the known limitations of the OpenAI API.

## Purpose

This module solves several challenges with the OpenAI invitation process:

1. **Project Assignments Not Possible via Invitations**: The OpenAI API does not support project assignments through invitations. Users must be invited to the organization first, then added to projects separately.

2. **Accepted Invitations Can't Be Deleted**: The API returns an error when trying to delete an already accepted invitation.

3. **Multi-Step Workflow Needed**: A lookup step is required to get user IDs after invitation acceptance before adding them to projects.

## Module Features

- **Complete Workflow Implementation**: Handles invitation, checking for acceptance, and project assignment
- **Robust Error Handling**: Gracefully handles "already accepted" errors during deletion
- **Automatic User ID Lookup**: Uses the organization_users data source to find user IDs
- **Skip Flag for Accepted Invitations**: Prevents sending duplicate invitations

## Usage

```hcl
# Step 1: Invite the user to the organization
module "invite" {
  source = "../../modules/invite"

  email = "user@example.com"
  role  = "reader"
}

# Step 2: Get the user ID after they've accepted the invitation
data "openai_organization_user" "new_user" {
  email = "user@example.com"
}

# Step 3: Add the user to a project
resource "openai_project_user" "project_access" {
  project_id = openai_project.example.id
  user_id    = data.openai_organization_user.new_user.id
  role       = "member"
}
```

## Recommended Workflow

### Step 1: Send Invitation

```bash
terraform apply
```

This sends the invitation and sets up the initial state.

### Step 2: Wait for User to Accept and Check Status

Wait for the user to accept the invitation, then run:

```bash
terraform apply
```

Review the data source outputs to check if the user has accepted the invitation.

### Step 3: Add User to Project

Once the user has accepted and their user ID is available, the `openai_project_user` resource will add them to the project.

## Handling State Issues

If you encounter errors related to deleting accepted invitations, you may need to remove the invitation from the Terraform state:

```bash
terraform state rm module.invite.openai_invite.invite[0]
```

## Provider Code Modification

This module works best with the modified provider that gracefully handles "already accepted" errors during invitation deletion:

```go
// In internal/client/client.go
func (c *OpenAIClient) DeleteInvite(inviteID string, customAPIKey string) error {
    // ... existing code ...
    
    if err != nil {
        // Check if the error is due to the invitation already being accepted
        if strings.Contains(err.Error(), "already accepted") {
            // If the invite is already accepted, we consider it deleted for Terraform purposes
            return nil
        }
        return fmt.Errorf("error deleting invite: %s", err)
    }
    
    return nil
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| email | Email address to invite | `string` | n/a | yes |
| role | Organization role (reader or owner) | `string` | n/a | yes |
| openai_admin_key | Admin API key to use for invite operations | `string` | `null` | no |
| list_all_invites | Whether to include data on all pending invitations | `bool` | `false` | no |
| create_invite | Whether to create a new invitation | `bool` | `true` | no |

## Outputs

| Name | Description |
|------|-------------|
| id | The unique identifier for the invitation |
| invite_id | The ID of the invitation |
| email | The email address of the invited user |
| role | The role assigned to the invited user |
| status | The status of the invitation |
| created_at | The timestamp when the invitation was created |
| expires_at | The timestamp when the invitation expires |
| all_invites | List of all pending invitations in the organization |
| invite_count | Number of pending invitations in the organization |

## License

This module is licensed under the terms of the license file included in the parent repository.