# List all projects in the organization
data "openai_projects" "all" {
  # Optional: Set include_archived to true to include archived projects
  # include_archived = true
}

# Output total project count
output "project_count" {
  value = length(data.openai_projects.all.projects)
}

