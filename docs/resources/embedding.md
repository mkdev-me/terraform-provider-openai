---
page_title: "OpenAI: openai_embedding Resource"
subcategory: ""
description: |-
  Creates vector embeddings for text using OpenAI embedding models.
---

# openai_embedding Resource

The `openai_embedding` resource creates vector embeddings for text using OpenAI's embedding models. Embeddings are numerical representations of text that capture semantic meaning, allowing for operations like semantic search, clustering, and recommendations.

## Example Usage

```hcl
resource "openai_embedding" "example" {
  model = "text-embedding-ada-002"
  input = "The quick brown fox jumps over the lazy dog."
  
  # Optional: Use your own API key for this specific resource
  api_key = var.openai_api_key
}

# Access the generated embedding
output "embedding_vector" {
  value = openai_embedding.example.embedding
}

# Process multiple inputs
resource "openai_embedding" "multiple" {
  model = "text-embedding-ada-002"
  input = [
    "The quick brown fox jumps over the lazy dog.",
    "The five boxing wizards jump quickly.",
    "Pack my box with five dozen liquor jugs."
  ]
}

output "multiple_embeddings" {
  value = openai_embedding.multiple.embeddings
}
```

## Argument Reference

* `model` - (Required) The ID of the model to use for generating embeddings. Common models include:
  * `text-embedding-ada-002`
  * `text-embedding-3-small`
  * `text-embedding-3-large`
* `input` - (Required) The text to convert to an embedding. Can be a string or a list of strings.
* `user` - (Optional) A unique identifier representing your end-user, which can help OpenAI monitor and detect abuse.
* `encoding_format` - (Optional) The format of the embeddings, either "float" or "base64". Defaults to "float".
* `dimensions` - (Optional) The number of dimensions the resulting output embeddings should have. Only supported in some models.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A unique identifier for this embedding resource.
* `embedding` - The embedding for a single input (when `input` is a string).
* `embeddings` - A list of embeddings for multiple inputs (when `input` is a list of strings).
* `usage` - Information about token usage:
  * `prompt_tokens` - The number of tokens in the input.
  * `total_tokens` - The total number of tokens used.

## Import

Embedding resources cannot be imported because they represent one-time API calls rather than persistent resources. 