terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

provider "openai" {
  # Admin key is loaded from OPENAI_ADMIN_KEY environment variable
}

