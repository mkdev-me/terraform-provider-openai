# openai_vector_store_file_batch Resource

Manages a batch of files within an OpenAI Vector Store. This resource allows adding multiple files to a vector store at once, which is more efficient than adding files individually.

## Example Usage

```hcl
resource "openai_file" "document1" {
  file    = "path/to/first_document.pdf"
  purpose = "assistants"
}

resource "openai_file" "document2" {
  file    = "path/to/second_document.pdf"
  purpose = "assistants"
}

resource "openai_vector_store" "knowledge_base" {
  name = "Knowledge Base"
}

resource "openai_vector_store_file_batch" "batch_upload" {
  vector_store_id = openai_vector_store.knowledge_base.id
  
  file_ids = [
    openai_file.document1.id,
    openai_file.document2.id
  ]
  
  # Optional: Add attributes for the entire batch
  attributes = {
    "batch_type" = "documentation",
    "department" = "engineering",
    "version"    = "2.0"
  }
  
  # Optional: Set chunking strategy for all files in the batch
  chunking_strategy {
    type = "auto"
  }
}
```

## Argument Reference

* `vector_store_id` - (Required) The ID of the vector store to add the files to.
* `file_ids` - (Optional/Computed) The list of file IDs to add to the vector store.
* `attributes` - (Optional) Dynamic description or tags for the files in the vector store.
* `chunking_strategy` - (Optional) The chunking strategy used to chunk the files.
  * `type` - (Required) The type of chunking strategy (auto, fixed, or semantic).
  * `size` - (Optional) The size in characters for fixed chunking strategy.
  * `max_tokens` - (Optional) The maximum tokens per chunk for semantic chunking strategy.

## Attributes Reference

* `id` - The ID of the vector store file batch operation.
* `created_at` - The timestamp for when the files were added to the vector store.
* `object` - The object type (always 'vector_store.file.batch').
* `status` - The current status of the file batch operation.

## Import

Vector store file batches can be imported using the format `vector_store_id/batch_id`:

```bash
terraform import openai_vector_store_file_batch.example vs_abc123def456/vsfb_xyz789
```

## Resource Behavior Notes

Due to API limitations:
1. This resource does not support updates to existing batches - any changes will cause a new batch to be created
2. This resource does not support deletion of batches via API - the delete operation is a no-op that simply removes the resource from Terraform state 