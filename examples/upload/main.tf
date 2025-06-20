# Example: OpenAI File Upload and Import
#
# This example demonstrates how to:
# 1. Upload a new file to OpenAI
# 2. Import an existing OpenAI file into Terraform
#
# To import an existing file, first configure the module as shown below,
# then run: terraform import module.fine_tune_upload.openai_file.file file-abc123
#
# The imported file will be managed by Terraform with a placeholder path
# and lifecycle configuration to ignore changes to the file path.

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "~> 1.0.0"
    }
  }
}

# Configure the OpenAI Provider
provider "openai" {
  # API key taken from OPENAI_API_KEY environment variable
}

# Local variables for file properties
locals {
  # Path to the training data file
  training_file_path = "${path.module}/training_data.jsonl"
}

# Use the upload module to manage the file
# This module works with both new uploads and imports
module "fine_tune_upload" {
  source = "../../modules/upload"

  # Required attributes
  purpose   = "fine-tune"
  file_path = "./training_data.jsonl"

  # For imported files, the file_path will be ignored
  # and replaced with a placeholder path
}

# Outputs
output "file_id" {
  description = "The ID of the uploaded file"
  value       = module.fine_tune_upload.file_id
}

output "file_bytes" {
  description = "The size of the file in bytes"
  value       = module.fine_tune_upload.bytes
}

output "file_created_at" {
  description = "The timestamp when the file was created"
  value       = module.fine_tune_upload.created_at
}

output "file_name" {
  description = "The name of the uploaded file"
  value       = module.fine_tune_upload.filename
}

# After importing a file, you can reference it in other resources
# For example, to use an imported file in fine-tuning:
/*
resource "openai_fine_tuning_job" "my_model" {
  training_file     = module.fine_tune_upload.file_id
  model             = "gpt-3.5-turbo"
  hyperparameters = {
    n_epochs = 3
  }
}
*/ 