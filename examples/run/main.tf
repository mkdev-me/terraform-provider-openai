# OpenAI Runs Example - Sequential Version
# ==============================
# This example demonstrates how to create and manage runs in OpenAI threads.
# Fixed to avoid "thread already has an active run" errors by using separate threads.

terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

provider "openai" {
  # API key is pulled from the OPENAI_API_KEY environment variable
}

# Create a single assistant that will be used by all runs
resource "openai_assistant" "example" {
  name         = "Example Run Assistant"
  model        = "gpt-4-turbo-preview"
  instructions = "You are a helpful AI assistant. Provide clear and concise answers."

  tools {
    type = "code_interpreter"
  }
}

# Example 1: Basic run on its own thread
resource "openai_thread" "basic" {
  messages {
    role    = "user"
    content = "What is 2 + 2? Just give me the number."
  }
}

resource "openai_run" "basic" {
  thread_id    = openai_thread.basic.id
  assistant_id = openai_assistant.example.id
}

# Example 2: Run with custom parameters on a separate thread
resource "openai_thread" "custom" {
  messages {
    role    = "user"
    content = "Explain the concept of recursion in one sentence."
  }
}

resource "openai_run" "custom" {
  thread_id    = openai_thread.custom.id
  assistant_id = openai_assistant.example.id

  # Override model
  model = "gpt-3.5-turbo"

  # Custom temperature
  temperature = 0.7

  metadata = {
    "type" = "custom_run"
  }
}

# Example 3: Combined thread and run creation
resource "openai_thread_run" "combined" {
  assistant_id = openai_assistant.example.id

  thread {
    messages {
      role    = "user"
      content = "What are the primary colors? List them."
    }
  }
}

# Data sources to read run information
data "openai_run" "basic_info" {
  run_id    = openai_run.basic.id
  thread_id = openai_thread.basic.id

  depends_on = [openai_run.basic]
}

# Outputs
output "assistant_id" {
  value = openai_assistant.example.id
}

output "basic_run" {
  value = {
    id     = openai_run.basic.id
    status = openai_run.basic.status
  }
}

output "custom_run" {
  value = {
    id     = openai_run.custom.id
    status = openai_run.custom.status
    model  = openai_run.custom.model
  }
}

output "combined_run" {
  value = {
    id        = openai_thread_run.combined.id
    thread_id = openai_thread_run.combined.thread_id
    status    = openai_thread_run.combined.status
  }
}

output "basic_run_details" {
  value = {
    status       = data.openai_run.basic_info.status
    created_at   = data.openai_run.basic_info.created_at
    completed_at = data.openai_run.basic_info.completed_at
  }
}