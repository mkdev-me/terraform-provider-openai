# OpenAI Embeddings Module

This module provides a simple way to generate embeddings using the OpenAI API. Embeddings are vector representations of text that capture semantic meaning, making them useful for tasks such as search, clustering, recommendations, and other natural language processing applications.

## Requirements

- Have the Terraform provider for OpenAI installed (`fjcorp/openai`)
- A valid OpenAI API key

## Usage

Basic example:

```hcl
module "text_embedding" {
  source = "../../modules/embeddings"
  
  input = "This is an example text to generate an embedding"
  model = "text-embedding-ada-002"
}

output "embedding_result" {
  value = module.text_embedding.embeddings
  sensitive = true
}
```

## Input Variables

| Name | Description | Type | Default | Required |
|--------|-------------|------|------------------|-----------|
| `model` | ID of the model to use (e.g., text-embedding-ada-002) | `string` | `"text-embedding-ada-002"` | No |
| `input` | Text to generate the embedding for. Can be a string or a JSON array of strings | `string` | - | Yes |
| `user` | Optional unique identifier representing the end user | `string` | `null` | No |
| `encoding_format` | Format to return the embeddings: 'float' or 'base64' | `string` | `"float"` | No |
| `dimensions` | Number of dimensions for the resulting embeddings (only available in text-embedding-3 and later models) | `number` | `null` | No |
| `project_id` | OpenAI project ID to use for this request | `string` | `null` | No |

## Outputs

| Name | Description |
|--------|-------------|
| `embeddings` | The generated embeddings |
| `model_used` | The model used to generate the embeddings |
| `usage` | Token usage statistics for the request |
| `embedding_id` | The ID of the generated embedding |

## Additional Notes

- Embeddings are calculated once during the Terraform apply and then maintained in the state.
- To generate embeddings for multiple texts in a single request, you can pass a JSON array as a string in the `input` variable.
- Example of JSON array: `"[\"First text\", \"Second text\", \"Third text\"]"`
- The maximum input size varies by model (e.g., 8192 tokens for text-embedding-ada-002).
- For newer models like text-embedding-3, you can specify the number of dimensions using the `dimensions` variable.

## API Limitations

**Important**: The OpenAI API does not provide endpoints to retrieve or list previously created embeddings. Due to this API limitation:

1. This module only supports creating new embeddings
2. There is no corresponding data source available for embeddings
3. Embeddings cannot be retrieved after creation by ID or any other identifier

### Import Handling

When importing existing embeddings, you'll encounter these challenges:

1. **Partial State Only**: Only basic metadata is imported (ID, timestamps, etc.)
2. **No Vector Data**: The actual embedding vectors cannot be retrieved from the API
3. **Simulation Required**: This module uses a simulation approach to represent imported embeddings

The module addresses these limitations by:

1. Using a fault-tolerant design that works with both new and imported resources
2. Simulating embedding vectors when the real ones cannot be retrieved
3. Providing consistent outputs regardless of whether the resource is newly created or imported

After importing an embedding, a subsequent `terraform apply` will replace the imported resource with a new one, since the complete state of the original resource cannot be determined from the API.

#### Import Example

```bash
# First import the resource
terraform import module.my_embedding.openai_chat_completion.embedding_simulation chatcmpl-XXXXXXXXXXXXXXXXXXXX

# Then accept that applying will recreate it
terraform apply
```

The module ensures full compatibility with all OpenAI API parameters for embeddings:

```json
{
  "input": "The text to embed",       // Required: String or array of strings
  "model": "text-embedding-ada-002",  // Required: Model ID
  "encoding_format": "float",         // Optional: "float" (default) or "base64"
  "dimensions": 1536,                 // Optional: Only for text-embedding-3+ models
  "user": "user-123"                  // Optional: End-user identifier
}
```

If you need to store or retrieve embeddings, consider using a vector database or another storage mechanism outside of Terraform. 