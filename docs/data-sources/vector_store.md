# openai_vector_store Data Source

Retrieves information about a specific OpenAI Vector Store.

## Example Usage

```hcl
data "openai_vector_store" "example" {
  id = "vs_12345abcde"
}

output "vector_store_details" {
  value = {
    name      = data.openai_vector_store.example.name
    status    = data.openai_vector_store.example.status
    file_ids  = data.openai_vector_store.example.file_ids
  }
}
```

## Argument Reference

* `id` - (Required) The ID of the vector store to retrieve details for.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `name` - The name of the vector store.
* `created_at` - The Unix timestamp when the vector store was created.
* `file_count` - The number of files in the vector store.
* `object` - The object type (always "vector_store").
* `status` - The current status of the vector store.
* `metadata` - Set of key-value pairs attached to the vector store.
* `file_ids` - The list of file IDs in the vector store.
* `expires_after` - The expiration policy for the vector store.
  * `days` - Number of days after which the vector store should expire.
  * `anchor` - The anchor time for the expiration.
* `chunking_strategy` - The chunking strategy used for the files in the store.
  * `type` - The type of chunking strategy. 