terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

variable "model_id" {
  description = "The ID of the model to retrieve information for"
  type        = string
}

variable "api_key" {
  description = "Optional project-specific API key to use for this module"
  type        = string
  sensitive   = true
  default     = ""
}

data "openai_model" "model" {
  model_id = var.model_id
  api_key  = var.api_key != "" ? var.api_key : null
}

output "model_details" {
  description = "Complete details about the model"
  value = {
    id       = data.openai_model.model.id
    owned_by = data.openai_model.model.owned_by
    created  = data.openai_model.model.created
    object   = data.openai_model.model.object
  }
}

output "model_id" {
  description = "ID of the model"
  value       = data.openai_model.model.id
} 