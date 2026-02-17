# List all project role assignments for a group
data "openai_project_group_roles" "engineering" {
  project_id = var.project_id
  group_id   = var.group_id
}

output "assigned_role_ids" {
  value = data.openai_project_group_roles.engineering.role_ids
}
