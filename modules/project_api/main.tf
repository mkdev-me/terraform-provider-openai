# OpenAI Project API Key Module
# ==============================
# IMPORTANT: This module is for RETRIEVING information about existing OpenAI project API keys.
# Project API keys CANNOT be created programmatically via the API.
# You must create them manually in the OpenAI dashboard first.

# Define required providers for this module
# Module consumers must pass in provider configuration
terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai" # Custom OpenAI provider source
    }
  }
}

# Module Input Variables
# ------------------------------

# Required: Project ID to retrieve keys from
variable "project_id" {
  description = "The ID of the project to retrieve API keys for"
  type        = string
}

# Optional: Specific API key ID to retrieve (if not retrieving all)
variable "api_key_id" {
  description = "The ID of a specific API key to look up (required when retrieve_all is false)"
  type        = string
  default     = null
}

# Optional: Flag to determine whether to retrieve all keys or a single key
variable "retrieve_all" {
  description = "Whether to retrieve all API keys for the project instead of a single key"
  type        = bool
  default     = false # Default to retrieving a single key
}

# Optional: Admin API key for authentication
variable "openai_admin_key" {
  description = "Admin API key to use for this resource. If not provided, the provider's default API key will be used."
  type        = string
  default     = null
  sensitive   = true # Marks as sensitive to hide in logs and output
}

# IMPORTANT: Provider configuration was removed from this module
# The provider configuration must now be passed from the root module
# This allows the module to be used with count, for_each, and depends_on

# Data Sources
# ------------------------------

# Data source for retrieving a single API key
# Only created when retrieve_all = false
data "openai_project_api_key" "api_key" {
  count      = var.retrieve_all ? 0 : 1 # Conditional creation
  project_id = var.project_id           # Project ID
  api_key_id = var.api_key_id           # Specific API key ID
}

# Data source for retrieving all API keys for a project
# Only created when retrieve_all = true
data "openai_project_api_keys" "all_keys" {
  count      = var.retrieve_all ? 1 : 0 # Conditional creation
  project_id = var.project_id           # Project ID
}

# Module Outputs
# ------------------------------

# Outputs for single API key retrieval (retrieve_all = false)
output "api_key_id" {
  description = "The ID of the API key"
  value       = var.retrieve_all ? null : one(data.openai_project_api_key.api_key[*].api_key_id)
  # Uses one() function to safely handle the list result from count
}

output "name" {
  description = "The name of the API key"
  value       = var.retrieve_all ? null : one(data.openai_project_api_key.api_key[*].name)
  # Returns null if retrieve_all is true
}

output "created_at" {
  description = "When the API key was created"
  value       = var.retrieve_all ? null : one(data.openai_project_api_key.api_key[*].created_at)
  # Returns null if retrieve_all is true
}

output "last_used_at" {
  description = "When the API key was last used"
  value       = var.retrieve_all ? null : one(data.openai_project_api_key.api_key[*].last_used_at)
  # Returns null if retrieve_all is true
}

# Outputs for all API keys retrieval (retrieve_all = true)
output "api_keys" {
  description = "List of all API keys for the project (only available when retrieve_all = true)"
  value       = var.retrieve_all ? one(data.openai_project_api_keys.all_keys[*].api_keys) : null
  # Returns null if retrieve_all is false
} 