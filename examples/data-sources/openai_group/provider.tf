terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

provider "openai" {
  # Admin API key is required for group data sources
  # Set via OPENAI_ADMIN_KEY environment variable or admin_key attribute
}
