terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

provider "openai" {
  # API key sourced from environment variable: OPENAI_API_KEY
}

# Empty Thread Example
resource "openai_thread" "empty_thread" {
  # This creates an empty thread with no initial messages or metadata
}

# Thread with Initial Messages Example
resource "openai_thread" "with_messages" {
  # Creating a thread with initial messages
  messages {
    role    = "user"
    content = "I need help understanding quantum computing concepts."
    # file_ids = ["file-abc123"] # Uncomment and replace with actual file IDs if needed
    metadata = {
      "importance" = "high"
      "category"   = "education"
    }
  }

  messages {
    role    = "user"
    content = "Specifically, I'd like to understand quantum entanglement."
  }

  # Thread-level metadata
  metadata = {
    "subject"    = "quantum_computing",
    "created_by" = "terraform",
    "priority"   = "medium"
  }
}

# Thread with Metadata Only
resource "openai_thread" "with_metadata" {
  # Creating a thread with only metadata
  metadata = {
    "purpose"     = "customer_support",
    "customer_id" = "cust_12345",
    "topic"       = "billing_inquiry"
  }
}

# Outputs
output "empty_thread_id" {
  description = "The ID of the empty thread"
  value       = openai_thread.empty_thread.id
}

output "with_messages_thread_id" {
  description = "The ID of the thread with initial messages"
  value       = openai_thread.with_messages.id
}

output "with_metadata_thread_id" {
  description = "The ID of the thread with only metadata"
  value       = openai_thread.with_metadata.id
} 