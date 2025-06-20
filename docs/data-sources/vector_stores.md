# openai_vector_stores Data Source

Retrieves a list of OpenAI Vector Stores associated with your organization.

## Example Usage

```hcl
data "openai_vector_stores" "all" {
  limit = 20
  order = "desc"
}

output "vector_stores" {
  value = data.openai_vector_stores.all.vector_stores
}
```

## Argument Reference

* `limit` - (Optional) A limit on the number of objects to be returned. Limit can range between 1 and 100, and the default is 20.
* `order` - (Optional) Sort order by the created_at timestamp of the objects. "asc" for ascending order and "desc" for descending order. Default is "desc".
* `after` - (Optional) A cursor for use in pagination. "after" is an object ID that defines your place in the list.
* `before` - (Optional) A cursor for use in pagination. "before" is an object ID that defines your place in the list.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `vector_stores` - A list of vector stores. Each element contains the following attributes:
  * `id` - The ID of the vector store.
  * `name` - The name of the vector store.
  * `object` - The object type (always "vector_store").
  * `created_at` - The Unix timestamp when the vector store was created.
  * `file_count` - The number of files in the vector store.
  * `status` - The current status of the vector store.
* `has_more` - Boolean indicating whether there are more vector stores available beyond the current response. 