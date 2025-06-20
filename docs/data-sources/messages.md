---
page_title: "OpenAI: openai_messages Data Source"
subcategory: ""
description: |-
  Lists messages in an OpenAI thread with pagination support.
---

# openai_messages Data Source

The `openai_messages` data source allows you to retrieve a list of messages from a specific thread using the OpenAI Assistants API. It includes pagination support so you can control how many messages are returned and in what order.

## Example Usage

```hcl
# List all messages in a thread
data "openai_messages" "all" {
  thread_id = "thread_abc123"
  limit     = 20
  order     = "desc"  # Most recent first
}

# Output the messages
output "thread_messages" {
  value = data.openai_messages.all.messages
}

# Get messages before a specific message
data "openai_messages" "older_messages" {
  thread_id = "thread_abc123"
  limit     = 10
  order     = "desc"
  before    = "msg_xyz789"  # Get messages before this one
}
```

## Argument Reference

* `thread_id` - (Required) The ID of the thread containing the messages.
* `limit` - (Optional) A limit on the number of messages to be returned. Default is 20, maximum is 100.
* `order` - (Optional) Sort order by creation timestamp. One of: `asc` (oldest first) or `desc` (newest first). Default is `desc`.
* `after` - (Optional) A cursor for pagination. Returns objects after this message ID.
* `before` - (Optional) A cursor for pagination. Returns objects before this message ID.

## Attribute Reference

* `messages` - The list of messages in the thread. Each message contains:
  * `id` - The identifier of this message.
  * `object` - The object type, always "thread.message".
  * `created_at` - The timestamp for when the message was created.
  * `thread_id` - The ID of the thread that contains the message.
  * `role` - The role of the entity that created the message.
  * `content` - The text content of the message.
  * `assistant_id` - If applicable, the ID of the assistant that authored this message.
  * `run_id` - If applicable, the ID of the run that generated this message.
  * `metadata` - Set of key-value pairs attached to the message.
  * `attachments` - A list of attachments in the message.
* `has_more` - Whether there are more items available in the list.
* `first_id` - The ID of the first message in the list.
* `last_id` - The ID of the last message in the list. 