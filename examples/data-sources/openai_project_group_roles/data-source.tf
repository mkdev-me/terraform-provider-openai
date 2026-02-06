# List all roles assigned to a specific group within a project
data "openai_project_group_roles" "engineering_roles" {
  project_id = "proj_abc123"
  group_id   = "group_01J1F8ABCDXYZ"
}

# Output all role IDs assigned to the group
output "assigned_role_ids" {
  value = data.openai_project_group_roles.engineering_roles.role_ids
}

# Output detailed role assignment information
output "role_assignments" {
  value = [
    for assignment in data.openai_project_group_roles.engineering_roles.role_assignments : {
      assignment_id = assignment.assignment_id
      role_id       = assignment.role_id
      role_name     = assignment.role_name
      permissions   = assignment.permissions
      group_name    = assignment.group_name
    }
  ]
}
