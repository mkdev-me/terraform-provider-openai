# OpenAI Runs Example
# ==============================
# This example demonstrates how to create and manage runs in OpenAI threads using the Assistants API.
# Runs are executions of an assistant on a thread, allowing the assistant to respond to messages.

terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

provider "openai" {
  # API key is pulled from the OPENAI_API_KEY environment variable
  # Organization ID can be set with OPENAI_ORGANIZATION environment variable (optional)
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

# Create a thread for our basic run
resource "openai_thread" "example_thread" {
  metadata = {
    "source"   = "terraform_example",
    "purpose"  = "run_demonstrations",
    "run_type" = "basic_run"
  }
}

# Create a thread for our custom run
resource "openai_thread" "custom_thread" {
  metadata = {
    "source"   = "terraform_example",
    "purpose"  = "run_demonstrations",
    "run_type" = "custom_run"
  }
}

# Create a thread for our module run
resource "openai_thread" "module_thread" {
  metadata = {
    "source"   = "terraform_example",
    "purpose"  = "run_demonstrations",
    "run_type" = "module_run"
  }
}

# Add an initial message to the first thread
resource "openai_message" "initial_message" {
  thread_id = openai_thread.example_thread.id
  role      = "user"
  content   = "Hello! Can you help me understand how runs work in OpenAI's Assistants API? Also, please calculate the sum of numbers from 1 to 10."

  metadata = {
    "type"   = "initial_question",
    "source" = "terraform_example"
  }
}

# Add an initial message to the custom thread
resource "openai_message" "custom_message" {
  thread_id = openai_thread.custom_thread.id
  role      = "user"
  content   = "Please explain what makes a run 'custom' and how it differs from a standard run. Give me some examples."

  metadata = {
    "type"   = "custom_question",
    "source" = "terraform_example"
  }
}

# Add an initial message to the module thread
resource "openai_message" "module_message" {
  thread_id = openai_thread.module_thread.id
  role      = "user"
  content   = "Explain how modules help organize Terraform code and improve reusability. Give a simple example."

  metadata = {
    "type"   = "module_question",
    "source" = "terraform_example"
  }
}

# Example 1: Basic run using the resource directly
resource "openai_run" "basic_run" {
  thread_id    = openai_thread.example_thread.id
  assistant_id = openai_assistant.example_assistant.id

  # Wait for the initial message to be created first
  depends_on = [openai_message.initial_message]
}

# Example 2: Run with custom parameters
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

  # Wait for the message to be created first
  depends_on = [openai_message.custom_message]
}

# Example 3: Using the runs module
module "module_run" {
  source = "../../modules/run"

  thread_id    = openai_thread.module_thread.id
  assistant_id = openai_assistant.example_assistant.id

  # Custom settings
  temperature = 0.5

  metadata = {
    "run_type" = "module_run",
    "purpose"  = "demonstration"
  }

  # Wait for the message to be created first
  depends_on = [openai_message.module_message]
}

# Example 4: Create a thread and run in one operation
resource "openai_thread_run" "combined_operation" {
  assistant_id = openai_assistant.example_assistant.id

  thread {
    messages {
      role    = "user"
      content = "What are the key benefits of using the thread run combined endpoint instead of separate thread and run resources?"
    }

    metadata = {
      "purpose" = "demonstration",
      "type"    = "combined_thread_run"
    }
  }

  instructions = "Provide a clear and concise response about the benefits of using the combined thread+run endpoint."
}

# Retrieve information about an existing run using the data source
data "openai_run" "run_info" {
  run_id    = openai_run.basic_run.id
  thread_id = openai_thread.example_thread.id

  depends_on = [openai_run.basic_run]
}

# Retrieve information about a thread run using the data source
data "openai_thread_run" "thread_run_info" {
  run_id    = openai_thread_run.combined_operation.id
  thread_id = openai_thread_run.combined_operation.thread_id

  depends_on = [openai_thread_run.combined_operation]
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
    module_thread = openai_thread.module_thread.id
  }
}

output "basic_run" {
  description = "Details of the basic run"
  value = {
    id           = openai_run.basic_run.id
    status       = openai_run.basic_run.status
    created_at   = openai_run.basic_run.created_at
    started_at   = openai_run.basic_run.started_at
    completed_at = openai_run.basic_run.completed_at
  }
}

output "custom_run" {
  description = "Details of the run with custom parameters"
  value = {
    id         = openai_run.custom_run.id
    status     = openai_run.custom_run.status
    model      = openai_run.custom_run.model
    created_at = openai_run.custom_run.created_at
  }
}

output "module_run" {
  description = "Details of the run created using the module"
  value = {
    id         = module.module_run.run_id
    status     = module.module_run.status
    created_at = module.module_run.created_at
  }
}

output "combined_thread_run" {
  description = "Details of the combined thread and run operation"
  value = {
    run_id     = openai_thread_run.combined_operation.id
    thread_id  = openai_thread_run.combined_operation.thread_id
    status     = openai_thread_run.combined_operation.status
    created_at = openai_thread_run.combined_operation.created_at
  }
}

output "run_info_from_data_source" {
  description = "Run information retrieved via data source"
  value = {
    id           = data.openai_run.run_info.id
    status       = data.openai_run.run_info.status
    assistant_id = data.openai_run.run_info.assistant_id
    usage        = data.openai_run.run_info.usage
  }
}

output "thread_run_info_from_data_source" {
  description = "Thread run information retrieved via data source"
  value = {
    id           = data.openai_thread_run.thread_run_info.id
    thread_id    = data.openai_thread_run.thread_run_info.thread_id
    status       = data.openai_thread_run.thread_run_info.status
    assistant_id = data.openai_thread_run.thread_run_info.assistant_id
  }
} 