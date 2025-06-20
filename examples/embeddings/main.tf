# OpenAI Embeddings Example
# This example demonstrates how to generate and use text embeddings with OpenAI

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
  }
}

# Configure the OpenAI Provider
provider "openai" {
  # API key will be sourced from environment variable OPENAI_API_KEY
  # Organization ID will be sourced from environment variable OPENAI_ORGANIZATION_ID
}

# Example 1: Basic text embedding
module "simple_embedding" {
  source = "../../modules/embeddings"

  input = "The food was delicious and the waiter was very friendly."
  model = "text-embedding-ada-002"
}

# Example 2: Embedding with different format (base64)
module "base64_embedding" {
  source = "../../modules/embeddings"

  input           = "Convert this text to a base64 embedding."
  model           = "text-embedding-ada-002"
  encoding_format = "base64"
}

# Example 3: Multiple texts in a single request
locals {
  multiple_texts = jsonencode([
    "First text to embed",
    "Second text to embed",
    "Third text to embed with different content"
  ])
}

module "multiple_embeddings" {
  source = "../../modules/embeddings"

  input = local.multiple_texts
  model = "text-embedding-ada-002"
}

# Example 4: Using a newer model with dimensions specification
# Note: text-embedding-3 models support specifying dimensions
module "embedding_with_dimensions" {
  source = "../../modules/embeddings"

  input      = "Generate an embedding with custom dimensions."
  model      = "text-embedding-3-small" # Requires OpenAI API that supports this model
  dimensions = 256                      # Specify custom dimensions (if supported by the model)
}

# Outputs
output "simple_embedding_usage" {
  description = "Token usage for the simple embedding"
  value       = module.simple_embedding.usage
}

output "multiple_embeddings_count" {
  description = "Number of embeddings generated in the batch request"
  value       = length(module.multiple_embeddings.embeddings)
}

# The actual embeddings are marked as sensitive to avoid cluttering the output
output "simple_embedding_id" {
  description = "ID of the simple embedding"
  value       = module.simple_embedding.embedding_id
}

output "base64_embedding_model" {
  description = "Model used for the base64 encoding"
  value       = module.base64_embedding.model_used
} 