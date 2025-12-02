# Create a new OpenAI project
resource "openai_project" "development" {
  title = "Development Project"
}

# Create a production project
resource "openai_project" "production" {
  title = "Production API Services"
}

# Output the project ID
output "dev_project_id" {
  value       = openai_project.development.id
  description = "The ID of the development project"
}