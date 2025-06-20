# OpenAI Embeddings Example

This example demonstrates how to generate and use text embeddings with the OpenAI API through the Terraform provider for OpenAI.

## What are embeddings?

Embeddings are vector representations of text that capture their semantic meaning. They are useful for:

- Semantic search
- Similarity comparison between texts
- Clustering and classification
- Recommendation systems
- And other natural language processing applications

## Prerequisites

1. Terraform installed
2. An OpenAI API key
3. The OpenAI provider installed in `~/.terraform.d/plugins/`

## Configuration

1. Make sure you have the OpenAI provider correctly installed:
   ```
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/mkdev-me/openai/1.0.0/darwin_arm64
   cp ~/path/to/binary/terraform-provider-openai ~/.terraform.d/plugins/registry.terraform.io/mkdev-me/openai/1.0.0/darwin_arm64/
   ```

2. Configure the necessary environment variables:
   ```
   export OPENAI_API_KEY="your-api-key"
   # If you belong to an organization:
   export OPENAI_ORGANIZATION_ID="your-organization-id"
   ```

## Usage

This example includes:

1. **Basic Embedding**: Embedding generation for a single text
2. **Base64 Format Embedding**: Example of using an alternative format
3. **Multiple Embeddings**: Generating embeddings for multiple texts in a single request
4. **Embeddings with Custom Dimensions**: Example of using newer models with specific dimensions

To run the example:

```
terraform init
terraform apply
```

## Understanding the code

The `main.tf` file demonstrates:

- How to configure the OpenAI provider
- How to use the embeddings module for different use cases
- How to work with different parameters (model, format, dimensions)
- How to handle multiple texts in a single request

## Important notes

- The generated embeddings can be large, so they are not shown directly in the Terraform output
- The `text-embedding-ada-002` model has a limit of 8192 input tokens
- The total number of embeddings is limited per request and per model
- For newer models like `text-embedding-3-small`, you can specify the number of dimensions of the resulting vector

## API and Provider Limitations

**Important**: The OpenAI API does not currently provide a way to list or retrieve existing embeddings. As a result, this provider only supports creating embeddings as a resource (`openai_embedding`) and does not include a data source for retrieving previously created embeddings.

### Import Limitations

When importing existing embeddings, you'll face the following limitations:

1. **Partial Resource State**: Only basic metadata is imported (ID, created date, etc.), but the actual embedding vectors are not available
2. **No Retrieval API**: The OpenAI API has no endpoint to retrieve previously created embeddings, so the import process cannot fetch the original vector data
3. **Resource Replacement**: After import, applying the configuration will replace the imported resource with a newly created one

### Import Workaround

This module handles imports by:
1. Using simulated embeddings rather than the actual vectors (which can't be retrieved)
2. Providing a fault-tolerant structure that works with both new and imported resources
3. Accepting that imports are primarily for tracking existing resources, not for retrieving the actual embedding vectors

To import an existing embedding resource:

```bash
terraform import module.my_embedding.openai_chat_completion.embedding_simulation chatcmpl-XXXXXXXXXXXXXXXXXXXX
```

After import, a subsequent `terraform apply` will replace the imported resource with a newly created one, since the original embedding vectors cannot be retrieved from the API.

The provider's implementation supports all the official OpenAI API parameters for embeddings:
- `input`: Required - The text to embed (string or array of strings)
- `model`: Required - ID of the model to use (e.g., "text-embedding-ada-002")
- `dimensions`: Optional - The number of dimensions for the embeddings (only for text-embedding-3 and later models)
- `encoding_format`: Optional - Format for the embeddings, either "float" (default) or "base64"
- `user`: Optional - A unique identifier representing your end-user

Unlike other OpenAI resources, embeddings cannot be retrieved after creation, so store the results as needed in your application.

## Example of use in real applications

The generated embeddings can be exported and used in:

- Vector databases like Pinecone, Milvus, or Weaviate
- Semantic search systems
- Sentiment analysis and text classification
- Content similarity or duplication detection 