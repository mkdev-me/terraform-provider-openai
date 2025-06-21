# OpenAI Invitation Workflow Example

This example demonstrates a complete workflow for inviting users to an OpenAI organization and adding them to projects, including solutions for handling the known limitations of the OpenAI API.

## Features

- Send invitations to join your organization
- Assign organization-level roles (owner or reader) to invited users
- Assign users to specific projects with designated roles (owner or member)
- Retrieve detailed information about invitations

## Prerequisites

- An OpenAI account with administrator privileges
- An OpenAI Admin API key with the `api.management.write` scope
- Your OpenAI Organization ID
- Terraform >= 0.13.0

## Known Issues with OpenAI Invitations

When working with OpenAI invitations, you'll encounter several challenges:

1. **Project Assignments Don't Apply**: The OpenAI API does not correctly apply project assignments through invitations. You must use a separate resource to add users to projects after they've accepted the invitation.

2. **Accepted Invitations Can't Be Deleted**: The OpenAI API returns an error when attempting to delete an invitation that has already been accepted.

3. **Workflow Requires Multiple Steps**: You need to perform a lookup step to find user IDs after invitation acceptance before you can add users to projects.

## Solution Overview

This example implements a complete solution:

1. **Proper Invitation Workflow**:
   - Send invitation to organization only
   - Check if user accepted (using organization_users data source)
   - Get user ID and add to project explicitly

2. **Error Handling for Accepted Invitations**:
   - Provider code updated to gracefully handle "already accepted" deletion errors
   - Example includes instructions for removing accepted invitations from state

3. **Automatic User ID Lookup**:
   - Uses local variables to create a map of email → user ID
   - Automatically finds the invited user's ID when available

## Setup

1. Copy `terraform.tfvars.example` to `terraform.tfvars` and update with your credentials:

```hcl
openai_admin_key = "sk-your-admin-api-key"  # Must have api.management.write scope
organization_id  = "org-your-org-id"
```

2. Update the email addresses in `main.tf` to use real email addresses for your invitations:

```hcl
resource "openai_invite" "basic_invite" {
  email = "real.user@example.com"  # Update with a real email
  role  = "reader"
}
```

## Usage

### Step 1: Send Invitation

```bash
terraform apply -var="invite_email=user@example.com"
```

This creates the invitation and sets up the terraform state.

### Step 2: Wait for User to Accept

After the user accepts the invitation, run Terraform again to check:

```bash
terraform apply -var="invite_email=user@example.com"
```

Look at the `invited_user_id` output to see if the user has accepted. If it shows a user ID (format: user-xxx...), proceed to step 3.

### Step 3: Add User to Project

```bash
terraform apply -var="invite_email=user@example.com" -var="user_id=USER_ID_FROM_STEP_2" -var="skip_invitation=true"
```

The `skip_invitation=true` flag prevents sending another invitation.

## Handling Accepted Invitation Deletion Errors

If you encounter an error about not being able to delete an accepted invitation:

```bash
terraform state rm openai_invite.user_invite[0]
```

Then run the apply command again.

## Provider Code Fix

The provider's client code has been updated to handle "already accepted" errors during deletion:

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

This allows Terraform to continue without errors when an accepted invitation cannot be deleted.

## File Structure

- `main.tf`: Main Terraform configuration implementing the workflow
- `data_sources.tf`: Examples of using invitation data sources
- `variables.tf`: Variable definitions
- `outputs.tf`: Outputs for viewing invitation status

## Resource Examples

This example demonstrates different ways to use the `openai_invite` resource:

### Basic Invite
A simple invite with only organization access:
```hcl
resource "openai_invite" "basic_invite" {
  email = "user1@example.com"
  role  = "reader"  # Can be "owner" or "reader"
}
```

### Custom API Key Invite
An invite using a custom API key:
```hcl
resource "openai_invite" "custom_key_invite" {
  email   = "user2@example.com"
  role    = "reader"
}
```

## Recommended Project Assignment Pattern
After the user accepts the invitation, assign them to projects:
```hcl
# First get the user ID
data "openai_organization_user" "invited_user" {
  email = "user1@example.com"
}

# Then add them to the project
resource "openai_project_user" "project_access" {
  project_id = openai_project.example.id
  user_id    = data.openai_organization_user.invited_user.id
  role       = "member"  # Can be "owner" or "member"
}
```

## Data Source Example

The `data_source_example.tf` file demonstrates how to use the `openai_invite` data source to retrieve information about an invitation:

```hcl
# Create an invitation first
resource "openai_invite" "example_invite" {
  email = "someone@example.com"
  role  = "reader"
}

# Then retrieve it using the data source
data "openai_invite" "invitation_details" {
  invite_id = openai_invite.example_invite.id
  depends_on = [openai_invite.example_invite]
}

# Access invitation details
output "invite_status" {
  value = data.openai_invite.invitation_details.status
}
```

## Outputs

After applying, the following outputs are available:

### From the resource examples:
- `basic_invite_id`: The ID of the basic invitation
- `basic_invite_status`: The status of the basic invitation
- `project_invite_id`: The ID of the project invitation
- `multi_project_invite_id`: The ID of the multi-project invitation

### From the data source example:
- `invite_id`: ID of the invitation
- `invite_email`: Email address of the invitation
- `invite_role`: Role assigned to the invited user
- `invite_status`: Status of the invitation
- `invite_created_at`: When the invitation was created
- `invite_expires_at`: When the invitation expires

## Cleanup

To delete all created resources:

```bash
terraform destroy
```

## Important Notes

- Invitations expire after a certain period (typically 7 days)
- Once an invitation is accepted, the user becomes a member of your organization
- You cannot modify an invitation after it's been sent; you must delete and recreate it
- The invited email address must belong to a valid OpenAI account to accept the invitation 