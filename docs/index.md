---
page_title: "OpenAI Terraform Provider"
description: |-
  The OpenAI provider enables infrastructure-as-code management of OpenAI resources.
---

# OpenAI Terraform Provider

The OpenAI Terraform Provider enables you to manage OpenAI resources as infrastructure-as-code, providing a declarative way to work with the OpenAI API.

## Features

- **Complete API Coverage**: Support for chat completions, assistants, fine-tuning, embeddings, images, audio, and more
- **Organization Management**: Manage projects, users, API keys, and rate limits
- **Resource Import**: Import existing OpenAI resources into Terraform state
- **Comprehensive Documentation**: Detailed guides for all resources and data sources

## Getting Started

See the [provider documentation](https://registry.terraform.io/providers/openai/openai/latest/docs) for detailed information on:

- [Installation and Configuration](https://registry.terraform.io/providers/openai/openai/latest/docs#schema)
- [Authentication](https://registry.terraform.io/providers/openai/openai/latest/docs#authentication)
- [Resources](https://registry.terraform.io/providers/openai/openai/latest/docs/resources)
- [Data Sources](https://registry.terraform.io/providers/openai/openai/latest/docs/data-sources)

## Example Usage

```hcl
terraform {
  required_providers {
    openai = {
      source  = "openai/openai"
      version = "~> 1.0"
    }
  }
}

provider "openai" {
  # Authentication via environment variables:
  # OPENAI_API_KEY and OPENAI_ORGANIZATION_ID
}

# Create a chat completion
resource "openai_chat_completion" "example" {
  model = "gpt-4"
  
  messages {
    role    = "system"
    content = "You are a helpful assistant."
  }
  
  messages {
    role    = "user"
    content = "Hello!"
  }
}

output "response" {
  value = openai_chat_completion.example.choices[0].message.content
}
```

## Resources

See the sidebar for a complete list of available resources.

## Data Sources

See the sidebar for a complete list of available data sources.
