# First, create a project
resource "openai_project" "development" {
  name = var.dev_project_name
}

resource "openai_project" "production" {
  name = var.prod_project_name
}

# IMPORTANT: Replace these group_id and role_id values with actual IDs from your organization
# Groups are synced from your identity provider via SCIM
# You can find group IDs using the openai_groups data source or OpenAI dashboard
# You can find role IDs using the openai_project_roles data source

# Add groups to the development project
# Commented out until valid group IDs and role IDs are provided
#
# resource "openai_project_group" "dev_team" {
#   project_id = openai_project.development.id
#   group_id   = "group_01J1F8ABCDXYZ"  # Replace with actual group ID from your IdP
#   role_id    = "role_01J1F8PROJ"      # Replace with actual role ID
# }
#
# resource "openai_project_group" "engineering" {
#   project_id = openai_project.development.id
#   group_id   = "group_01J1F8DEFGHI"  # Replace with actual group ID from your IdP
#   role_id    = "role_01J1F8PROJ"     # Replace with actual role ID
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

# Output development project ID
output "development_project_id" {
  value = openai_project.development.id
}
