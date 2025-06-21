# OpenAI File Management Module
# ==============================
# This module handles file uploads and management for the OpenAI API.
# Files can be used for various purposes like fine-tuning, assistants, batch processing, etc.

terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

# Input variables
variable "file_path" {
  description = "Path to the file to upload"
  type        = string
  default     = null
}

variable "file_id" {
  description = "ID of an existing file to retrieve (used when use_data_source is true)"
  type        = string
  default     = null
}

variable "purpose" {
  description = "Purpose of the file (fine-tune, assistants, batch, vision, user_data, evals)"
  type        = string
  default     = null

  validation {
    condition = var.purpose == null || anytrue([
      var.purpose == "fine-tune",
      var.purpose == "assistants",
      var.purpose == "batch",
      var.purpose == "vision",
      var.purpose == "user_data",
      var.purpose == "evals"
    ])
    error_message = "The purpose must be one of: fine-tune, assistants, batch, vision, user_data, evals."
  }
}

variable "project_id" {
  description = "The ID of the OpenAI project to associate this file with (for reference only)"
  type        = string
  default     = ""
}

variable "use_data_source" {
  description = "Whether to use the data source (true) to retrieve an existing file or create a new one (false)"
  type        = bool
  default     = false
}

variable "list_files" {
  description = "Whether to list all files (can be used alongside other operations)"
  type        = bool
  default     = false
}

variable "list_files_purpose" {
  description = "Purpose filter when listing files (optional)"
  type        = string
  default     = null
}

# Local variables to simplify conditional logic
locals {
  use_resource    = !var.use_data_source
  use_data_source = var.use_data_source
  use_list_files  = var.list_files

  # Check if the required parameters are provided based on the mode
  validate_resource = local.use_resource && var.file_path == null ? tobool("When use_data_source is false, file_path is required") : true
  validate_data     = local.use_data_source && var.file_id == null ? tobool("When use_data_source is true, file_id is required") : true
}

# Upload file to OpenAI (Resource mode)
resource "openai_file" "this" {
  count      = local.use_resource ? 1 : 0
  file       = var.file_path
  purpose    = var.purpose
  project_id = var.project_id != "" ? var.project_id : null
}

# Retrieve existing file from OpenAI (Data Source mode)
data "openai_file" "this" {
  count      = local.use_data_source ? 1 : 0
  file_id    = var.file_id
  project_id = var.project_id != "" ? var.project_id : null
}

# List all files from OpenAI (optional)
data "openai_files" "all" {
  count      = local.use_list_files ? 1 : 0
  purpose    = var.list_files_purpose
  project_id = var.project_id != "" ? var.project_id : null
}

# Outputs that work in both resource and data source mode
output "file_id" {
  description = "The ID of the file"
  value       = local.use_resource ? openai_file.this[0].id : (local.use_data_source ? data.openai_file.this[0].id : null)
}

output "filename" {
  description = "The name of the file"
  value       = local.use_resource ? openai_file.this[0].filename : (local.use_data_source ? data.openai_file.this[0].filename : null)
}

output "bytes" {
  description = "The size of the file in bytes"
  value       = local.use_resource ? openai_file.this[0].bytes : (local.use_data_source ? data.openai_file.this[0].bytes : null)
}

output "created_at" {
  description = "Timestamp when the file was created"
  value       = local.use_resource ? openai_file.this[0].created_at : (local.use_data_source ? data.openai_file.this[0].created_at : null)
}

output "purpose" {
  description = "The purpose of the file"
  value       = local.use_resource ? openai_file.this[0].purpose : (local.use_data_source ? data.openai_file.this[0].purpose : null)
}

# List mode outputs
output "all_files" {
  description = "List of all files (only when list_files = true)"
  value       = local.use_list_files ? data.openai_files.all[0].files : []
}

output "file_count" {
  description = "Number of files retrieved (only when list_files = true)"
  value       = local.use_list_files ? length(data.openai_files.all[0].files) : 0
}

output "files_by_purpose" {
  description = "Files grouped by purpose (only when list_files = true)"
  value = local.use_list_files ? {
    for file in data.openai_files.all[0].files :
    file.purpose => file.filename...
  } : {}
} 