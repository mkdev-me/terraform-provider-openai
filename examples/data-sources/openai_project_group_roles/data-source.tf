# List all roles assigned to a specific group within a project
data "openai_project_group_roles" "engineering_roles" {
  project_id = "proj_abc123"
  group_id   = "group_01J1F8ABCDXYZ"
}

# Output all role IDs assigned to the group
output "assigned_role_ids" {
  value = data.openai_project_group_roles.engineering_roles.role_ids
}

# Output role names assigned to the group
output "assigned_role_names" {
  value = [for a in data.openai_project_group_roles.engineering_roles.role_assignments : a.role_name]
}
