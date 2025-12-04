# Fetch a specific project by ID
data "openai_project" "production" {
  project_id = "proj_1d8XmJiB5LCRIvUofz0uGqGK"
}

# Output project ID
output "project_id" {
  value = data.openai_project.production.id
}

# Example of conditional logic based on project status

# Use project data to set variables
locals {
  project_active = data.openai_project.production.status == "active"
  project_name   = data.openai_project.production.name
}
