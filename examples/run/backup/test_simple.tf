# Minimal OpenAI Run Example
# This is a simplified version to test basic run functionality

terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

provider "openai" {}

# Create an assistant
resource "openai_assistant" "test" {
  name         = "Test Run Assistant"
  model        = "gpt-3.5-turbo"
  instructions = "You are a helpful assistant."
}

# Create a thread with a message
resource "openai_thread" "test" {
  messages {
    role    = "user"
    content = "Hello, please respond with 'Test successful!'"
  }
}

# Create a single run
resource "openai_run" "test" {
  thread_id    = openai_thread.test.id
  assistant_id = openai_assistant.test.id
}

# Outputs
output "run_status" {
  value = openai_run.test.status
}

output "run_id" {
  value = openai_run.test.id
}