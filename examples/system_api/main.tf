# OpenAI Admin API Key Management Example
# This example demonstrates creating and retrieving admin API keys

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }
}

# Configure the OpenAI Provider
provider "openai" {
  # API keys are automatically loaded from environment variables:
  # - OPENAI_API_KEY for project operations
  # - OPENAI_ADMIN_KEY for admin operations
}

#------------------------------------------------------------------------------
# Create an admin API key
#------------------------------------------------------------------------------
resource "openai_admin_api_key" "test_key" {
  name       = "test-direct-admin-key"
  expires_at = 1735689600 # Unix timestamp (Dec 31, 2024)
  scopes     = ["api.management.read"]
}

# Save the API key to a file directly using the local_file resource
# This captures the key value directly at creation time before it's redacted
resource "local_file" "api_key_file" {
  content         = openai_admin_api_key.test_key.api_key_value
  filename        = "${path.module}/admin_api_key.txt"
  file_permission = "0600" # Restricted permissions for security
}

#------------------------------------------------------------------------------
# DATA SOURCES - Retrieving Admin API Keys
#------------------------------------------------------------------------------
# Wait a few seconds to ensure the API key is available for retrieval
resource "time_sleep" "wait_for_api_key" {
  depends_on      = [openai_admin_api_key.test_key]
  create_duration = "5s"
}

# Retrieve the specific admin API key we just created by ID
data "openai_admin_api_key" "test_key_data" {
  depends_on = [time_sleep.wait_for_api_key]
  api_key_id = openai_admin_api_key.test_key.id
}

# List all admin API keys
data "openai_admin_api_keys" "all_keys" {
  depends_on = [time_sleep.wait_for_api_key]
  limit      = 10 # Limit the number of keys returned
}

#------------------------------------------------------------------------------
# Outputs
#------------------------------------------------------------------------------
# Output created API key details
output "created_admin_key" {
  description = "Admin API key details (created resource)"
  value = {
    id         = openai_admin_api_key.test_key.id
    name       = openai_admin_api_key.test_key.name
    created_at = openai_admin_api_key.test_key.created_at
    expires_at = openai_admin_api_key.test_key.expires_at
  }
}

# Output API key value (sensitive)
output "created_admin_key_value" {
  description = "Admin API key value (available only after creation)"
  value       = openai_admin_api_key.test_key.api_key_value
  sensitive   = true
}

# Output retrieved API key details
output "retrieved_admin_key" {
  description = "Admin API key details (retrieved via data source)"
  value = {
    id         = data.openai_admin_api_key.test_key_data.id
    name       = data.openai_admin_api_key.test_key_data.name
    created_at = data.openai_admin_api_key.test_key_data.created_at
    scopes     = data.openai_admin_api_key.test_key_data.scopes
  }
}

# Output count of all API keys
output "all_admin_keys_count" {
  description = "Number of admin API keys retrieved"
  value       = length(data.openai_admin_api_keys.all_keys.api_keys)
}

# Output list of all API keys (names only for brevity)
output "all_admin_key_names" {
  description = "Names of all admin API keys"
  value       = [for key in data.openai_admin_api_keys.all_keys.api_keys : key.name]
}

# Output file path
output "api_key_file_path" {
  description = "Path to the file containing the API key"
  value       = local_file.api_key_file.filename
}
