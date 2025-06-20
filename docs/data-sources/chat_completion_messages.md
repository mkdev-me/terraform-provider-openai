---
page_title: "OpenAI: openai_chat_completion_messages Data Source"
subcategory: ""
description: |-
  Retrieves messages from a specific OpenAI chat completion.
---

# openai_chat_completion_messages Data Source

> **Important Note:** This data source requires the "Chat Completions Store" feature to be enabled on your OpenAI account. This is a relatively new and experimental feature that is not available by default for all accounts. Additionally, when creating chat completions, you must use a compatible model (e.g., gpt-4o) and set the `store` parameter to true.

The `openai_chat_completion_messages` data source allows you to retrieve the messages associated with a specific chat completion. This is useful for examining the conversation context and history of a chat completion, including both user inputs and model responses.

## Example Usage

```hcl
data "openai_chat_completion_messages" "example" {
  completion_id = "chat_abc123"
  limit = 50
  order = "desc"
}

output "chat_history" {
  value = data.openai_chat_completion_messages.example.messages
}

# Extract the last user message
output "last_user_message" {
  value = [
    for msg in data.openai_chat_completion_messages.example.messages :
    msg.content if msg.role == "user"
  ][0]
}
```

## Argument Reference

* `completion_id` - (Required) The ID of the chat completion to retrieve messages from (format: chat_xxx).
* `api_key` - (Optional) Custom API key to use for this data source. If not provided, the provider's default API key will be used.
* `after` - (Optional) Identifier for the last message from the previous pagination request.
* `limit` - (Optional) Number of messages to retrieve. Defaults to 20, maximum is 100.
* `order` - (Optional) Sort order for messages by timestamp. Use "asc" for ascending order or "desc" for descending order. Defaults to "asc".

## Attribute Reference

* `id` - A unique identifier for this data source (not the completion ID).
* `messages` - The list of messages from the chat completion. Each message contains:
  * `role` - The role of the message author (e.g., "system", "user", "assistant", or "function").
  * `content` - The content of the message.
  * `name` - The name of the author of this message, if provided.
  * `function_call` - The function call details, if present:
    * `name` - The name of the function to call.
    * `arguments` - The arguments to call the function with, as a JSON string.
* `has_more` - Whether there are more messages to retrieve.
* `first_id` - The ID of the first message in the response.
* `last_id` - The ID of the last message in the response. 