# OpenAI Messages Module

This module provides a simplified interface for creating and retrieving messages in OpenAI threads. It supports both creating new messages and retrieving existing ones, with support for file attachments and metadata.

## Features

- Create messages within threads
- Attach files to messages
- Add metadata to messages
- Support for future data source functionality
- Comprehensive output attributes for accessing message details

## Usage

### Creating a New Message

```hcl
module "simple_message" {
  source = "path/to/modules/messages"

  thread_id = openai_thread.example.id
  content   = "This is a simple text message."
}

module "message_with_attachment" {
  source = "path/to/modules/messages"

  thread_id = openai_thread.example.id
  content   = "This message includes a file attachment."
  
  attachments = [
    {
      file_id = openai_file.document.id
      tools   = ["retrieval"]
    }
  ]
  
  metadata = {
    type    = "document_analysis"
    source  = "terraform_module"
  }
}
```

### Retrieving an Existing Message

```hcl
module "retrieve_message" {
  source = "path/to/modules/messages"

  use_data_source     = true
  thread_id           = "thread_abc123"
  existing_message_id = "msg_xyz789"
}

output "message_content" {
  value = module.retrieve_message.content
}

output "message_attachments" {
  value = module.retrieve_message.attachments
}
```

## Input Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| `thread_id` | The ID of the thread to add the message to | `string` | n/a | yes |
| `role` | The role of the entity creating the message | `string` | `"user"` | no |
| `content` | The content of the message | `string` | `null` | no |
| `attachments` | List of file attachments to include with the message | `list(object({ file_id = string, tools = list(string) }))` | `[]` | no |
| `metadata` | Set of key-value pairs that can be attached to the message | `map(string)` | `{}` | no |
| `use_data_source` | Whether to use the data source to retrieve an existing message | `bool` | `false` | no |
| `existing_message_id` | ID of an existing message to retrieve (when use_data_source is true) | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| `message_id` | The ID of the message |
| `created_at` | The timestamp for when the message was created |
| `role` | The role of the entity that created the message |
| `content` | The content of the message |
| `metadata` | Set of key-value pairs attached to the message |
| `assistant_id` | If applicable, the ID of the assistant that authored this message |
| `run_id` | If applicable, the ID of the run associated with this message |
| `attachments` | A list of attachments in the message |

## Limitations and Notes

- Currently, only the `user` role is supported for message creation (OpenAI API limitation)
- The data source functionality for retrieving existing messages is commented out as it depends on the implementation of `data_source_openai_message.go`

## Dependencies

This module relies on the `openai_message` resource from the `mkdev-me/openai` provider.

## License

This module is licensed under the same terms as the main provider. 