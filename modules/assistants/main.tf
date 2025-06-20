# OpenAI Assistants Module

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
  }
}

# Assistants Resource
resource "openai_assistant" "this" {
  count        = var.enable_assistant ? 1 : 0
  name         = var.assistant_name
  model        = var.assistant_model
  instructions = var.assistant_instructions
  description  = var.assistant_description

  # Add tools if specified
  dynamic "tools" {
    for_each = var.assistant_tools
    content {
      type = tools.value.type

      # Add function configuration if it's a function tool
      dynamic "function" {
        for_each = tools.value.type == "function" && try(tools.value.function, null) != null ? [tools.value.function] : []
        content {
          name        = function.value.name
          description = try(function.value.description, null)
          parameters  = function.value.parameters
        }
      }
    }
  }

  # Add file IDs if specified
  file_ids = var.assistant_file_ids

  # Add metadata if specified
  metadata = var.assistant_metadata
} 