terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}


provider "openai" {
  # API keys are automatically loaded from environment variables:
  # - OPENAI_API_KEY for project operations
  # - OPENAI_ADMIN_KEY for admin operations
}

# Simple project example
resource "openai_project" "test" {
  name = "Test Project from Simple Config"
}

# Output the basic project information
output "project" {
  value = {
    id   = openai_project.test.id
    name = openai_project.test.name
  }
}

# Example of a more complex project for production use
resource "openai_project" "production" {
  name = "Production Environment"

  # Note: The project will be created with the organization ID specified in the provider
}

# Additional examples to demonstrate organization:

# Create a development environment project
resource "openai_project" "development" {
  name = "Development Environment"

  # Note: Using count = 0 to make this example optional
  count = 0
}

# Create a staging environment project  
resource "openai_project" "staging" {
  name = "Staging Environment"

  # Note: Using count = 0 to make this example optional
  count = 0
}

# Output showing how to access various project attributes
output "production_project" {
  value = {
    id         = openai_project.production.id
    name       = openai_project.production.name
    created_at = openai_project.production.created_at
    status     = openai_project.production.status
  }
}
