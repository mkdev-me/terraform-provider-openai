---
page_title: "OpenAI: openai_assistants Data Source"
subcategory: ""
description: |-
  Retrieves a list of assistants from OpenAI.
---

# openai_assistants Data Source

Retrieves a list of assistants from OpenAI. This data source allows you to get information about assistants that have been created in your account, including their configuration, capabilities, and metadata.

## Example Usage

```hcl
data "openai_assistants" "all" {
  # Optional filters
  limit = 10
  order = "desc"
  
  # Optional: Filter by a specific ID after which to retrieve assistants
  after = "asst_abc123"
  
  # Optional: Filter assistants by those created before a specific timestamp
  before = "asst_xyz789"
}

# Access assistant details
output "assistant_count" {
  value = length(data.openai_assistants.all.assistants)
}

output "assistant_names" {
  value = [for a in data.openai_assistants.all.assistants : a.name]
}

# Find a specific assistant by name
locals {
  my_assistant = [
    for a in data.openai_assistants.all.assistants : a
    if a.name == "My AI Assistant"
  ][0]
}

output "my_assistant_id" {
  value = local.my_assistant.id
}
```

## Argument Reference

* `limit` - (Optional) The maximum number of assistants to retrieve. Defaults to 20 if not specified.
* `order` - (Optional) The sort order for the assistants. Can be "asc" or "desc". Defaults to "desc".
* `after` - (Optional) A cursor for pagination. Only retrieve assistants created after this ID.
* `before` - (Optional) A cursor for pagination. Only retrieve assistants created before this ID.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A unique identifier for this data source.
* `has_more` - Indicates if there are more assistants available beyond the current response.
* `first_id` - The ID of the first assistant in the returned list.
* `last_id` - The ID of the last assistant in the returned list.
* `assistants` - A list of assistants. Each assistant has the following attributes:
  * `id` - The ID of the assistant.
  * `name` - The name of the assistant.
  * `description` - The description of the assistant.
  * `model` - The model used by the assistant.
  * `instructions` - The instructions that define the assistant's behavior.
  * `tools` - A list of tools enabled for the assistant, such as code_interpreter, retrieval, or function calling.
  * `tool_resources` - Additional resources configured for tools.
  * `file_ids` - A list of file IDs attached to the assistant.
  * `metadata` - A map of metadata associated with the assistant.
  * `created_at` - The timestamp (in Unix time) when the assistant was created.

## Permission Requirements

To use this data source, your API key must have permission to list assistants in your account.

## Import

Assistants data sources cannot be imported. 