# List all users in a specific project
data "openai_project_users" "production_team" {
  project_id = "proj-abc123"
}

# List users with specific role in project
data "openai_project_users" "project_owners" {
  project_id = "proj-abc123"
  role       = "owner"
}

# Output total project user count
output "total_project_users" {
  value = length(data.openai_project_users.production_team.users)
}
