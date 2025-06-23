# First, create a project
resource "openai_project" "development" {
  name        = "Development Environment"
  description = "Project for development and testing"
}

resource "openai_project" "production" {
  name        = "Production API"
  description = "Production API project with critical services"
}

# IMPORTANT: Replace these user_id values with actual user IDs from your organization
# You can find user IDs using the organization users data source or OpenAI dashboard

# Add users to the development project
# Commented out until valid user IDs are provided
# 
# resource "openai_project_user" "dev_lead" {
#   project_id = openai_project.development.id
#   user_id    = "user-123abc" # Replace with actual user ID
#   role       = "owner"
# }
# 
# resource "openai_project_user" "dev_member1" {
#   project_id = openai_project.development.id
#   user_id    = "user-456def" # Replace with actual user ID
#   role       = "member"
# }
# 
# resource "openai_project_user" "dev_member2" {
#   project_id = openai_project.development.id
#   user_id    = "user-789ghi" # Replace with actual user ID
#   role       = "member"
# }
# 
# # Add users to the production project with more restricted access
# resource "openai_project_user" "prod_owner" {
#   project_id = openai_project.production.id
#   user_id    = "user-123abc" # Same user can be in multiple projects
#   role       = "owner"
# }
# 
# resource "openai_project_user" "prod_member" {
#   project_id = openai_project.production.id
#   user_id    = "user-999xyz" # Replace with actual user ID
#   role       = "member"
# }

# Output user details (commented out since user resources are commented)
# 
# output "dev_lead_email" {
#   value       = openai_project_user.dev_lead.email
#   description = "Email of the development lead"
# }
# 
# output "dev_lead_added_at" {
#   value       = openai_project_user.dev_lead.added_at
#   description = "When the development lead was added"
# }
# 
# output "project_user_summary" {
#   value = {
#     development = {
#       owner   = openai_project_user.dev_lead.user_id
#       members = [
#         openai_project_user.dev_member1.user_id,
#         openai_project_user.dev_member2.user_id
#       ]
#     }
#     production = {
#       owner   = openai_project_user.prod_owner.user_id
#       members = [openai_project_user.prod_member.user_id]
#     }
#   }
#   description = "Summary of users in each project"
# }

# Output development project ID
output "development_project_id" {
  value = openai_project.development.id
}