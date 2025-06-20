# OpenAI Project Data Source Example
# ------------------------------------

# Create a project that we'll retrieve data for
resource "openai_project" "example" {
  name = "Project Data Source Example"
}

# Retrieve information about the project using the data source
data "openai_project" "project_info" {
  project_id = openai_project.example.id
  admin_key  = var.openai_admin_key
  depends_on = [openai_project.example]
}

# Retrieve all projects using the projects data source
data "openai_projects" "all_projects" {
  depends_on = [openai_project.example]
  admin_key  = var.openai_admin_key
}

# Outputs
output "project_name" {
  description = "The name of the project"
  value       = data.openai_project.project_info.name
}

output "project_status" {
  description = "The status of the project"
  value       = data.openai_project.project_info.status
}

output "project_created_at" {
  description = "When the project was created"
  value       = data.openai_project.project_info.created_at
}

# Output usage limits if available
output "project_usage_limits" {
  description = "Usage limits for the project"
  value       = data.openai_project.project_info.usage_limits
}

# Outputs for all projects
output "all_projects_count" {
  description = "The number of projects in the OpenAI account"
  value       = length(data.openai_projects.all_projects.projects)
}

output "all_project_names" {
  description = "The names of all projects in the OpenAI account"
  value       = [for p in data.openai_projects.all_projects.projects : p.name]
}

output "all_project_ids" {
  description = "The IDs of all projects in the OpenAI account"
  value       = [for p in data.openai_projects.all_projects.projects : p.id]
} 