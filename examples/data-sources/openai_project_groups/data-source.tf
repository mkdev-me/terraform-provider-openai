# List all groups in a specific project
data "openai_project_groups" "production_groups" {
  project_id = "proj_abc123"
}

# Output total project group count
output "total_project_groups" {
  value = data.openai_project_groups.production_groups.group_count
}

# Output all group IDs
output "all_group_ids" {
  value = data.openai_project_groups.production_groups.group_ids
}

# Output detailed group information
output "group_details" {
  value = [
    for group in data.openai_project_groups.production_groups.groups : {
      id         = group.group_id
      name       = group.group_name
      created_at = group.created_at
    }
  ]
}

# To get role assignments for each group, use the openai_project_group_roles data source
