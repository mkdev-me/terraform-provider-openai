terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

provider "openai" {
  # API key is loaded from OPENAI_API_KEY environment variable
}