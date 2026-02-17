# Look up a specific group in a project by group ID
data "openai_project_group" "engineering" {
  project_id = "proj_abc123"
  group_id   = "group_01J1F8ABCDXYZ"
}

# Look up a group by name (useful when you don't know the group ID)
data "openai_project_group" "support_team" {
  project_id = "proj_abc123"
  group_name = "Support Team"
}

# Output when the group was added to the project
output "engineering_group_added_at" {
  value = data.openai_project_group.engineering.created_at
}

output "support_team_group_id" {
  value = data.openai_project_group.support_team.group_id
}

# To get role assignments for a group, use the openai_project_group_roles data source
