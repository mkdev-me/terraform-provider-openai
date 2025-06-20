# OpenAI Runs Example - Fixed Version
# ==============================
# This example demonstrates how to create and manage runs in OpenAI threads using the Assistants API.
# Fixed to ensure proper sequential creation of runs.

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
  }
}

provider "openai" {
  # API key is pulled from the OPENAI_API_KEY environment variable
}

# Create an assistant for this example
resource "openai_assistant" "example_assistant" {
  name         = "Example Run Assistant"
  model        = "gpt-4o"
  instructions = "You are a helpful AI assistant created for demonstrating run operations in Terraform."

  tools {
    type = "code_interpreter"
  }

  metadata = {
    "created_by" = "terraform",
    "purpose"    = "example"
  }
}

# Create threads for our runs
resource "openai_thread" "example_thread" {
  metadata = {
    "source"   = "terraform_example",
    "purpose"  = "run_demonstrations",
    "run_type" = "basic_run"
  }
}

resource "openai_thread" "custom_thread" {
  metadata = {
    "source"   = "terraform_example",
    "purpose"  = "run_demonstrations",
    "run_type" = "custom_run"
  }
}

# Add messages to threads
resource "openai_message" "initial_message" {
  thread_id = openai_thread.example_thread.id
  role      = "user"
  content   = "Hello! Can you help me understand how runs work in OpenAI's Assistants API?"

  metadata = {
    "type"   = "initial_question",
    "source" = "terraform_example"
  }
}

resource "openai_message" "custom_message" {
  thread_id = openai_thread.custom_thread.id
  role      = "user"
  content   = "Please explain what makes a run 'custom' and how it differs from a standard run."

  metadata = {
    "type"   = "custom_question",
    "source" = "terraform_example"
  }
}

# Example 1: Basic run
resource "openai_run" "basic_run" {
  thread_id    = openai_thread.example_thread.id
  assistant_id = openai_assistant.example_assistant.id

  # Wait for the initial message to be created first
  depends_on = [openai_message.initial_message]
}

# Example 2: Custom run with sequential dependency
resource "openai_run" "custom_run" {
  thread_id    = openai_thread.custom_thread.id
  assistant_id = openai_assistant.example_assistant.id

  # Override the assistant's model for this run
  model = "gpt-3.5-turbo"

  # Override instructions for this specific run
  instructions = "For this run only, be extremely concise in your answers and use simple language."

  # Set custom parameters
  temperature = 0.7

  # Add metadata for this run
  metadata = {
    "run_type" = "custom_parameters",
    "purpose"  = "demonstration"
  }

  # Ensure sequential creation
  depends_on = [
    openai_message.custom_message,
    openai_run.basic_run # Wait for the first run to complete
  ]
}

# Example 3: Combined thread and run operation
# This creates its own thread, so no conflicts
resource "openai_thread_run" "combined_operation" {
  assistant_id = openai_assistant.example_assistant.id

  thread {
    messages {
      role    = "user"
      content = "What are the key benefits of using the thread run combined endpoint?"
    }

    metadata = {
      "purpose" = "demonstration",
      "type"    = "combined_thread_run"
    }
  }

  instructions = "Provide a clear and concise response about the benefits."

  # Ensure this is created after other runs
  depends_on = [openai_run.custom_run]
}

# Outputs
output "assistant_id" {
  description = "The ID of the created assistant"
  value       = openai_assistant.example_assistant.id
}

output "thread_ids" {
  description = "The IDs of the threads used for runs"
  value = {
    basic_thread  = openai_thread.example_thread.id
    custom_thread = openai_thread.custom_thread.id
  }
}

output "basic_run_status" {
  description = "Status of the basic run"
  value       = openai_run.basic_run.status
}

output "custom_run_status" {
  description = "Status of the custom run"
  value       = openai_run.custom_run.status
}

output "combined_thread_run_status" {
  description = "Status of the combined thread run"
  value       = openai_thread_run.combined_operation.status
}