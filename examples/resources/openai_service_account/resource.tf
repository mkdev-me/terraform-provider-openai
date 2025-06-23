# Create a project first
resource "openai_project" "microservices" {
  name        = "Microservices Platform"
  description = "Project for microservices API access"
}

# Create a service account for API automation
resource "openai_project_service_account" "api_automation" {
  project_id = openai_project.microservices.id
  name       = "api-automation-service"
}

# Create a service account for CI/CD pipeline
resource "openai_project_service_account" "cicd_pipeline" {
  project_id = openai_project.microservices.id
  name       = "github-actions-service"
}

# Create a service account for monitoring and analytics
resource "openai_project_service_account" "monitoring" {
  project_id = openai_project.microservices.id
  name       = "monitoring-service-account"
}

# Create a service account for testing
resource "openai_project_service_account" "testing" {
  project_id = openai_project.microservices.id
  name       = "integration-test-service"
}

# Create service accounts in different projects
resource "openai_project" "development" {
  name        = "Development Environment"
  description = "Project for development services"
}

resource "openai_project_service_account" "dev_service" {
  project_id = openai_project.development.id
  name       = "dev-api-service"
}

# Output API automation service account ID
output "api_automation_id" {
  value       = openai_project_service_account.api_automation.service_account_id
  description = "ID of the API automation service account"
}