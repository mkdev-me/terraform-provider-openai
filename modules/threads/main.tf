# OpenAI Threads Module

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
    }
  }
}

# Thread Resource
resource "openai_thread" "this" {
  count = var.enable_thread ? 1 : 0

  # Add initial messages if specified
  dynamic "messages" {
    for_each = var.thread_messages
    content {
      role     = messages.value.role
      content  = messages.value.content
      file_ids = try(messages.value.file_ids, null)
      metadata = try(messages.value.metadata, null)
    }
  }

  # Add metadata if specified
  metadata = var.thread_metadata
} 