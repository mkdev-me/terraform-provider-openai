terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

provider "openai" {}

# Example 1: Simple Vector Store without files
module "basic_vector_store" {
  source = "../../modules/vector_store"

  name = "Basic Knowledge Base"

  # Add some metadata for organization
  metadata = {
    "category" = "general",
    "purpose"  = "demonstration",
    "version"  = "1.0"
  }

}

# Example 2: Vector Store with individually added files
module "support_vector_store" {
  source = "../../modules/vector_store"

  name     = "Support Knowledge Base"
  file_ids = []

  use_file_batches = true

  file_attributes = {
    "department" = "support",
    "language"   = "english"
  }

}

# Example 3: Vector Store with file batches
module "api_vector_store" {
  source = "../../modules/vector_store"

  name     = "API Documentation Store"
  file_ids = []

  use_file_batches = true

}

# Example 4: Create a resource directly (not through the module)
resource "openai_vector_store" "custom_store" {
  name = "Custom Vector Store"

  # Start without files, add them later
  file_ids = []

  metadata = {
    "project" = "terraform-demo",
    "owner"   = "infrastructure-team"
  }

}
