terraform {
  required_version = ">= 1.0"

  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = ">= 1.0.0"
    }
  }
}

