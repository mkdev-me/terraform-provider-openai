# List all project role assignments for a user
data "openai_project_user_roles" "developer" {
  project_id = var.project_id
  user_id    = var.user_id
}

output "assigned_role_ids" {
  value = data.openai_project_user_roles.developer.role_ids
}
