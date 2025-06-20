# OpenAI Threads Example

This example demonstrates how to use the OpenAI Terraform Provider to create, retrieve, and import OpenAI Threads.

## What are OpenAI Threads?

Threads are conversation sessions in the OpenAI Assistants API. They maintain the history of messages between a user and an assistant, providing context for ongoing conversations. Threads can be:

- Created empty (to add messages later)
- Created with initial messages
- Annotated with metadata for organization and retrieval

## Example Components

This example demonstrates:

1. Creating an empty thread with no initial configuration
2. Creating a thread with initial user messages
3. Creating a thread with only metadata
4. Using the thread data source to retrieve thread details
5. Importing an existing thread from OpenAI into Terraform
6. Outputting thread properties and attributes

## Usage

To run this example:

1. Set your OpenAI API key:
```bash
export OPENAI_API_KEY="your-api-key"
```

2. Initialize Terraform:
```bash
terraform init
```

3. Apply the configuration:
```bash
terraform apply
```

## Terraform Resources Used

| Resource/Data Source | Description |
|----------------------|-------------|
| `openai_thread` | Create and manage OpenAI threads |
| `openai_thread` (data source) | Retrieve information about a specific thread |

## Thread Properties

The `openai_thread` resource supports the following properties:

- `messages` - Initial messages to include in the thread (optional)
  - `role` - The role of the entity creating the message (currently only "user" is supported)
  - `content` - The content of the message
  - `file_ids` - List of file IDs to attach to the message (optional)
  - `metadata` - Key-value pairs for message organization (optional)
- `metadata` - Key-value pairs for thread organization (optional)

## Advanced Usage

### Adding Messages to Threads

After creating a thread, you typically need to add messages to it to have a conversation. This can be done using the `openai_message` resource, which is not covered in this example.

### Using Threads with Assistants

Threads are most useful when combined with assistants and runs. See the assistants example for details on how to use threads with assistants.

## Importing Threads

If you have threads that were created outside of Terraform (e.g., through the OpenAI API or dashboard), you can import them into Terraform management:

1. Create a placeholder resource in your Terraform configuration:
```hcl
resource "openai_thread" "imported_thread" {
  # No configuration needed for import
}
```

2. Run the import command:
```bash
terraform import openai_thread.imported_thread thread_abc123
```

3. After importing, you can view the imported state and modify metadata as needed:
```bash
terraform state show openai_thread.imported_thread
```

## Output Example

After applying this configuration, you'll get outputs similar to:

```
empty_thread_id = "thread_abc123..."
with_messages_thread_id = "thread_def456..."
with_metadata_thread_id = "thread_ghi789..."
data_thread_id = "thread_def456..."
data_thread_created_at = 1699125678
data_thread_metadata = {
  "created_by" = "terraform",
  "priority" = "medium",
  "subject" = "quantum_computing"
}
``` 