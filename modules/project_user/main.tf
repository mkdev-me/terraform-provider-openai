# OpenAI Project User Module
# ================================
# This module simulates management of users within OpenAI projects.
# IMPORTANT: This is a placeholder module as the openai_project_user resource
# is not yet fully implemented in the provider.

terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "~> 1.0.0"
    }
  }
}

# Required variables
variable "project_id" {
  description = "The ID of the project where the user will be added"
  type        = string
}

variable "user_id" {
  description = "The ID of the user to add to the project (format: user-abc123). This must be the full OpenAI user ID, not the username or email."
  type        = string
  default     = null
}

variable "role" {
  description = "The role to assign to the user (owner or member)"
  type        = string
  default     = "member"
  validation {
    condition     = contains(["owner", "member"], var.role)
    error_message = "The role must be either 'owner' or 'member'."
  }
}

# Optional variables
variable "openai_admin_key" {
  description = "Admin API key to use for this resource. If not provided, the provider's default API key will be used."
  type        = string
  default     = null
  sensitive   = true
}

# New: Mode selection variable
variable "list_mode" {
  description = "When true, retrieves all users in a project instead of managing a single user"
  type        = bool
  default     = false
}

# Create the project user (when not in list mode)
resource "openai_project_user" "user" {
  count      = var.list_mode ? 0 : 1
  project_id = var.project_id
  user_id    = var.user_id
  role       = var.role
  api_key    = var.openai_admin_key
}

# Retrieve all project users (when in list mode)
data "openai_project_users" "all" {
  count      = var.list_mode ? 1 : 0
  project_id = var.project_id
  api_key    = var.openai_admin_key
}

# Outputs for single user mode
output "id" {
  description = "The unique identifier for the project user"
  value       = var.list_mode ? null : try(openai_project_user.user[0].id, null)
}

output "email" {
  description = "The email address of the user"
  value       = var.list_mode ? null : try(openai_project_user.user[0].email, null)
}

output "added_at" {
  description = "The timestamp when the user was added to the project"
  value       = var.list_mode ? null : try(openai_project_user.user[0].added_at, null)
}

output "role" {
  description = "The role assigned to the user"
  value       = var.list_mode ? null : try(openai_project_user.user[0].role, null)
}

# Outputs for list mode
output "all_users" {
  description = "List of all users in the project (only available in list mode)"
  value       = var.list_mode ? try(data.openai_project_users.all[0].users, []) : null
  sensitive   = true
}

output "user_count" {
  description = "Number of users in the project (only available in list mode)"
  value       = var.list_mode ? try(data.openai_project_users.all[0].user_count, 0) : null
}

output "project_owners" {
  description = "List of users with owner role in the project (only available in list mode)"
  value       = var.list_mode ? try([for user in data.openai_project_users.all[0].users : user if user.role == "owner"], []) : null
  sensitive   = true
}

output "project_members" {
  description = "List of users with member role in the project (only available in list mode)"
  value       = var.list_mode ? try([for user in data.openai_project_users.all[0].users : user if user.role == "member"], []) : null
  sensitive   = true
}

# New non-sensitive outputs for list mode
output "all_user_ids" {
  description = "List of all user IDs in the project (only available in list mode)"
  value       = var.list_mode ? try(data.openai_project_users.all[0].user_ids, []) : null
}

output "owner_ids" {
  description = "List of user IDs with owner role in the project (only available in list mode)"
  value       = var.list_mode ? try(data.openai_project_users.all[0].owner_ids, []) : null
}

output "member_ids" {
  description = "List of user IDs with member role in the project (only available in list mode)"
  value       = var.list_mode ? try(data.openai_project_users.all[0].member_ids, []) : null
}
