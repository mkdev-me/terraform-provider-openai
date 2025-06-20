# OpenAI User Management Example
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


# IMPORTANT: This example uses a real email address
# The user must already exist in your OpenAI organization
# You must first invite this user through the OpenAI dashboard
# (OpenAI Settings > Members > Invite)
variable "user_email" {
  type        = string
  description = "Email of an existing user in your OpenAI organization"
  default     = "pabloinigo@gmail.com"
}

# Create an OpenAI project
resource "openai_project" "example" {
  name = "Example Project for User Management"
}

# STEP 1: First, retrieve the user's information
# You can get this using the organization_user data source
data "openai_organization_user" "user" {
  # Use the actual user ID retrieved from the API
  user_id = "user-yatSd6LuWvgeoqZbd89xzPlJ" # ID example
}

# Get all organization users to identify owners
data "openai_organization_users" "all_users" {
  # Use a reasonable limit for your organization
  limit = 20
}

# Create a local variable to identify if the user is an organization owner
locals {
  is_org_owner = data.openai_organization_user.user.role == "owner"

  # List of organization owner IDs for reference
  org_owner_ids = [
    for user in data.openai_organization_users.all_users.users :
    user.id if user.role == "owner"
  ]
}

# STEP 2: Add the user to the project
# The user must already exist in your OpenAI organization
resource "openai_project_user" "project_user" {
  project_id = openai_project.example.id
  # Use the actual user ID retrieved from the API
  user_id = data.openai_organization_user.user.id

  # IMPORTANT: For organization owners, always use "owner" role
  # For regular users, you can use either "owner" or "member"
  role = local.is_org_owner ? "owner" : "owner" # Change the second value to "member" for non-owners

  depends_on = [openai_project.example]
}

# STEP 3: For a second project, respect organization owner limitations
resource "openai_project_user" "additional_project" {
  project_id = openai_project.additional.id
  user_id    = data.openai_organization_user.user.id

  # IMPORTANT: For organization owners, always use "owner" role
  # For regular users, you can use either "owner" or "member"
  role = local.is_org_owner ? "owner" : "member"

  depends_on = [openai_project.additional]
}

# Additional project for demonstration
resource "openai_project" "additional" {
  name = "Additional Project Example"
}

# Outputs for verification
output "project_id" {
  value = openai_project.example.id
}

output "project_user_email" {
  value = data.openai_organization_user.user.email
}

output "project_user_role" {
  value = openai_project_user.project_user.role
}

output "organization_user_role" {
  value = data.openai_organization_user.user.role
}

output "user_name" {
  value = data.openai_organization_user.user.name
}

data "openai_project_users" "all_users" {
  project_id = openai_project.additional.id
}

output "all_users" {
  value = data.openai_project_users.all_users.users
}

# Add additional non-sensitive outputs
output "user_ids" {
  value = data.openai_project_users.all_users.user_ids
}

output "owner_ids" {
  value = data.openai_project_users.all_users.owner_ids
}

output "member_ids" {
  value = data.openai_project_users.all_users.member_ids
}

output "user_count" {
  value = data.openai_project_users.all_users.user_count
}
