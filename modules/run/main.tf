# OpenAI Run Module
# ==============================
# This module simplifies the creation and management of runs in the OpenAI Assistants API.
# Runs are executions of an assistant on a thread, allowing the assistant to respond to messages.

terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

# Create the run using the openai_run resource
resource "openai_run" "run" {
  thread_id       = var.thread_id
  assistant_id    = var.assistant_id
  model           = var.model
  instructions    = var.instructions
  tools           = var.tools
  metadata        = var.metadata
  temperature     = var.temperature
  top_p           = var.top_p
  stream_for_tool = var.stream_for_tool
} 