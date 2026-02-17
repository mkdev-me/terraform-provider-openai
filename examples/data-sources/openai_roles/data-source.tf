# List all roles configured for the organization
data "openai_roles" "all" {
}

# Output all role IDs
output "all_role_ids" {
  value = data.openai_roles.all.role_ids
}

# Output the count of roles
output "role_count" {
  value = data.openai_roles.all.role_count
}

# Output role names
output "role_names" {
  value = [for role in data.openai_roles.all.roles : role.name]
}
