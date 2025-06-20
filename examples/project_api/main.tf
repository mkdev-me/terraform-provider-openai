# Example: OpenAI Project API Keys
# This example demonstrates how to retrieve information about existing OpenAI project API keys in Terraform

# ===================================================================
# USAGE INSTRUCTIONS
# ===================================================================
# This example provides ways to work with OpenAI Project API keys:
#
# 1. RETRIEVE EXISTING API KEYS:
#    - With api_key_id: Retrieve a specific API key
#    - With retrieve_all = true: Retrieve all keys for a project
#
# 2. IMPORT EXISTING API KEYS (using the script):
#    - Use the provided import script:
#      ./import_project_key.sh <project_id> <key_name>
#
# Important Note:
# - OpenAI DOES NOT support programmatically creating project API keys
# - You must create project API keys manually in the OpenAI dashboard first
# - See the README.md for detailed instructions
#
# Script Features:
# - Automatically looks up the API key ID if you provide only the name
# - Validates inputs and displays helpful error messages
# - Updates this Terraform file with the correct values
# - Requires OPENAI_ADMIN_KEY environment variable to be set
# ===================================================================

# Define Terraform configuration and required providers
terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai" # Custom OpenAI provider source
      version = "~> 1.0.0"      # Provider version constraint
    }
  }
}

# Input Variables
# ------------------------------
# Admin API key for authenticating with OpenAI

# Optional: ID of an existing API key to look up
variable "existing_api_key_id" {
  description = "ID of an existing API key to look up"
  type        = string
  default     = "" # Default empty string means no specific key is requested
}

# Provider Configuration
# ------------------------------
# Configure the OpenAI Provider with the admin API key
provider "openai" {
  # API keys are automatically loaded from environment variables:
  # - OPENAI_API_KEY for project operations
  # - OPENAI_ADMIN_KEY for admin operations
}

# Resources
# ------------------------------
# Create a demo project to work with
resource "openai_project" "example" {
  name = "Project API Key Example" # Name of the example project
}

# Data Sources
# ------------------------------
# 1. RETRIEVE A SPECIFIC EXISTING API KEY (when ID is provided)
# This data source retrieves a single API key by its ID
# Only created when existing_api_key_id is provided (using count conditional)
data "openai_project_api_key" "specific_key" {
  count      = var.existing_api_key_id != "" ? 1 : 0 # Conditional creation
  project_id = openai_project.example.id             # ID of the project
  api_key_id = var.existing_api_key_id               # ID of the specific API key to retrieve
}

# 2. RETRIEVE ALL API KEYS FOR THE PROJECT
# This data source retrieves all API keys for the specified project
data "openai_project_api_keys" "all_keys" {
  project_id = openai_project.example.id # ID of the project
}

# Outputs
# ------------------------------
# Output the project ID for reference
output "project_id" {
  value = openai_project.example.id # ID of the created project
}

# Output information about a specific API key (if requested)
output "existing_api_key_name" {
  value = var.existing_api_key_id != "" ? data.openai_project_api_key.specific_key[0].name : "No specific key ID provided"
  # Uses conditional to either output the key name or a message if no key ID was provided
}

output "existing_api_key_created_at" {
  value = var.existing_api_key_id != "" ? data.openai_project_api_key.specific_key[0].created_at : "No specific key ID provided"
  # Uses conditional to either output the key creation date or a message if no key ID was provided
}

# Output summary information about all API keys for the project
output "project_api_key_count" {
  value = length(data.openai_project_api_keys.all_keys.api_keys) # Count of API keys
}

# Output detailed information about all API keys, formatted as a list of objects
output "project_api_keys" {
  value = [for key in data.openai_project_api_keys.all_keys.api_keys : {
    id         = key.id         # API key ID
    name       = key.name       # API key name
    created_at = key.created_at # API key creation timestamp
  }]
}
