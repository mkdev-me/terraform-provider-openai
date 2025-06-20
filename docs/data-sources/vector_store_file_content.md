# openai_vector_store_file_content Data Source

Retrieves the content of a file within an OpenAI Vector Store.

## Example Usage

```hcl
data "openai_vector_store_file_content" "example" {
  vector_store_id = "vs_12345abcde"
  file_id         = "file_67890fghij"
}

output "file_content" {
  value = data.openai_vector_store_file_content.example.content
}
```

## Argument Reference

* `vector_store_id` - (Required) The ID of the vector store.
* `file_id` - (Required) The ID of the file within the vector store.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `content` - The content of the file as a string. 