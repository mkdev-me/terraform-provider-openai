---
page_title: "OpenAI: openai_message Resource"
subcategory: ""
description: |-
  Manages messages within OpenAI assistant threads.
---

# openai_message Resource

The `openai_message` resource creates and manages individual messages within threads in OpenAI's Assistant API. Messages are the content exchanged between users and assistants, and can include text and file attachments.

## Example Usage

```hcl
# Create a thread
resource "openai_thread" "example" {
  # An empty thread will be created
}

# Add a text message to the thread
resource "openai_message" "user_query" {
  thread_id = openai_thread.example.id
  role      = "user"
  content   = "Can you analyze the global temperature trends over the past 50 years?"
}

# Add a message with file attachments
resource "openai_file" "data_file" {
  file    = "/path/to/temperature_data.csv"
  purpose = "assistants"
}

resource "openai_message" "with_data" {
  thread_id = openai_thread.example.id
  role      = "user"
  content   = "Here's the temperature dataset I'd like you to analyze."
  
  attachments {
    file_id = openai_file.data_file.id
    tools   = ["retrieval"]
  }
  
  metadata = {
    source      = "noaa",
    data_format = "csv"
  }
}
```

## Argument Reference

* `thread_id` - (Required) The ID of the thread to add this message to.
* `role` - (Required) The role of the message creator. Currently only "user" is supported for creating messages.
* `content` - (Required) The text content of the message.
* `attachments` - (Optional) A list of file attachments to include with the message. Each attachment block supports:
  * `file_id` - (Required) The ID of the file to attach.
  * `tools` - (Required) A list of tools that should use this file, e.g., `["retrieval"]`.
* `metadata` - (Optional) A map of metadata that can help you categorize or organize the message. Maximum of 16 key-value pairs.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the message.
* `object` - The object type, always "thread.message".
* `created_at` - The timestamp when the message was created.
* `assistant_id` - The ID of the assistant that authored this message (for assistant messages only).
* `run_id` - The ID of the run that generated this message (for assistant messages only).

## Import

Messages can be imported using the format `{thread_id}:{message_id}`, e.g.,

```bash
terraform import openai_message.example thread_abc123:msg_xyz456
``` 