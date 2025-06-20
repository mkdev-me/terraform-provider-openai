# openai_vector_store_file Data Source

Retrieves information about a specific file within an OpenAI Vector Store.

## Example Usage

```hcl
data "openai_vector_store_file" "example" {
  vector_store_id = "vs_12345abcde"
  file_id         = "file_67890fghij"
}

output "file_details" {
  value = {
    id         = data.openai_vector_store_file.example.id
    created_at = data.openai_vector_store_file.example.created_at
    status     = data.openai_vector_store_file.example.status
    attributes = data.openai_vector_store_file.example.attributes
  }
}
```

## Argument Reference

* `vector_store_id` - (Required) The ID of the vector store.
* `file_id` - (Required) The ID of the file to retrieve details for.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the file.
* `object` - The object type (always "vector_store.file").
* `created_at` - The Unix timestamp when the file was created.
* `status` - The status of the file.
* `attributes` - Attributes of the file.
  * `size` - The size of the file in bytes.
  * `filename` - The name of the file.
  * `purpose` - The purpose of the file. 