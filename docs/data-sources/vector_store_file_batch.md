# openai_vector_store_file_batch Data Source

Retrieves information about a specific file batch within an OpenAI Vector Store.

## Example Usage

```hcl
data "openai_vector_store_file_batch" "example" {
  vector_store_id = "vs_12345abcde"
  batch_id        = "vsfb_67890fghij"
}

output "batch_details" {
  value = {
    id         = data.openai_vector_store_file_batch.example.id
    created_at = data.openai_vector_store_file_batch.example.created_at
    status     = data.openai_vector_store_file_batch.example.status
    file_ids   = data.openai_vector_store_file_batch.example.file_ids
  }
}
```

## Argument Reference

* `vector_store_id` - (Required) The ID of the vector store.
* `batch_id` - (Required) The ID of the file batch to retrieve details for.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the file batch.
* `object` - The object type (always "vector_store.file_batch").
* `created_at` - The Unix timestamp when the file batch was created.
* `status` - The status of the file batch.
* `file_ids` - The list of file IDs in the batch.
* `batch_type` - The type of the batch.
* `purpose` - The purpose of the file batch. 