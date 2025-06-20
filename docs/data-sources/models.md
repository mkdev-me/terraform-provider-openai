---
page_title: "OpenAI: openai_models Data Source"
subcategory: ""
description: |-
  Retrieves a list of available models from OpenAI.
---

# openai_models Data Source

Retrieves a list of available models from OpenAI. This data source allows you to get information about all models available to your API key, including both GPT models and specialized models.

## Example Usage

```hcl
data "openai_models" "available" {}

# Access all model IDs
output "available_model_ids" {
  value = data.openai_models.available.model_ids
}

# Check if a specific model is available
output "has_gpt4" {
  value = contains(data.openai_models.available.model_ids, "gpt-4")
}
```

## Argument Reference

This data source doesn't require any configuration arguments.

## Attribute Reference

The following attributes are exported:

* `id` - A unique identifier for this data source.
* `model_ids` - A list of all available model IDs.

## API Key Configuration

You can provide API keys in the following ways:

1. **Provider-level API key** - Set via the provider block or `OPENAI_API_KEY` environment variable.

If the environment variable approach isn't working, explicitly set the key in your provider configuration:

```hcl
provider "openai" {
  api_key = var.openai_api_key
}
```

Then define the variable and provide it via a tfvars file or command line argument:

```hcl
variable "openai_api_key" {
  description = "OpenAI API Key"
  type        = string
  sensitive   = true
}
```

## Permission Requirements

To use this data source, your API key must have permission to list available models.

## Import

Models data sources cannot be imported. 