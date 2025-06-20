# OpenAI User Invitation and Project Assignment Workflow
# --------------------------------------------------------
# This example demonstrates a 3-step workflow:
# 1. Invite a user to the organization
# 2. After they accept, add them to a project 
# 3. Assign them appropriate roles

terraform {
  required_providers {
    openai = {
      source = "fjcorp/openai"
    }
  }
}

provider "openai" {
  # API key should be provided via environment variable OPENAI_API_KEY
  # Admin key should be provided via environment variable OPENAI_ADMIN_KEY
}

/*
# Create projects for demonstration
resource "openai_project" "project1" {
  name        = "Project 1"
  description = "First test project"
}

resource "openai_project" "project2" {
  name        = "Project 2"
  description = "Second test project"
}
*/

# Use existing projects instead
variable "project1_id" {
  default = "proj_JGhw44csZsbtjw2yxuyPjMZN"
}

variable "project2_id" {
  default = "proj_HFklSiW6icFV61P5UTo1WR77"
}

# Add variable for admin key
variable "admin_key" {
  description = "OpenAI Admin API Key with api.management.write scope"
  type        = string
  default     = ""
  sensitive   = true
}

# Example 1: Invite a new user with project assignments
# The invitation will:
# 1. Create an invitation for the user (if they don't exist)
# 2. Automatically assign them to the specified projects
resource "openai_invite" "new_user" {
  email = "pablo+terraformtest2@mkdev.me"
  role  = "reader" # Organization-level role

  # Use admin key explicitly from environment variable
  api_key = var.admin_key

  # Projects will be assigned automatically after invitation
  projects {
    id   = var.project1_id
    role = "member" # Project-level role
  }

  projects {
    id   = var.project2_id
    role = "member" # Project-level role
  }
}

# Example 2: Handle existing user (will skip invitation and assign to projects)
resource "openai_invite" "existing_user" {
  email = "pablo+existingtest@mkdev.me"
  role  = "reader" # Organization-level role

  projects {
    id   = var.project1_id
    role = "member"
  }
}

# Example 3: Alternative approach using separate resources
# This gives more control over the process

# Step 1: Create invitation (without projects)
resource "openai_invite" "manual_user" {
  email = "pablo+manualtest@mkdev.me"
  role  = "reader"

  # No projects specified here
}

/* Commenting out to avoid errors during initial test
# Step 2: Use data source to get user info after they're in the organization
data "openai_organization_user" "manual_user" {
  email = openai_invite.manual_user.email
  
  # This ensures the data source waits for the invitation
  depends_on = [openai_invite.manual_user]
}

# Step 3: Manually assign user to projects
resource "openai_project_user" "manual_assignment1" {
  project_id = var.project1_id
  user_id    = data.openai_organization_user.manual_user.id
  role       = "member"
  
  depends_on = [data.openai_organization_user.manual_user]
}

resource "openai_project_user" "manual_assignment2" {
  project_id = var.project2_id
  user_id    = data.openai_organization_user.manual_user.id
  role       = "owner"
  
  depends_on = [data.openai_organization_user.manual_user]
}
*/

# Outputs to show the results
output "new_user_invite" {
  value = {
    invite_id = openai_invite.new_user.invite_id
    status    = openai_invite.new_user.status
    user_id   = openai_invite.new_user.user_id
  }
  description = "New user invitation details"
}

output "existing_user_invite" {
  value = {
    invite_id = openai_invite.existing_user.invite_id
    status    = openai_invite.existing_user.status
    user_id   = openai_invite.existing_user.user_id
  }
  description = "Existing user invitation details"
}

/*
output "manual_user_assignments" {
  value = {
    user_id = data.openai_organization_user.manual_user.id
    project1_assignment = openai_project_user.manual_assignment1.id
    project2_assignment = openai_project_user.manual_assignment2.id
  }
  description = "Manual user assignment details"
}
*/

/* Commenting out the old workflow example
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
variable "openai_admin_key" {
  description = "OpenAI Admin API Key"
  type        = string
  sensitive   = true
}

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
*/

