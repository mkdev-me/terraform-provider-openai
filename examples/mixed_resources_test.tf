# Test example that uses both project and admin API keys
# This verifies that the provider correctly routes to the appropriate key

terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

provider "openai" {
  # Keys are loaded from environment variables:
  # - OPENAI_API_KEY for project operations
  # - OPENAI_ADMIN_KEY for admin operations
}

# ========== Admin Key Resources ==========

# Create a project (requires admin key)
resource "openai_project" "test" {
  name = "Mixed Resources Test Project"
}

# List all projects (requires admin key)
data "openai_projects" "all" {
  depends_on = [openai_project.test]
}

# ========== Project Key Resources ==========

# Create an embedding (requires project key)
resource "openai_embedding" "test" {
  model = "text-embedding-ada-002"
  input = ["Hello, this is a test embedding"]
}

# Generate a chat completion (requires project key)
data "openai_chat_completion" "test" {
  model = "gpt-3.5-turbo"

  messages = [
    {
      role    = "system"
      content = "You are a helpful assistant."
    },
    {
      role    = "user"
      content = "Say 'Provider test successful!' in 5 words or less."
    }
  ]

  max_tokens = 20
}

# ========== Outputs ==========

output "project_info" {
  value = {
    id     = openai_project.test.id
    name   = openai_project.test.name
    status = openai_project.test.status
  }
  description = "Created project information (used admin key)"
}

output "total_projects" {
  value       = length(data.openai_projects.all.projects)
  description = "Total number of projects (used admin key)"
}

output "embedding_dimensions" {
  value       = length(openai_embedding.test.embedding)
  description = "Embedding vector dimensions (used project key)"
}

output "chat_response" {
  value       = data.openai_chat_completion.test.choices[0].message.content
  description = "Chat completion response (used project key)"
}

output "summary" {
  value       = <<-EOT
    ✓ Admin API key successfully used for:
      - Creating project: ${openai_project.test.name}
      - Listing ${length(data.openai_projects.all.projects)} total projects
    
    ✓ Project API key successfully used for:
      - Creating embedding with ${length(openai_embedding.test.embedding)} dimensions
      - Chat completion: ${data.openai_chat_completion.test.choices[0].message.content}
  EOT
  description = "Summary of API key usage"
}