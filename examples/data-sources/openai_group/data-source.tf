# Look up a group by ID
data "openai_group" "by_id" {
  id = "group-abc123"
}

# Look up a group by name
data "openai_group" "by_name" {
  name = "Engineering Team"
}

# Output the group details
output "group_id" {
  value = data.openai_group.by_name.id
}

output "group_created_at" {
  value = data.openai_group.by_name.created_at
}

output "is_scim_managed" {
  value = data.openai_group.by_name.is_scim_managed
}
