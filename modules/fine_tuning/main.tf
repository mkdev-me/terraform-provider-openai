terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
  }
}

# Simulate fine-tuning job with chat completion
resource "openai_chat_completion" "fine_tuning_simulation" {
  model = "gpt-3.5-turbo"

  messages {
    role    = "system"
    content = "You are a fine-tuning system that simulates the creation of fine-tuned models."
  }

  messages {
    role    = "user"
    content = <<-EOT
      Simulate creating a fine-tuned model with the following parameters:
      - Base model: ${var.model}
      - Training file: ${var.training_file}
      - Validation file: ${var.validation_file != null ? var.validation_file : "None"}
      - Suffix: ${var.suffix != null ? var.suffix : "None"}
      - Hyperparameters: ${var.hyperparameters != null ? jsonencode(var.hyperparameters) : "Default"}
    EOT
  }
}

# Local values to simulate fine-tuning outputs
locals {
  # Generate a deterministic ID based on input parameters
  model_id_suffix = sha256("${var.model}-${var.training_file}-${var.suffix != null ? var.suffix : "none"}")

  # Simulate fine-tuned model ID
  fine_tuned_model = "${var.model}:ft-${substr(local.model_id_suffix, 0, 12)}"

  # Simulate status
  status = "succeeded"

  # Simulate timestamps
  current_time = timestamp()
  created_at   = formatdate("YYYY-MM-DD'T'hh:mm:ssZ", local.current_time)
  finished_at  = var.completion_window > 0 ? formatdate("YYYY-MM-DD'T'hh:mm:ssZ", timeadd(local.current_time, "${var.completion_window}s")) : null
}

# Outputs
output "fine_tuned_model_id" {
  description = "The ID of the fine-tuned model created"
  value       = local.fine_tuned_model
}

output "status" {
  description = "The current status of the fine-tuning job"
  value       = local.status
}

output "created_at" {
  description = "The timestamp when the fine-tuning job was created"
  value       = local.created_at
}

output "finished_at" {
  description = "The timestamp when the fine-tuning job was completed"
  value       = local.finished_at
} 