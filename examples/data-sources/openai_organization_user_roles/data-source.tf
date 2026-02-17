# List all organization role assignments for a user
data "openai_organization_user_roles" "admin" {
  user_id = var.user_id
}

output "assigned_role_ids" {
  value = data.openai_organization_user_roles.admin.role_ids
}
