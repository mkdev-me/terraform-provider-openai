terraform {
  required_providers {
    openai = {
      source = "fjcorp/openai"
    }
  }
}

provider "openai" {
  # API key can be provided via environment variable OPENAI_API_KEY
  # Admin key can be provided via environment variable OPENAI_ADMIN_KEY
}

# Example 1: Retrieve information about a specific model using provider API key
data "openai_model" "gpt4o" {
  model_id = "gpt-4o"
}

# Example 2: Retrieve information about a specific model (alternate example)
data "openai_model" "gpt4o_project" {
  model_id = "gpt-4o"
}

# Example 3: Retrieve information about all available models
data "openai_models" "all" {}

# Variables for sensitive information that shouldn't be hardcoded

output "model_info" {
  description = "Details about the GPT-4o model"
  value = {
    id       = data.openai_model.gpt4o.id
    created  = data.openai_model.gpt4o.created
    owned_by = data.openai_model.gpt4o.owned_by
    object   = data.openai_model.gpt4o.object
  }
}

output "model_count" {
  description = "Number of available models"
  value       = length(data.openai_models.all.models)
}

output "available_models" {
  description = "IDs of all available models"
  value       = [for model in data.openai_models.all.models : model.id]
}
