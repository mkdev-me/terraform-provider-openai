###############################################################################
# OpenAI Admin API Key Module
###############################################################################
# This module creates an OpenAI Admin API key using the provider's resource

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
  }
}

###############################################################################
# Variables
###############################################################################

variable "name" {
  description = "The name of the admin API key"
  type        = string
}

variable "expires_at" {
  description = "Unix timestamp for when the API key expires"
  type        = number
  default     = null
}

variable "scopes" {
  description = "The scopes this key is restricted to (e.g. api.management.read, api.management.write)"
  type        = list(string)
  default     = []
}

variable "api_key" {
  description = "Custom API key to use for this operation. If not provided, the provider's default API key will be used."
  type        = string
  default     = null
  sensitive   = true
}

variable "list_keys" {
  description = "Whether to list all admin API keys"
  type        = bool
  default     = false
}

variable "list_limit" {
  description = "Maximum number of API keys to retrieve when listing"
  type        = number
  default     = 20
}

variable "list_after" {
  description = "API key ID to start listing from (for pagination)"
  type        = string
  default     = null
}

###############################################################################
# Resources
###############################################################################

# Create an admin API key
resource "openai_admin_api_key" "key" {
  name = var.name

  # Only set expires_at if provided
  expires_at = var.expires_at

  # Only set scopes if provided and not empty
  scopes = length(var.scopes) > 0 ? var.scopes : null

  # Only set api_key if provided
  api_key = var.api_key != null ? var.api_key : null
}

###############################################################################
# Data Sources
###############################################################################

# List all admin API keys if list_keys is true
data "openai_admin_api_keys" "all_keys" {
  count   = var.list_keys ? 1 : 0
  api_key = var.api_key != null ? var.api_key : null
  limit   = var.list_limit
  after   = var.list_after != null ? var.list_after : null
}

###############################################################################
# Outputs
###############################################################################

output "key_id" {
  description = "The ID of the API key"
  value       = openai_admin_api_key.key.id
}

output "key_value" {
  description = "The value of the API key. Only returned at creation time."
  value       = openai_admin_api_key.key.api_key_value
  sensitive   = true
}

output "created_at" {
  description = "Creation timestamp of the API key"
  value       = openai_admin_api_key.key.created_at
}

output "name" {
  description = "The name of the API key as set during creation"
  value       = openai_admin_api_key.key.name
}

output "expires_at" {
  description = "Expiration date of the API key (Unix timestamp, if specified)"
  value       = var.expires_at
}

output "all_api_keys" {
  description = "List of all admin API keys (only available if list_keys is true)"
  value       = var.list_keys ? data.openai_admin_api_keys.all_keys[0].api_keys : []
}

output "has_more_keys" {
  description = "Whether there are more API keys available beyond the limit"
  value       = var.list_keys ? data.openai_admin_api_keys.all_keys[0].has_more : false
} 