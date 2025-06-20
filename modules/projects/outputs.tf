# Outputs - dynamically reference either the resource, data source, or list mode based on variables

output "project_id" {
  description = "The ID of the project (only in single project mode)"
  value       = var.list_mode ? null : (var.create_project ? one(openai_project.project[*].id) : one(data.openai_project.project[*].id))
}

output "project_name" {
  description = "The name of the project (only in single project mode)"
  value       = var.list_mode ? null : (var.create_project ? one(openai_project.project[*].name) : one(data.openai_project.project[*].name))
}

output "project_status" {
  description = "The status of the project (only in single project mode)"
  value       = var.list_mode ? null : (var.create_project ? one(openai_project.project[*].status) : one(data.openai_project.project[*].status))
}

output "project_created_at" {
  description = "When the project was created (only in single project mode)"
  value       = var.list_mode ? null : (var.create_project ? one(openai_project.project[*].created_at) : one(data.openai_project.project[*].created_at))
}

output "project_usage_limits" {
  description = "Usage limits for the project (only in single project mode)"
  value       = var.list_mode ? null : (var.create_project ? null : one(data.openai_project.project[*].usage_limits))
}

# List mode outputs
output "projects" {
  description = "List of all projects (only in list mode)"
  value       = var.list_mode ? try(one(data.openai_projects.all[*].projects), []) : null
}

output "project_count" {
  description = "Number of projects (only in list mode)"
  value       = var.list_mode ? try(length(one(data.openai_projects.all[*].projects)), 0) : null
}

output "project_names" {
  description = "Names of all projects (only in list mode)"
  value       = var.list_mode ? try([for p in one(data.openai_projects.all[*].projects) : p.name], []) : null
}

output "project_ids" {
  description = "IDs of all projects (only in list mode)"
  value       = var.list_mode ? try([for p in one(data.openai_projects.all[*].projects) : p.id], []) : null
} 