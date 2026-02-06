# List all roles configured for a project
data "openai_project_roles" "all" {
  project_id = "proj_abc123"
}

# Output all role IDs
output "all_role_ids" {
  value = data.openai_project_roles.all.role_ids
}

# Output the count of roles
output "role_count" {
  value = data.openai_project_roles.all.role_count
}

# Output role names
output "role_names" {
  value = [for role in data.openai_project_roles.all.roles : role.name]
}
