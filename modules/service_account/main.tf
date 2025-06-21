# OpenAI Service Account Module
# ================================
# This module creates and manages service accounts within OpenAI projects.
# Service accounts are bot users that are not associated with individual human users,
# making them ideal for applications and automation.

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
    }
  }
}

# Required variables
variable "project_id" {
  description = "The ID of the project where the service account will be created"
  type        = string
}

variable "name" {
  description = "The name of the service account"
  type        = string
  default     = null
}

# Optional variables
variable "openai_admin_key" {
  description = "Custom API key to use for this resource. If not provided, the provider's default API key will be used"
  type        = string
  default     = null
  sensitive   = true
}

variable "use_data_source" {
  description = "Whether to use a data source for existing service accounts"
  type        = bool
  default     = false
}

variable "service_account_id" {
  description = "The ID of an existing service account to look up (required when use_data_source is true)"
  type        = string
  default     = null
}

# If use_data_source is true, retrieve an existing service account
# This is wrapped in a try block to handle permission errors gracefully
data "openai_project_service_account" "this" {
  count              = var.use_data_source ? 1 : 0
  project_id         = var.project_id
  service_account_id = var.service_account_id != null ? var.service_account_id : ""
}

# Otherwise, create a new service account - but only if name is provided
resource "openai_project_service_account" "this" {
  count      = (!var.use_data_source && var.name != null && var.name != "") ? 1 : 0
  project_id = var.project_id
  name       = var.name
}

# Local variables to handle both data source and resource outputs
locals {
  # Safely check if resources or data sources exist and can be accessed
  has_data_source = var.use_data_source && try(
    length(data.openai_project_service_account.this) > 0 &&
    data.openai_project_service_account.this[0].id != "",
    false
  )
  has_resource = !var.use_data_source && try(
    length(openai_project_service_account.this) > 0,
    false
  )

  # Set values based on whether we're using data source or resource
  # Use try() to safely handle errors
  id = try(
    local.has_data_source ? data.openai_project_service_account.this[0].id :
    (local.has_resource ? openai_project_service_account.this[0].id : ""),
    ""
  )
  service_account_id = try(
    local.has_data_source ? data.openai_project_service_account.this[0].service_account_id :
    (local.has_resource ? openai_project_service_account.this[0].service_account_id : ""),
    ""
  )
  name = try(
    local.has_data_source ? data.openai_project_service_account.this[0].name :
    (local.has_resource ? openai_project_service_account.this[0].name : var.name),
    var.name
  )
  created_at = try(
    local.has_data_source ? data.openai_project_service_account.this[0].created_at :
    (local.has_resource ? openai_project_service_account.this[0].created_at : 0),
    0
  )
  role = try(
    local.has_data_source ? data.openai_project_service_account.this[0].role :
    (local.has_resource ? openai_project_service_account.this[0].role : ""),
    ""
  )
  api_key_id = try(
    local.has_data_source ? data.openai_project_service_account.this[0].api_key_id :
    (local.has_resource ? openai_project_service_account.this[0].api_key_id : ""),
    ""
  )
  api_key_value = try(
    local.has_resource ? openai_project_service_account.this[0].api_key_value : "",
    ""
  )
}

# Outputs
output "id" {
  description = "The composite ID of the service account (project_id:service_account_id)"
  value       = local.id
}

output "service_account_id" {
  description = "The ID of the service account"
  value       = local.service_account_id
}

output "name" {
  description = "The name of the service account"
  value       = local.name
}

output "created_at" {
  description = "The timestamp when the service account was created"
  value       = local.created_at
}

output "role" {
  description = "The role of the service account"
  value       = local.role
}

output "project_id" {
  description = "The project ID where the service account was created"
  value       = var.project_id
}

output "api_key_id" {
  description = "The ID of the API key for the service account"
  value       = local.api_key_id
}

output "api_key_value" {
  description = "The value of the API key (only available when creating a new service account)"
  value       = local.api_key_value
  sensitive   = true
} 