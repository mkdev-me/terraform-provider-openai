# Create a new OpenAI project
resource "openai_project" "development" {
  name        = "Development Project"
  description = "Project for development and testing purposes"
}

# Create a production project
resource "openai_project" "production" {
  name        = "Production API Services"
  description = "Project for production API services and deployments"
}

# Output the project ID
output "dev_project_id" {
  value       = openai_project.development.id
  description = "The ID of the development project"
}