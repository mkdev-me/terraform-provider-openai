terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

provider "openai" {
  # The admin_key is required for reading project roles
  # Set via OPENAI_ADMIN_KEY environment variable or provider configuration
}
