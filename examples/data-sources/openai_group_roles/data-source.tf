# List all organization roles assigned to a group
data "openai_group_roles" "engineering" {
  group_id = "group_01J1F8ABCDXYZ"
}

# Output role IDs assigned to the group
output "assigned_role_ids" {
  value = data.openai_group_roles.engineering.role_ids
}

# Output role names
output "assigned_role_names" {
  value = [for ra in data.openai_group_roles.engineering.role_assignments : ra.role_name]
}
