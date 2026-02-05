# List all groups in a specific project
data "openai_project_groups" "production_groups" {
  project_id = "proj-abc123"
}

# Output total project group count
output "total_project_groups" {
  value = data.openai_project_groups.production_groups.group_count
}

# Output all group IDs
output "all_group_ids" {
  value = data.openai_project_groups.production_groups.group_ids
}

# Output groups with owner role
output "owner_group_ids" {
  value = data.openai_project_groups.production_groups.owner_ids
}

# Output groups with member role
output "member_group_ids" {
  value = data.openai_project_groups.production_groups.member_ids
}

# Output detailed group information
output "group_details" {
  value = [
    for group in data.openai_project_groups.production_groups.groups : {
      id         = group.group_id
      name       = group.group_name
      role       = group.role
      created_at = group.created_at
    }
  ]
}
