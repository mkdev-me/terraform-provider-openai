---
page_title: "OpenAI: openai_assistant Resource"
subcategory: ""
description: |-
  Manages OpenAI assistants that can generate responses, call tools, and handle conversations.
---

# openai_assistant Resource

The `openai_assistant` resource allows you to create and manage assistants in OpenAI's Assistant API. Assistants are specialized AI entities that can be customized with specific instructions, tools, and knowledge to handle various tasks through conversations.

## Example Usage

```hcl
# Create a simple assistant
resource "openai_assistant" "basic" {
  name         = "Customer Support Assistant"
  description  = "Assists users with product questions and issues"
  model        = "gpt-4"
  instructions = "You are a helpful customer support agent. Be friendly and concise."
}

# Create an assistant with tools and files
resource "openai_file" "knowledge_base" {
  content_type = "application/pdf"
  filename     = "/path/to/product_manual.pdf"
  purpose      = "assistants"
}

resource "openai_assistant" "advanced" {
  name         = "Research Assistant"
  model        = "gpt-4-turbo"
  instructions = "You are a research assistant that helps find information."
  
  tool {
    type = "code_interpreter"
  }
  
  tool {
    type = "retrieval"
  }
  
  file_ids = [openai_file.knowledge_base.id]
  
  metadata = {
    environment = "production"
    version     = "1.0"
  }
}
```

## Argument Reference

* `name` - (Optional) The name of the assistant. Maximum of 256 characters.
* `description` - (Optional) The description of the assistant. Maximum of 512 characters.
* `model` - (Required) The ID of the model to use (e.g., "gpt-4", "gpt-3.5-turbo").
* `instructions` - (Optional) Instructions that guide the assistant's behavior.
* `tool` - (Optional) One or more tool blocks representing tools the assistant may use:
  * `type` - (Required) The type of tool ("code_interpreter", "retrieval", or "function").
  * `function` - (Required when type is "function") A function block:
    * `name` - (Required) The name of the function.
    * `description` - (Optional) A description of what the function does.
    * `parameters` - (Required) The parameters as a JSON Schema object.
* `file_ids` - (Optional) A list of file IDs that the assistant can access.
* `metadata` - (Optional) A map of metadata for the assistant.
* `api_key` - (Optional) Custom API key to use for this resource.

## Attribute Reference

* `id` - The ID of the assistant.
* `object` - The object type, always "assistant".
* `created_at` - The timestamp when the assistant was created.

## Import

Assistants can be imported using the assistant ID:

```bash
terraform import openai_assistant.example asst_abc123
``` 