---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "openai_thread Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  
---

# openai_thread (Data Source)



## Example Usage

```terraform
# Fetch a specific thread by ID
data "openai_thread" "customer_support" {
  thread_id = "thread_abc123"
}

# Output thread creation timestamp
output "thread_created_at" {
  value = data.openai_thread.customer_support.created_at
}

# Use thread data in a message creation
resource "openai_thread_message" "follow_up" {
  thread_id = data.openai_thread.customer_support.id
  role      = "assistant"
  content   = "Following up on your support request..."
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `thread_id` (String) The ID of the thread to retrieve

### Read-Only

- `created_at` (Number) The timestamp for when the thread was created
- `id` (String) The ID of this resource.
- `metadata` (Map of String) Metadata attached to the thread
- `object` (String) The object type, which is always 'thread'
