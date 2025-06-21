# OpenAI Upload Module
# ==============================
# This module creates and manages file uploads in OpenAI using the provider resource.
# It supports both creating new file uploads and importing existing files.
#
# To import an existing file:
# 1. Configure this module with the required purpose (other attributes are optional)
# 2. Run: terraform import module.MODULE_NAME.openai_file.file file-XYZ123
# 3. The module will automatically detect import mode and handle the file path appropriately
#
# For imported files:
# - A placeholder file path is used ("./placeholder-for-import.file")
# - The lifecycle configuration ignores changes to the file path
# - All file metadata is fetched from the OpenAI API

terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

# Input variables
variable "purpose" {
  description = "The intended purpose of the uploaded file. Can be 'fine-tune', 'assistants', or 'vision'."
  type        = string

  validation {
    condition     = contains(["fine-tune", "assistants", "vision", "batch", "user_data", "evals"], var.purpose)
    error_message = "The purpose must be one of: 'fine-tune', 'assistants', 'vision', 'batch', 'user_data', 'evals'."
  }
}

variable "file_path" {
  description = "Path to the file to upload."
  type        = string
  default     = ""
}

variable "project_id" {
  description = "The ID of the OpenAI project to associate this upload with (for Terraform reference only)."
  type        = string
  default     = ""
}

# Detect if we're in import mode
locals {
  is_imported = (var.file_path == "" ||
    var.file_path == "./placeholder-for-import.file" ||
  fileexists(var.file_path) == false)
}

# Create the file upload using the proper OpenAI resource
resource "openai_file" "file" {
  file       = local.is_imported ? "./placeholder-for-import.file" : var.file_path
  purpose    = var.purpose
  project_id = var.project_id != "" ? var.project_id : null

  lifecycle {
    ignore_changes = [
      file,
    ]
  }
}

# Outputs
output "file_id" {
  description = "The unique identifier for the uploaded file"
  value       = openai_file.file.id
}

output "filename" {
  description = "The name of the uploaded file"
  value       = openai_file.file.filename
}

output "bytes" {
  description = "The size of the file in bytes"
  value       = openai_file.file.bytes
}

output "created_at" {
  description = "The timestamp when the upload was created"
  value       = openai_file.file.created_at
}

output "purpose" {
  description = "The purpose of the file"
  value       = openai_file.file.purpose
} 