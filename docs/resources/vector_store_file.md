# openai_vector_store_file Resource

Manages a file within an OpenAI Vector Store.

## Example Usage

```hcl
resource "openai_file" "document" {
  file    = "path/to/local/document.pdf"
  purpose = "assistants"
}

resource "openai_vector_store" "knowledge_base" {
  name = "Knowledge Base"
}

resource "openai_vector_store_file" "document_in_store" {
  vector_store_id = openai_vector_store.knowledge_base.id
  file_id         = openai_file.document.id
  
  # Optional: Add attributes for organization and filtering
  attributes = {
    "category" = "documentation",
    "language" = "english",
    "version"  = "1.0"
  }
  
  # Optional: Set custom chunking strategy for this file
  chunking_strategy {
    type = "auto"
  }
}
```

## Argument Reference

* `vector_store_id` - (Required) The ID of the vector store to add the file to.
* `file_id` - (Required) The ID of the file to add to the vector store.
* `attributes` - (Optional) Dynamic description or tags for the file in the vector store.
* `chunking_strategy` - (Optional) The chunking strategy used to chunk the file.
  * `type` - (Required) The type of chunking strategy (auto, fixed, or semantic).
  * `size` - (Optional) The size in characters for fixed chunking strategy.
  * `max_tokens` - (Optional) The maximum tokens per chunk for semantic chunking strategy.

## Attributes Reference

* `id` - The ID of the file (same as file_id).
* `created_at` - The timestamp for when the file was added to the vector store.
* `object` - The object type (always 'vector_store.file').
* `status` - The current status of the file in the vector store.

## Import

Vector store files can be imported using the format `vector_store_id/file_id`:

```bash
terraform import openai_vector_store_file.example vs_abc123def456/file-xyz789
```

The import process:

1. Retrieves the file's metadata directly from the OpenAI API
2. Sets both required attributes (`vector_store_id` and `file_id`) from the import ID
3. Populates all computed attributes like `created_at`, `object`, and `status` from the API
4. Includes any custom attributes or chunking strategy that was applied to the file

### Import Example

```bash
# First, remove the resource from state if it exists
terraform state rm openai_vector_store_file.document_in_store

# Then import using the combined ID format
terraform import openai_vector_store_file.document_in_store vs_abc123def456/file-xyz789
```

After importing, ensure your Terraform configuration has at minimum:

```hcl
resource "openai_vector_store_file" "document_in_store" {
  vector_store_id = "vs_abc123def456"  # From the import ID
  file_id         = "file-xyz789"      # From the import ID
}
```

### Import Considerations

1. The import command requires both the vector store ID and the file ID in the specified format
2. You may need to import the referenced file separately if you want to manage it in Terraform
3. The vector store resource should also be imported separately if not already managed in Terraform 