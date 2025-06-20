terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

resource "openai_model_response" "this" {
  input = var.input
  model = var.model

  # Optional parameters with default values
  temperature       = var.temperature
  max_output_tokens = var.max_output_tokens
  top_p             = var.top_p
  top_k             = var.top_k
  frequency_penalty = var.frequency_penalty
  presence_penalty  = var.presence_penalty
  user              = var.user
  instructions      = var.instructions

  # List parameters
  stop_sequences = var.stop_sequences
  include        = var.include
} 