---
page_title: "openai_assistant Data Source - OpenAI Terraform Provider"
subcategory: ""
description: |-
  Retrieves information about a specific OpenAI Assistant by ID.
---

# openai_assistant Data Source

This data source allows you to retrieve detailed information about a specific OpenAI Assistant using its ID.

## Example Usage

```terraform
# Retrieve information about an existing assistant
data "openai_assistant" "my_assistant" {
  assistant_id = "asst_abc123xyz456"
}

# Use the data in outputs or other resources
output "assistant_name" {
  value = data.openai_assistant.my_assistant.name
}

output "assistant_model" {
  value = data.openai_assistant.my_assistant.model
}

output "assistant_tools" {
  value = data.openai_assistant.my_assistant.tools
}
```

## Argument Reference

* `assistant_id` - (Required) The ID of the assistant to retrieve.

## Attributes Reference

In addition to the argument above, the following attributes are exported:

* `id` - The ID of the assistant.
* `object` - The object type, which is always "assistant".
* `created_at` - The timestamp for when the assistant was created.
* `name` - The name of the assistant.
* `description` - The description of the assistant.
* `model` - The model used by the assistant.
* `instructions` - The system instructions of the assistant.
* `tools` - The list of tools enabled on the assistant. Each tool has the following structure:
  * `type` - The type of tool, which can be "code_interpreter", "retrieval", "function", or "file_search".
  * `function` - If the tool type is "function", this contains details about the function:
    * `name` - The name of the function.
    * `description` - The description of the function.
    * `parameters` - The parameters of the function in JSON format.
* `file_ids` - The list of file IDs attached to the assistant.
* `metadata` - Metadata attached to the assistant.
* `response_format` - The format of responses from the assistant.
* `reasoning_effort` - Constrains the effort spent on reasoning for reasoning models (low, medium, or high).
* `temperature` - The sampling temperature to use with this assistant (0-2).
* `top_p` - The nucleus sampling parameter to use with this assistant (0-1).

## Import

The assistant data source does not support import, as it's a read-only data source. 