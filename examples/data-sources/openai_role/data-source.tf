# Look up a specific role by name in the organization
data "openai_role" "owner" {
  name = "Owner"
}

# Use the role ID in other resources
output "owner_role_id" {
  value = data.openai_role.owner.role_id
}

output "owner_role_permissions" {
  value = data.openai_role.owner.permissions
}
