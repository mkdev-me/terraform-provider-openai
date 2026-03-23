# Create a new OpenAI project
resource "openai_project" "development" {
  name = "Development Project"
}

# Create a production project pinned to the US region
resource "openai_project" "production_us" {
  name      = "Production API Services (US)"
  geography = "US"
}

# Create a production project pinned to the EU region
resource "openai_project" "production_eu" {
  name      = "Production API Services (EU)"
  geography = "EU"
}

output "dev_project_id" {
  value       = openai_project.development.id
  description = "The ID of the development project"
}

output "us_project_id" {
  value       = openai_project.production_us.id
  description = "The ID of the US production project"
}

output "eu_project_id" {
  value       = openai_project.production_eu.id
  description = "The ID of the EU production project"
}