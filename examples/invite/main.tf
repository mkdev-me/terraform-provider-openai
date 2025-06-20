# OpenAI User Invitation and Project Assignment Workflow
# --------------------------------------------------------
# This example demonstrates a 3-step workflow:
# 1. Invite a user to the organization
# 2. After they accept, add them to a project 
# 3. Assign them appropriate roles

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "~> 1.0.0"
    }
  }
}

provider "openai" {
  # API keys are automatically loaded from environment variables:
  # - OPENAI_API_KEY for project operations
  # - OPENAI_ADMIN_KEY for admin operations
}

# Create a project
resource "openai_project" "example" {
  name = "Example Project for invite"
}

# ========= STEP 1: INVITE USER =========
# Variable for the email to invite
variable "invite_email" {
  description = "Email address to invite to the organization"
  type        = string
  default     = "pablo+newtest10@mkdev.me" # Change this to the actual email
}

# Organization-level invitation without project assignment
resource "openai_invite" "user_invite" {
  # Set count to 0 to disable inviting (after user has already been invited)
  count = var.skip_invitation ? 0 : 1

  email = var.invite_email
  role  = "reader" # Organization-level role (reader or owner)

  # Add user to the example project
  projects {
    id   = openai_project.example.id
    role = "member"
  }

  # This prevents Terraform from trying to delete ACCEPTED invitations, which would fail
  # Pending invitations can be deleted, but once accepted, they cannot be deleted via the API
  lifecycle {
    # The OpenAI API will return an error if trying to delete an accepted invitation
    prevent_destroy = false

    # Ignore changes to these attributes as they can't be updated after creation
    ignore_changes = [
      email,
      role
    ]
  }
}

# ========= STEP 2: ADD USER TO PROJECT =========
# After user accepts invitation, get their user_id
# This requires running 'terraform apply' again after acceptance

# Variable to hold the user's ID after they've accepted the invitation
variable "user_id" {
  description = "User ID of the invited user (after they've accepted)"
  type        = string
  default     = "" # You'll provide this after the user accepts the invite
}

# Flag to skip sending invitation (use after first run)
variable "skip_invitation" {
  description = "Set to true after the invitation has been sent and accepted"
  type        = bool
  default     = false
}

# Add user to project after they've accepted the invitation
resource "openai_project_user" "project_assignment" {
  # Only run if a user ID is provided
  count = var.user_id != "" ? 1 : 0

  project_id = openai_project.example.id
  user_id    = var.user_id
  role       = "member" # Project-level role (member or owner)
}

# ========= STEP 3: RETRIEVE USER INFORMATION =========
# Use organization_users data source to verify user information
data "openai_organization_users" "org_users" {
  # Only used to confirm successful invitation & display user information
}

# Variables

# Outputs
output "project_id" {
  value = openai_project.example.id
}

output "invitation_info" {
  value = var.skip_invitation ? {
    id         = "skipped"
    email      = var.invite_email
    status     = "skipped"
    expires_at = 0
    } : (
    length(openai_invite.user_invite) > 0 ? {
      id         = openai_invite.user_invite[0].id
      email      = openai_invite.user_invite[0].email
      status     = openai_invite.user_invite[0].status
      expires_at = openai_invite.user_invite[0].expires_at
      } : {
      id         = "none"
      email      = var.invite_email
      status     = "not_created"
      expires_at = 0
    }
  )
}

output "project_user_info" {
  value = var.user_id != "" ? {
    project_id = openai_project.example.id
    user_id    = var.user_id
    role       = length(openai_project_user.project_assignment) > 0 ? openai_project_user.project_assignment[0].role : "Not assigned"
    } : {
    project_id = ""
    user_id    = ""
    role       = "User not yet assigned to project (provide user_id after invitation is accepted)"
  }
}

output "org_users" {
  description = "All users in the organization (use to find the user ID after acceptance)"
  value = {
    for user in data.openai_organization_users.org_users.users :
    user.email => user.id
  }
}

# Add this to the existing locals block or create a new one if needed
locals {
  # Create a map of email to user ID
  org_email_to_id = {
    for user in data.openai_organization_users.org_users.users : user.email => user.id
  }

  # Check if the invited email exists in the organization
  invite_email_exists = contains(keys(local.org_email_to_id), var.invite_email)

  # Get the user ID for the invited email if it exists
  invited_user_id = local.invite_email_exists ? local.org_email_to_id[var.invite_email] : ""
}

output "invited_user_id" {
  description = "The user ID of the invited user (if they've accepted and are in the organization)"
  value       = local.invited_user_id != "" ? local.invited_user_id : "User has not yet accepted the invitation or is not in the organization"
}

output "workflow_instructions" {
  description = "Step-by-step workflow instructions"
  value       = <<EOT
WORKFLOW INSTRUCTIONS:

STEP 1: Send invitation
  terraform apply -var="invite_email=real.user@example.com"

STEP 2: Wait for user to accept invitation

STEP 3: Check if user has accepted and get their ID
  terraform apply -var="invite_email=real.user@example.com"
  Look at the "invited_user_id" output - if it shows a user ID (format: user-xxx...), proceed to step 4
  If it says "User has not yet accepted...", wait longer for acceptance

STEP 4: Assign user to project
  terraform apply -var="invite_email=real.user@example.com" -var="user_id=USER_ID_FROM_STEP_3" -var="skip_invitation=true"

IMPORTANT NOTES:
- The -var="skip_invitation=true" prevents sending another invitation
- Pending invitations can be deleted via the API, but accepted invitations cannot
- If you encounter an error about not being able to delete an accepted invitation:
  terraform state rm openai_invite.user_invite[0]
  Then run the apply command again
- To test with a new email, first remove the previous invitation from the state
EOT
}

