# Example demonstrating the OpenAI invite data source
# -------------------------------------------------

# Variable to control whether to create the example invitation
variable "create_example_invite" {
  description = "Whether to create the example invitation used in data sources"
  type        = bool
  default     = false
}

# 1. Create an invitation that we'll reference
resource "openai_invite" "example_invite_data" {
  count = var.create_example_invite ? 1 : 0

  email = "someone-new-example-a@example.com"
  role  = "reader" # Can be "owner" or "reader"

  projects {
    id   = openai_project.example.id
    role = "member" # Can be "owner" or "member"
  }
}

# 2. Retrieve the invitation details using the data source
data "openai_invite" "invitation_details" {
  count     = var.create_example_invite ? 1 : 0
  invite_id = openai_invite.example_invite_data[0].id

  # Data source will only be evaluated after the invitation is created
  depends_on = [openai_invite.example_invite_data]
}

# 3. List all pending invitations in the organization
# This is disabled by default as it can cause timeouts in organizations with many invites
# Set var.list_invites to true if you want to enable this
variable "list_invites" {
  description = "Whether to list all invitations (caution: can cause timeouts in organizations with many invites)"
  type        = bool
  default     = false
}

data "openai_invites" "all_invitations" {
  count = var.list_invites ? 1 : 0 # Only run this if list_invites is true

  # Ensure the data source is only evaluated after our example invite is created
  depends_on = [openai_invite.example_invite_data]
}

# Outputs for single invite
output "invite_id" {
  description = "ID of the invitation"
  value       = var.create_example_invite ? data.openai_invite.invitation_details[0].id : "example_invite_disabled"
}

output "invite_email" {
  description = "Email address of the invitation"
  value       = var.create_example_invite ? data.openai_invite.invitation_details[0].email : "example_invite_disabled"
}

output "invite_role" {
  description = "Role assigned to the invited user"
  value       = var.create_example_invite ? data.openai_invite.invitation_details[0].role : "example_invite_disabled"
}

output "invite_status" {
  description = "Status of the invitation"
  value       = var.create_example_invite ? data.openai_invite.invitation_details[0].status : "example_invite_disabled"
}

output "invite_created_at" {
  description = "When the invitation was created"
  value       = var.create_example_invite ? data.openai_invite.invitation_details[0].created_at : "example_invite_disabled"
}

output "invite_expires_at" {
  description = "When the invitation expires"
  value       = var.create_example_invite ? data.openai_invite.invitation_details[0].expires_at : "example_invite_disabled"
}

# Outputs for all invites - only produced when list_invites = true
output "all_invitations_count" {
  description = "Number of pending invitations in the organization"
  value       = var.list_invites ? length(data.openai_invites.all_invitations[0].invites) : 0
}

output "all_invitation_emails" {
  description = "Email addresses of all pending invitations"
  value       = var.list_invites ? [for invite in data.openai_invites.all_invitations[0].invites : invite.email] : []
}

output "invitation_by_id" {
  description = "Map of invitation IDs to their details"
  value = var.list_invites ? {
    for invite in data.openai_invites.all_invitations[0].invites :
    invite.id => {
      email      = invite.email
      role       = invite.role
      status     = invite.status
      created_at = invite.created_at
      expires_at = invite.expires_at
    }
  } : {}
}
