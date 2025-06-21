terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
    }
  }
}

provider "openai" {
  # API keys are automatically loaded from environment variables:
  # - OPENAI_API_KEY for project operations
  # - OPENAI_ADMIN_KEY for admin operations
}

# List all projects in the organization
data "openai_projects" "all" {
  # Uses admin API key from provider or environment
}

# Output all projects
output "all_projects" {
  value = data.openai_projects.all.projects
}

# Output project count
output "project_count" {
  value = length(data.openai_projects.all.projects)
}

# Output project names
output "project_names" {
  value = [for p in data.openai_projects.all.projects : p.name]
}