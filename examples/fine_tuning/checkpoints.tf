# This file contains configuration for checkpoint permissions
# Comment or uncomment the resource block to control when it's applied

# Variables for checkpoint permissions
variable "admin_api_key" {
  description = "Admin API key for creating checkpoint permissions"
  type        = string
  default     = ""
  sensitive   = true
}

variable "use_admin_key" {
  description = "Whether admin key is being used for operations that require it"
  type        = bool
  default     = false
}

variable "project_id" {
  description = "The OpenAI project ID to grant permission to"
  type        = string
  default     = "proj_BkBLUjk3B5HZM4XD0uyuypfS"
}

# Optional variable to directly specify a checkpoint ID when the job is already completed
variable "checkpoint_id" {
  description = "Explicit checkpoint ID to use for permissions (if job is already completed)"
  type        = string
  default     = "ft:gpt-4o-2024-08-06:fodoj-gmbh::BGscDbx0" # Default for testing
}

# Get details of the fine-tuning job - only if it exists and is successful
# data "openai_fine_tuning_job" "job_details" { ... }

# Get checkpoints from the fine-tuning job - only if the job exists
# data "openai_fine_tuning_checkpoints" "job_checkpoints" { ... }

# Only create permission if admin key is provided and use_admin_key is true
locals {
  # Only create permission if admin key is provided and use_admin_key is true
  create_permission = var.use_admin_key && var.admin_api_key != "" && var.project_id != "" ? 1 : 0
}

# Use external data source with our script to create the checkpoint permission
# Commented out since we're now using the Terraform resource in main.tf
# data "external" "checkpoint_permission" {
#   # Only create when explicitly enabled via variable
#   count = var.use_admin_key ? 1 : 0
#
#   program = ["bash", "${path.module}/create_checkpoint_permission.sh"]
#   
#   # Pass arguments to the script
#   query = {
#     admin_api_key = var.admin_api_key
#     checkpoint_id = var.checkpoint_id
#     project_ids   = var.project_id
#   }
# }

# Admin key status output
output "admin_key_status" {
  value = var.use_admin_key ? "Admin key enabled - using external script for permission creation" : "Admin key not provided"
}

# Outputs for checkpoint information
output "job_status" {
  value = try(openai_fine_tuning_job.supervised_example.status, "Job status not available yet")
}

output "fine_tuned_model_id" {
  value = var.checkpoint_id
}

output "available_checkpoints" {
  value = "Using explicit checkpoint: ${var.checkpoint_id}"
}

output "checkpoint_status" {
  value = "Using explicit checkpoint"
}

# Output the permission ID if created
output "checkpoint_permission_id" {
  value = var.use_admin_key && var.checkpoint_id != "" && var.project_id != "" ? (
    length(openai_fine_tuning_checkpoint_permission.checkpoint_permission) > 0 ?
    openai_fine_tuning_checkpoint_permission.checkpoint_permission[0].id : "Not created yet"
  ) : "Not created (admin key not provided or parameters missing)"
  description = "ID of the checkpoint permission, if created"
}

# Output the permission details
output "checkpoint_permission_details" {
  value = var.use_admin_key && var.checkpoint_id != "" && var.project_id != "" ? (
    length(openai_fine_tuning_checkpoint_permission.checkpoint_permission) > 0 ? {
      checkpoint_id = openai_fine_tuning_checkpoint_permission.checkpoint_permission[0].checkpoint_id
      project_ids   = openai_fine_tuning_checkpoint_permission.checkpoint_permission[0].project_ids
      id            = openai_fine_tuning_checkpoint_permission.checkpoint_permission[0].id
      created_at    = openai_fine_tuning_checkpoint_permission.checkpoint_permission[0].created_at
      } : {
      checkpoint_id = var.checkpoint_id
      project_ids   = [var.project_id]
      id            = "Not available yet"
      created_at    = "Not available yet"
    }
    ) : {
    checkpoint_id = "Not created (admin key not provided)"
    project_ids   = []
    id            = "Not created (admin key not provided)"
    created_at    = "Not created (admin key not provided)"
  }
  description = "Details of the checkpoint permission"
}