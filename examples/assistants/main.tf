terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

provider "openai" {
  # API key sourced from environment variable: OPENAI_API_KEY
}

# Assistant Example
resource "openai_assistant" "math_tutor" {
  name         = "Math Tutor"
  model        = "gpt-4o"
  instructions = "You are a personal math tutor. When asked a question, write and run Python code to answer the question."

  tools {
    type = "code_interpreter"
  }

  description = "A math tutor that uses code interpreter to solve problems"
  metadata = {
    "created_by" = "terraform",
    "version"    = "1.0"
  }
}

# Creating a second assistant to demonstrate listing functionality
resource "openai_assistant" "writing_assistant" {
  name         = "Writing Assistant"
  model        = "gpt-4o"
  instructions = "You are a helpful assistant for writing and editing content."
  description  = "An assistant that helps with writing tasks"
  metadata = {
    "created_by" = "terraform",
    "purpose"    = "content_creation"
  }
}

# Data sources to retrieve assistant information
data "openai_assistants" "all" {
  # This data source fetches all assistants
  # Optional parameters:
  # limit  = 10   # Number of assistants to retrieve (1-100)
  # order  = "desc" # Sort order (asc or desc)
  # after  = "asst_abc123" # Pagination cursor
  # before = "asst_xyz456" # Pagination cursor
}

# Outputs
output "math_tutor_id" {
  description = "The ID of the created math tutor assistant"
  value       = openai_assistant.math_tutor.id
}

output "math_tutor_created_at" {
  description = "Timestamp when the math tutor assistant was created"
  value       = openai_assistant.math_tutor.created_at
}

output "writing_assistant_id" {
  description = "The ID of the created writing assistant"
  value       = openai_assistant.writing_assistant.id
}

output "all_assistants_count" {
  description = "The total number of assistants retrieved by the data source"
  value       = length(data.openai_assistants.all.assistants)
}

output "all_assistants_ids" {
  description = "The IDs of all assistants retrieved by the data source"
  value       = [for a in data.openai_assistants.all.assistants : a.id]
} 