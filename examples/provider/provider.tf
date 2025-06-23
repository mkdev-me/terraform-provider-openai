# Basic provider configuration
provider "openai" {
  # Admin key is loaded from OPENAI_ADMIN_KEY environment variable
  # API key is loaded from OPENAI_API_KEY environment variable
}

# Provider with organization ID
provider "openai" {
  # Admin key is loaded from OPENAI_ADMIN_KEY environment variable
  # API key is loaded from OPENAI_API_KEY environment variable
  alias           = "org"
  api_key         = var.openai_api_key
  organization_id = var.organization_id
}

# Provider with admin API key for organization management
provider "openai" {
  # Admin key is loaded from OPENAI_ADMIN_KEY environment variable
  # API key is loaded from OPENAI_API_KEY environment variable
  alias = "admin"
}

