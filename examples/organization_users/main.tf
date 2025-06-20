# OpenAI Organization Users Example
# ================================
# This example demonstrates how to use the openai_organization_users 
# and openai_organization_user data sources
# to retrieve information about users in an organization.
#
# NOTE: These are CUSTOM data sources and not part of the official provider.
# You need to build the provider locally with these modifications.
#
# IMPORTANT: API KEY PERMISSIONS
# ==============================
# This example requires an OpenAI API key with administrative permissions.
# Specifically, your API key MUST have the "api.management.read" scope.
# Regular API keys created through the OpenAI dashboard typically don't 
# have these permissions. You need an organization owner API key or a 
# key specifically scoped for organization management.

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

# Variable for enabling/disabling organization user functionality
# Set this to false if you don't have the required permissions
variable "enable_organization_users" {
  description = "Whether to enable organization user functionality"
  type        = bool
  default     = true
}

# Retrieve all users in the organization
# This requires an API key with api.management.read scope
data "openai_organization_users" "all" {
  count = var.enable_organization_users ? 1 : 0
  limit = 50
}

# Retrieve a specific user by ID
# This requires an API key with api.management.read scope
data "openai_organization_user" "specific" {
  count   = var.enable_organization_users ? 1 : 0
  user_id = "user-udjrDA1SqpU8CnkH28BGq5JY" # Replace with a real user ID
}

# Create a project to demonstrate using the obtained user IDs
resource "openai_project" "example" {
  name = "Organization Users Demo"
}

# Only add users to project if organization user functionality is enabled
resource "openai_project_user" "specific_user" {
  count      = var.enable_organization_users ? 1 : 0
  project_id = openai_project.example.id
  user_id    = var.enable_organization_users ? data.openai_organization_user.specific[0].id : "user-abc123xyz"
  role       = "owner"
}

# Only use organization data if enabled, otherwise skip
locals {
  owners = var.enable_organization_users ? [
    for user in data.openai_organization_users.all[0].users :
    user if user.role == "owner" && user.id != data.openai_organization_user.specific[0].id
  ] : []
}

resource "openai_project_user" "owners" {
  for_each = {
    for i, user in local.owners :
    user.id => user
  }

  project_id = openai_project.example.id
  user_id    = each.key
  role       = "owner"
}

# Output information about the specific user if enabled
output "specific_user" {
  value = var.enable_organization_users ? {
    id    = data.openai_organization_user.specific[0].id
    email = data.openai_organization_user.specific[0].email
    role  = data.openai_organization_user.specific[0].role
    } : {
    id    = "disabled"
    email = "Organization user functionality disabled or insufficient permissions"
    role  = "disabled"
  }
}

# Output information about all organization users if enabled
output "organization_users" {
  value = var.enable_organization_users ? [
    for user in data.openai_organization_users.all[0].users : {
      id    = user.id
      email = user.email
      role  = user.role
    }
  ] : []
  description = "List of organization users. Will be empty if functionality is disabled."
}

# Output organization user statistics if enabled
output "organization_user_count" {
  value       = var.enable_organization_users ? length(data.openai_organization_users.all[0].users) : 0
  description = "Count of organization users. Will be 0 if functionality is disabled."
}

output "owner_count" {
  value       = var.enable_organization_users ? length([for user in data.openai_organization_users.all[0].users : user if user.role == "owner"]) : 0
  description = "Count of organization owners. Will be 0 if functionality is disabled."
}

output "member_count" {
  value       = var.enable_organization_users ? length([for user in data.openai_organization_users.all[0].users : user if user.role == "member"]) : 0
  description = "Count of organization members. Will be 0 if functionality is disabled."
}

# Output a message indicating whether organization user functionality is enabled
output "organization_users_status" {
  value       = var.enable_organization_users ? "Enabled" : "Disabled or insufficient permissions"
  description = "Status of the organization users functionality"
}

