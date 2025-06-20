# openai_vector_store_files Data Source

Retrieves a list of files within an OpenAI Vector Store.

## Example Usage

```hcl
data "openai_vector_store_files" "example" {
  vector_store_id = "vs_12345abcde"
  limit           = 20
  order           = "desc"
  filter          = "completed"
}

output "vector_store_files" {
  value = data.openai_vector_store_files.example.files
}
```

## Argument Reference

* `vector_store_id` - (Required) The ID of the vector store that the files belong to.
* `limit` - (Optional) A limit on the number of objects to be returned. Limit can range between 1 and 100, and the default is 20.
* `order` - (Optional) Sort order by the created_at timestamp of the objects. "asc" for ascending order and "desc" for descending order. Default is "desc".
* `after` - (Optional) A cursor for use in pagination. "after" is an object ID that defines your place in the list.
* `before` - (Optional) A cursor for use in pagination. "before" is an object ID that defines your place in the list.
* `filter` - (Optional) Filter by file status. One of "in_progress", "completed", "failed", "cancelled".

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `files` - A list of files in the vector store. Each element contains the following attributes:
  * `id` - The ID of the file.
  * `object` - The object type (always "vector_store.file").
  * `created_at` - The Unix timestamp when the file was created.
  * `status` - The status of the file.
* `has_more` - Boolean indicating whether there are more files available beyond the current response. 