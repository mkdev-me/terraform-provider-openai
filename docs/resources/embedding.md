---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "openai_embedding Resource - terraform-provider-openai"
subcategory: ""
description: |-
  
---

# openai_embedding (Resource)



## Example Usage

```terraform
resource "openai_embedding" "example" {
  model = "text-embedding-3-small"
  input = jsonencode(["The quick brown fox jumps over the lazy dog."])
}

output "embedding_vector" {
  value     = openai_embedding.example.embeddings[0].embedding
  sensitive = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `input` (String) The text to embed
- `model` (String) ID of the model to use for the embedding

### Optional

- `dimensions` (Number) The number of dimensions to use for the embedding (only specific models support this)
- `encoding_format` (String) The format of the embeddings. One of 'float', 'base64'
- `project_id` (String) The project to use for this request
- `user` (String) A unique identifier representing your end-user

### Read-Only

- `embedding_id` (String) The ID of the embedding
- `embeddings` (List of Object) The embeddings generated for the input (see [below for nested schema](#nestedatt--embeddings))
- `id` (String) The ID of this resource.
- `model_used` (String) The model used for the embedding
- `object` (String) The object type, which is always 'embedding'
- `usage` (Map of Number) Usage statistics for the embedding request

<a id="nestedatt--embeddings"></a>
### Nested Schema for `embeddings`

Read-Only:

- `embedding` (List of Number)
- `index` (Number)
- `object` (String)
