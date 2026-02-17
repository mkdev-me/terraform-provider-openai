# List all roles in a project
data "openai_project_roles" "all" {
  project_id = var.project_id
}

output "role_count" {
  value = data.openai_project_roles.all.role_count
}

output "role_ids" {
  value = data.openai_project_roles.all.role_ids
}
