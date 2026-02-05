# List all groups in the organization
data "openai_groups" "all" {}

# Output total group count
output "total_groups" {
  value = data.openai_groups.all.group_count
}

# Output all group IDs
output "all_group_ids" {
  value = data.openai_groups.all.group_ids
}

# Output detailed group information
output "group_details" {
  value = [
    for group in data.openai_groups.all.groups : {
      id         = group.id
      name       = group.name
      is_scim    = group.is_scim_managed
      created_at = group.created_at
    }
  ]
}
