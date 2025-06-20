---
page_title: "openai_thread Data Source - OpenAI Terraform Provider"
subcategory: ""
description: |-
  Retrieves information about a specific OpenAI Thread by ID.
---

# openai_thread Data Source

This data source allows you to retrieve detailed information about a specific OpenAI Thread using its ID.

## Example Usage

```terraform
# Retrieve information about an existing thread
data "openai_thread" "my_thread" {
  thread_id = "thread_abc123xyz456"
}

# Use the data in outputs or other resources
output "thread_created_at" {
  value = data.openai_thread.my_thread.created_at
}

output "thread_metadata" {
  value = data.openai_thread.my_thread.metadata
}
```

## Argument Reference

* `thread_id` - (Required) The ID of the thread to retrieve.

## Attributes Reference

In addition to the argument above, the following attributes are exported:

* `id` - The ID of the thread.
* `object` - The object type, which is always "thread".
* `created_at` - The timestamp for when the thread was created.
* `metadata` - Metadata attached to the thread.

## Notes

This data source only retrieves information about the thread itself, not the messages within the thread. To retrieve messages, use the OpenAI API directly or explore the `openai_message` data source.

## Import

The thread data source does not support import, as it's a read-only data source. 