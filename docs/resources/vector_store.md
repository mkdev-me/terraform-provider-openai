# openai_vector_store Resource

Manages an OpenAI Vector Store resource for storing, organizing, and searching vector embeddings of files.

## Example Usage

```hcl
resource "openai_vector_store" "example" {
  name = "Knowledge Base"
  
  # Optional: Add files directly to the vector store
  file_ids = [
    openai_file.document1.id,
    openai_file.document2.id
  ]
  
  # Optional: Add metadata for organization
  metadata = {
    "department" = "support",
    "version"    = "1.0"
  }
  
  # Optional: Set chunking strategy
  chunking_strategy {
    type = "auto"
  }
  
  # Optional: Set expiration policy
  expires_after {
    days   = 90
    anchor = "last_active_at"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the vector store.
* `file_ids` - (Optional) A list of File IDs to add to the vector store.
* `metadata` - (Optional) Set of key-value pairs that can be attached to the vector store.
* `chunking_strategy` - (Optional) The chunking strategy used to chunk the files.
  * `type` - (Required) The type of chunking strategy. Valid values are "auto" or "static".
* `expires_after` - (Optional) The expiration policy for the vector store.
  * `days` - (Required) Number of days after which the vector store should expire. Must be 1 or greater.
  * `anchor` - (Required) The anchor time for the expiration. Currently only supports "last_active_at".

## Attributes Reference

* `id` - The ID of the vector store.
* `name` - The name of the vector store.
* `created_at` - The timestamp for when the vector store was created.
* `file_count` - The number of files in the vector store.
* `object` - The object type (always 'vector_store').
* `status` - The current status of the vector store.

## Import

Vector stores can be imported using the vector store ID:

```bash
terraform import openai_vector_store.example vs_abc123def456
``` 