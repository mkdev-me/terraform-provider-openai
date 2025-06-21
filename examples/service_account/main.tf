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

# Define the Admin API Key variable

# Add a variable to control whether to try creating a service account
variable "try_create_service_account" {
  description = "Whether to try creating a service account (requires api.organization.projects.service_accounts.write permission)"
  type        = bool
  default     = false
}

# Use existing project
locals {
  project_id = "proj_JGhw44csZsbtjw2yxuyPjMZN"
}

# Create a service account in the project - but only if explicitly enabled
# Note: This requires an admin API key with api.organization.projects.service_accounts.write scope
resource "openai_project_service_account" "demo" {
  count      = var.try_create_service_account ? 1 : 0
  project_id = local.project_id
  name       = "Terraform Demo Account"
}

output "project_id" {
  value = local.project_id
}

output "service_account_id" {
  value = var.try_create_service_account ? try(openai_project_service_account.demo[0].service_account_id, "Permission error - Unable to read service account") : "Service account creation disabled"
}

output "service_account_role" {
  value = var.try_create_service_account ? try(openai_project_service_account.demo[0].role, "Permission error - Unable to read service account role") : "Service account creation disabled"
}

output "api_key_id" {
  value = var.try_create_service_account ? try(openai_project_service_account.demo[0].api_key_id, "Not available") : "Service account creation disabled"
}

output "api_key" {
  value     = var.try_create_service_account ? try(openai_project_service_account.demo[0].api_key_value, "Not available") : "Service account creation disabled"
  sensitive = true
}

# Note: API keys must be created manually in the OpenAI dashboard
# OpenAI does not support programmatic creation of project API keys.

