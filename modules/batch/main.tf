# OpenAI Batch Processing Module
# ==============================
# This module supports both direct batch creation and data retrieval options
# for OpenAI batch processing jobs.

terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

# Input variables
variable "input_file_id" {
  description = "The ID of the file containing the inputs for the batch (must be uploaded with purpose='batch')"
  type        = string
}

variable "project_id" {
  description = "The ID of the OpenAI project to use for this batch. If not specified, the default project will be used."
  type        = string
  default     = ""
}

variable "endpoint" {
  description = "The endpoint to use for all requests in the batch (e.g., '/v1/chat/completions')"
  type        = string

  validation {
    condition     = contains(["/v1/responses", "/v1/chat/completions", "/v1/embeddings", "/v1/completions"], var.endpoint)
    error_message = "The endpoint must be one of: '/v1/responses', '/v1/chat/completions', '/v1/embeddings', '/v1/completions'."
  }
}

variable "completion_window" {
  description = "The time frame within which the batch should be processed (currently only '24h' is supported)"
  type        = string
  default     = "24h"
}

variable "model" {
  description = "The ID of the model to use for this batch"
  type        = string
}

variable "metadata" {
  description = "Set of key-value pairs that can be attached to the batch object (max 16 pairs)"
  type        = map(string)
  default     = {}
}

variable "list_mode" {
  description = "Whether to list all batches or retrieve a specific batch"
  type        = bool
  default     = false
}

variable "batch_id" {
  description = "The ID of the batch to retrieve"
  type        = string
  default     = ""
}

variable "api_key" {
  description = "The API key to use for accessing OpenAI"
  type        = string
  default     = ""
}

# If list_mode is true, use the openai_batches data source to get all batches
data "openai_batches" "all_batches" {
  count      = var.list_mode ? 1 : 0
  project_id = var.project_id
  api_key    = var.api_key
}

# If list_mode is false but batch_id is provided, retrieve a specific batch
data "openai_batch" "specific_batch" {
  count      = (!var.list_mode && var.batch_id != "") ? 1 : 0
  batch_id   = var.batch_id
  project_id = var.project_id
  api_key    = var.api_key
}

# Create a new batch only if not in list_mode and no batch_id is provided
resource "openai_batch" "new_batch" {
  count             = (!var.list_mode && var.batch_id == "" && var.input_file_id != "") ? 1 : 0
  input_file_id     = var.input_file_id
  endpoint          = var.endpoint
  model             = var.model
  completion_window = var.completion_window
  project_id        = var.project_id
  metadata          = var.metadata
}

# Local values
locals {
  # Determine if we have a list of batches
  has_batch_list = var.list_mode && length(data.openai_batches.all_batches) > 0

  # Determine if we have a specific batch 
  has_specific_batch = !var.list_mode && var.batch_id != "" && length(data.openai_batch.specific_batch) > 0

  # Determine if we created a new batch
  has_new_batch = !var.list_mode && var.batch_id == "" && length(openai_batch.new_batch) > 0

  # Get all batches when in list mode
  all_batches = local.has_batch_list ? try(data.openai_batches.all_batches[0].batches, []) : []

  # Set the batch properties based on which source we're using
  batch_id = local.has_specific_batch ? data.openai_batch.specific_batch[0].id : (
    local.has_new_batch ? openai_batch.new_batch[0].id : ""
  )

  status = local.has_specific_batch ? data.openai_batch.specific_batch[0].status : (
    local.has_new_batch ? openai_batch.new_batch[0].status : "unknown"
  )

  input_file_id = local.has_specific_batch ? data.openai_batch.specific_batch[0].input_file_id : (
    local.has_new_batch ? openai_batch.new_batch[0].input_file_id : var.input_file_id
  )

  output_file_id = local.has_specific_batch ? data.openai_batch.specific_batch[0].output_file_id : (
    local.has_new_batch ? openai_batch.new_batch[0].output_file_id : ""
  )

  created_at = local.has_specific_batch ? data.openai_batch.specific_batch[0].created_at : (
    local.has_new_batch ? openai_batch.new_batch[0].created_at : 0
  )

  expires_at = local.has_specific_batch ? data.openai_batch.specific_batch[0].expires_at : (
    local.has_new_batch ? openai_batch.new_batch[0].expires_at : 0
  )
}

# Outputs
output "batch_id" {
  description = "The unique identifier for the batch job"
  value       = local.batch_id
}

output "status" {
  description = "The status of the batch job"
  value       = local.status
}

output "input_file_id" {
  description = "The ID of the input file used for the batch"
  value       = local.input_file_id
}

output "output_file_id" {
  description = "The ID of the output file (if available)"
  value       = local.output_file_id
}

output "created_at" {
  description = "The timestamp when the batch job was created"
  value       = local.created_at
}

output "expires_at" {
  description = "The timestamp when the batch job expires"
  value       = local.expires_at
}

output "all_batches" {
  description = "List of all batches (only populated when list_mode = true)"
  value       = local.all_batches
} 