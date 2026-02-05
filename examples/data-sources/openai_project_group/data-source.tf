# Look up a specific group in a project by group ID
data "openai_project_group" "engineering" {
  project_id = "proj-abc123"
  group_id   = "group-xyz789"
}

# Look up a group by name (useful when you don't know the group ID)
data "openai_project_group" "support_team" {
  project_id = "proj-abc123"
  group_name = "Support Team"
}

# Output the group's role and when it was added
output "engineering_group_role" {
  value = data.openai_project_group.engineering.role
}

output "engineering_group_added_at" {
  value = data.openai_project_group.engineering.created_at
}

output "support_team_group_id" {
  value = data.openai_project_group.support_team.group_id
}
