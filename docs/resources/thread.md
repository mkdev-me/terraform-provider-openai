---
page_title: "OpenAI: openai_thread Resource"
subcategory: ""
description: |-
  Manages conversation threads for OpenAI assistants.
---

# openai_thread Resource

The `openai_thread` resource creates and manages conversation threads for use with OpenAI's Assistant API. Threads maintain the history of conversations between users and assistants, allowing for context-aware interactions.

## Example Usage

```hcl
# Create a simple empty thread
resource "openai_thread" "example" {
  # An empty thread will be created
}

# Create a thread with initial messages
resource "openai_thread" "with_messages" {
  message {
    role     = "user"
    content  = "Hello, I need help with a research project on climate change."
    file_ids = [] # Optional file attachments
  }
  
  message {
    role    = "user"
    content = "I'm specifically looking at rising sea levels."
  }
  
  metadata = {
    conversation_type = "research",
    user_id           = "u_123456"
  }
}
```

## Argument Reference

* `message` - (Optional) One or more message blocks to add as initial messages in the thread. Each block supports:
  * `role` - (Required) The role of the message creator. Currently only "user" is supported for thread creation.
  * `content` - (Required) The content of the message.
  * `file_ids` - (Optional) A list of file IDs that are part of this message. The files must have been uploaded with purpose "assistants".
* `metadata` - (Optional) A map of metadata that can help you categorize or organize the thread. Maximum of 16 key-value pairs.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the thread.
* `object` - The object type, always "thread".
* `created_at` - The timestamp when the thread was created.

## Import

Threads can be imported using the thread ID, e.g.,

```bash
terraform import openai_thread.example thread_abc123
``` 