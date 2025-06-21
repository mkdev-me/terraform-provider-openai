# OpenAI Projects Module

terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

# Create an OpenAI project
resource "openai_project" "project" {
  provider   = openai
  name       = var.name
  is_default = var.is_default
}

# Outputs from the actual project resource
output "id" {
  description = "The ID of the OpenAI project"
  value       = openai_project.project.id
}

output "name" {
  description = "The name of the OpenAI project"
  value       = openai_project.project.name
}

output "created_at" {
  description = "When the OpenAI project was created"
  value       = openai_project.project.created_at
}

output "is_default" {
  description = "Whether the OpenAI project is the default project"
  value       = openai_project.project.is_default
}
