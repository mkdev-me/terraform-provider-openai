# OpenAI Embeddings Module (Simulated)
# ======================
# This module simulates the creation of embeddings since the openai_embedding resource
# is not available at runtime
#
# IMPORTANT: API and Import Limitations
# ------------------------------------
# The OpenAI API does not provide endpoints to retrieve previously created embeddings.
# When importing:
# 1. Only basic metadata is imported (ID, timestamps)
# 2. The actual embedding vectors cannot be retrieved
# 3. After import, apply will recreate the resource since the original state can't be fully retrieved
#
# This module uses a simulation approach with fault-tolerant design to handle both new and imported resources.

terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

# Input variables
variable "model" {
  description = "ID of the model to use for generating embeddings (e.g., text-embedding-ada-002)"
  type        = string
  default     = "text-embedding-ada-002"
}

variable "input" {
  description = "Text to embed, can be a string or a JSON array of strings"
  type        = string
}

variable "user" {
  description = "Optional unique identifier representing your end-user"
  type        = string
  default     = null
}

variable "encoding_format" {
  description = "Format to return the embeddings in, either 'float' or 'base64'"
  type        = string
  default     = "float"

  validation {
    condition     = contains(["float", "base64"], var.encoding_format)
    error_message = "The encoding_format must be either 'float' or 'base64'."
  }
}

variable "dimensions" {
  description = "Number of dimensions the resulting output embeddings should have (only for text-embedding-3 and later models)"
  type        = number
  default     = null
}

variable "project_id" {
  description = "The OpenAI project ID to use for this request"
  type        = string
  default     = null
}

# Simulate the creation of embeddings using chat completion
resource "openai_chat_completion" "embedding_simulation" {
  model = "gpt-3.5-turbo"

  messages {
    role    = "system"
    content = "You are an embedding system that simulates the creation of text embeddings."
  }

  messages {
    role    = "user"
    content = "Simulate creating embeddings for the following text with model=${var.model}, encoding_format=${var.encoding_format}${var.dimensions != null ? ", dimensions=${var.dimensions}" : ""}: ${var.input}"
  }
}

# Simulated values
locals {
  # Generate a unique identifier based on the input
  embedding_id = "emb_sim_${sha256(var.input)}"

  # Generate a simulated vector of size 5 (in reality would be hundreds or thousands of dimensions)
  simulated_embeddings = [
    for i in range(5) : 0.1 * (i + 1) + 0.05 * length(var.input) % 0.7
  ]

  # Simulate token usage statistics
  simulated_usage = {
    prompt_tokens = length(var.input) / 4
    total_tokens  = length(var.input) / 4
  }

  # Extract content safely with fallback
  simulation_content = try(
    openai_chat_completion.embedding_simulation.choices[0].message[0].content,
    "Embedding simulation response unavailable"
  )
}

# Outputs
output "embeddings" {
  description = "The simulated embeddings (note: these are not real embeddings)"
  value       = local.simulated_embeddings
}

output "model_used" {
  description = "The model that would have been used (simulated)"
  value       = var.model
}

output "usage" {
  description = "Simulated usage statistics"
  value       = local.simulated_usage
}

output "embedding_id" {
  description = "The simulated embedding ID"
  value       = local.embedding_id
}

output "simulation_response" {
  description = "The response from the model simulating the creation of embeddings"
  value       = local.simulation_content
} 