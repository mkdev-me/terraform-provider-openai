# OpenAI Invite Module
# =================
# This module manages invitations to OpenAI organizations

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "~> 1.0.0"
    }
  }
}

# Required variables
variable "email" {
  description = "The email address of the user to invite"
  type        = string
}

variable "role" {
  description = "The role to assign to the invited user (owner or reader)"
  type        = string
  validation {
    condition     = contains(["owner", "reader"], var.role)
    error_message = "The role must be either 'owner' or 'reader'."
  }
}

# Optional variables
variable "openai_admin_key" {
  description = "Admin API key to use for invite operations. If not provided, the provider's default API key will be used."
  type        = string
  default     = null
  sensitive   = true
}

variable "list_all_invites" {
  description = "Whether to include data on all pending invitations in the organization"
  type        = bool
  default     = false
}

variable "create_invite" {
  description = "Whether to create a new invitation or only retrieve existing ones"
  type        = bool
  default     = true
}

# Local values
locals {
  should_create_invite = var.create_invite
  should_list_invites  = var.list_all_invites
}

# Create the invitation
resource "openai_invite" "invite" {
  count   = local.should_create_invite ? 1 : 0
  email   = var.email
  role    = var.role
  api_key = var.openai_admin_key
}

# List all invitations
data "openai_invites" "all" {
  count   = local.should_list_invites ? 1 : 0
  api_key = var.openai_admin_key
}

# Outputs for created invite
output "id" {
  description = "The unique identifier for the invitation"
  value       = local.should_create_invite ? openai_invite.invite[0].id : null
}

output "invite_id" {
  description = "The ID of the invitation"
  value       = local.should_create_invite ? openai_invite.invite[0].invite_id : null
}

output "email" {
  description = "The email address of the invited user"
  value       = local.should_create_invite ? openai_invite.invite[0].email : null
}

output "role" {
  description = "The role assigned to the invited user"
  value       = local.should_create_invite ? openai_invite.invite[0].role : null
}

output "status" {
  description = "The status of the invitation"
  value       = local.should_create_invite ? openai_invite.invite[0].status : null
}

output "created_at" {
  description = "The timestamp when the invitation was created"
  value       = local.should_create_invite ? openai_invite.invite[0].created_at : null
}

output "expires_at" {
  description = "The timestamp when the invitation expires"
  value       = local.should_create_invite ? openai_invite.invite[0].expires_at : null
}

# Outputs for all invites
output "all_invites" {
  description = "List of all pending invitations in the organization (only when list_all_invites = true)"
  value       = local.should_list_invites ? data.openai_invites.all[0].invites : []
}

output "invite_count" {
  description = "Number of pending invitations in the organization (only when list_all_invites = true)"
  value       = local.should_list_invites ? length(data.openai_invites.all[0].invites) : 0
} 