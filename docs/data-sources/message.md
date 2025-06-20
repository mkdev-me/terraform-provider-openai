---
page_title: "OpenAI: openai_message Data Source"
subcategory: ""
description: |-
  Retrieves information about a specific message in an OpenAI thread.
---

# openai_message Data Source

The `openai_message` data source allows you to retrieve detailed information about a specific message within a thread using the OpenAI Assistants API.

## Example Usage

```hcl
# Retrieve a specific message from a thread
data "openai_message" "example" {
  thread_id  = "thread_abc123" 
  message_id = "msg_xyz789"
}

# Output the message content
output "message_content" {
  value = data.openai_message.example.content
}

# Check if the message has attachments
output "has_attachments" {
  value = length(data.openai_message.example.attachments) > 0
}
```

## Argument Reference

* `thread_id` - (Required) The ID of the thread that contains the message.
* `message_id` - (Required) The ID of the message to retrieve.

## Attribute Reference

* `id` - The ID of the message.
* `object` - The object type, always "thread.message".
* `created_at` - The timestamp for when the message was created.
* `thread_id` - The ID of the thread that contains the message.
* `role` - The role of the entity that created the message (e.g., "user" or "assistant").
* `content` - The text content of the message.
* `assistant_id` - If applicable, the ID of the assistant that authored this message.
* `run_id` - If applicable, the ID of the run that generated this message.
* `metadata` - Set of key-value pairs attached to the message.
* `attachments` - A list of attachments in the message. Each attachment contains:
  * `id` - The ID of the attachment.
  * `type` - The type of the attachment.
  * `assistant_id` - If applicable, the ID of the assistant this attachment is associated with.
  * `created_at` - The timestamp for when the attachment was created. 