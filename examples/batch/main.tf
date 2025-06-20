# OpenAI Batch Processing Example
# This example demonstrates using the openai_batch resource for batch processing

terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "~> 1.0.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "2.4.0"
    }
  }
}

# Configure the OpenAI Provider
provider "openai" {
  # API key taken from OPENAI_API_KEY environment variable
}

variable "openai_api_key" {
  description = "OpenAI API Key"
  type        = string
  sensitive   = true
  default     = ""
}

# Local variables for sample files
locals {
  # Sample JSONL content for embeddings batch
  embedding_batch_content = <<EOT
{"custom_id": "embed-1", "method": "POST", "url": "/v1/embeddings", "body": {"model": "text-embedding-ada-002", "input": "The food was delicious and the service was excellent."}}
{"custom_id": "embed-2", "method": "POST", "url": "/v1/embeddings", "body": {"model": "text-embedding-ada-002", "input": "I had a terrible experience at the restaurant."}}
{"custom_id": "embed-3", "method": "POST", "url": "/v1/embeddings", "body": {"model": "text-embedding-ada-002", "input": "The price was reasonable for the quality of food."}}
EOT

  # Sample JSONL content for chat completions batch
  chat_batch_content = <<EOT
{"custom_id": "request-1", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Explain quantum computing in simple terms"}]}}
{"custom_id": "request-2", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Write a short poem about artificial intelligence"}]}}
{"custom_id": "request-3", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "How can I improve my time management skills?"}]}}
EOT

  # Paths for local files
  embedding_file_path = "${path.module}/embedding_requests.jsonl"
  chat_file_path      = "${path.module}/chat_requests.jsonl"
}

# Create local files for uploading
resource "local_file" "embedding_requests" {
  content  = local.embedding_batch_content
  filename = local.embedding_file_path
}

resource "local_file" "chat_requests" {
  content  = local.chat_batch_content
  filename = local.chat_file_path
}

# Upload files to OpenAI
resource "openai_file" "embedding_file" {
  file    = local_file.embedding_requests.filename
  purpose = "batch"

  depends_on = [local_file.embedding_requests]
}

resource "openai_file" "chat_requests" {
  file    = local_file.chat_requests.filename
  purpose = "batch"

  depends_on = [local_file.chat_requests]
}

# Example 1: Create a batch job for embeddings
resource "openai_batch" "embedding_batch" {
  input_file_id     = openai_file.embedding_file.id
  endpoint          = "/embeddings"
  completion_window = "24h"
}

# Example 2: Create a batch job for chat completions
resource "openai_batch" "chat_batch" {
  input_file_id     = openai_file.chat_requests.id
  endpoint          = "/chat/completions"
  completion_window = "24h"
}

# Example 3: Using metadata
resource "openai_batch" "chat_batch_with_metadata" {
  input_file_id     = openai_file.chat_requests.id
  endpoint          = "/chat/completions"
  completion_window = "24h"

  metadata = {
    environment = "test"
    purpose     = "demo"
    source      = "terraform-example"
  }
}

# Example 4: Using the batch data source to monitor a batch job
data "openai_batch" "monitor_embedding_batch" {
  batch_id = openai_batch.embedding_batch.id
}

# Instead of using count with a conditional that depends on values not known at plan time,
# we'll use locals and outputs to provide the same functionality
locals {
  # This will be an empty string if the batch is not completed
  output_file_id = data.openai_batch.monitor_embedding_batch.status == "completed" ? data.openai_batch.monitor_embedding_batch.output_file_id : ""
}

# Data source outputs
output "monitored_batch_status" {
  description = "The current status of the monitored batch job"
  value       = data.openai_batch.monitor_embedding_batch.status
}

output "monitored_batch_request_counts" {
  description = "Request processing statistics for the monitored batch job"
  value       = data.openai_batch.monitor_embedding_batch.request_counts
}

output "batch_results_available" {
  description = "Whether the batch results are available"
  value       = data.openai_batch.monitor_embedding_batch.status == "completed"
}

# This output will contain the output_file_id when the batch is completed
output "batch_output_file_id" {
  description = "The output file ID to retrieve the results (only available when completed)"
  value       = local.output_file_id
}

# Outputs
output "embedding_batch_id" {
  description = "The ID of the embeddings batch job"
  value       = openai_batch.embedding_batch.id
}

output "embedding_batch_status" {
  description = "The status of the embeddings batch job"
  value       = openai_batch.embedding_batch.status
}

output "embedding_output_file" {
  description = "The output file ID for the embeddings batch job"
  value       = openai_batch.embedding_batch.output_file_id
}

output "chat_batch_id" {
  description = "The ID of the chat batch job"
  value       = openai_batch.chat_batch.id
}

output "chat_batch_status" {
  description = "The status of the chat batch job"
  value       = openai_batch.chat_batch.status
}

output "chat_output_file_id" {
  description = "The output file ID for the chat batch job"
  value       = openai_batch.chat_batch.output_file_id
}
