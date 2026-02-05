# First, create a project
resource "openai_project" "development" {
  name = var.dev_project_name
}

resource "openai_project" "production" {
  name = var.prod_project_name
}

# IMPORTANT: Replace these group_id values with actual group IDs from your organization
# Groups are synced from your identity provider via SCIM
# You can find group IDs using the openai_project_groups data source or OpenAI dashboard

# Add groups to the development project
# Commented out until valid group IDs are provided
#
# resource "openai_project_group" "dev_team" {
#   project_id = openai_project.development.id
#   group_id   = "group-123abc"  # Replace with actual group ID from your IdP
#   role       = "owner"
# }
#
# resource "openai_project_group" "engineering" {
#   project_id = openai_project.development.id
#   group_id   = "group-456def"  # Replace with actual group ID from your IdP
#   role       = "member"
# }
#
# # Add groups to the production project with more restricted access
# resource "openai_project_group" "prod_admins" {
#   project_id = openai_project.production.id
#   group_id   = "group-123abc"  # Same group can be in multiple projects
#   role       = "owner"
# }
#
# resource "openai_project_group" "prod_viewers" {
#   project_id = openai_project.production.id
#   group_id   = "group-789ghi"  # Replace with actual group ID
#   role       = "member"
# }

# Output group details (commented out since group resources are commented)
#
# output "dev_team_group_name" {
#   value       = openai_project_group.dev_team.group_name
#   description = "Name of the development team group"
# }
#
# output "dev_team_added_at" {
#   value       = openai_project_group.dev_team.created_at
#   description = "When the development team group was added"
# }
#
# output "project_group_summary" {
#   value = {
#     development = {
#       owner_groups  = [openai_project_group.dev_team.group_id]
#       member_groups = [openai_project_group.engineering.group_id]
#     }
#     production = {
#       owner_groups  = [openai_project_group.prod_admins.group_id]
#       member_groups = [openai_project_group.prod_viewers.group_id]
#     }
#   }
#   description = "Summary of groups in each project"
# }

# Output development project ID
output "development_project_id" {
  value = openai_project.development.id
}
