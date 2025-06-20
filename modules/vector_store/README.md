# OpenAI Vector Store Module

This Terraform module provides a streamlined interface for creating and managing OpenAI Vector Stores. Vector stores allow you to store and search through vector embeddings of files for use with tools like file_search in assistants.

## Features

- Create and manage OpenAI Vector Stores
- Add files to Vector Stores individually or in batches
- Configure chunking strategies for optimal embedding extraction
- Set expiration policies for Vector Stores
- Attach metadata to Vector Stores and files

## Usage

### Basic Vector Store

```hcl
module "knowledge_base" {
  source = "../../modules/vector_store"
  
  name = "Company Knowledge Base"
}
```

### Vector Store with Files (Individual)

```hcl
module "support_faq" {
  source = "../../modules/vector_store"
  
  name     = "Support FAQ"
  file_ids = [
    openai_file.faq_document.id,
    openai_file.troubleshooting_guide.id
  ]
  
  # Add files individually (default)
  use_file_batches = false
  
  # Add metadata
  metadata = {
    "department" = "support",
    "version"    = "1.0"
  }
}
```

### Vector Store with File Batches

```hcl
module "product_documentation" {
  source = "../../modules/vector_store"
  
  name     = "Product Documentation"
  file_ids = [
    openai_file.user_manual.id,
    openai_file.api_reference.id,
    openai_file.examples.id
  ]
  
  # Use file batches for better performance with many files
  use_file_batches = true
  
  # Optional: Specify chunking strategy
  chunking_strategy = {
    type = "fixed",
    size = 1000
  }
  
  # Optional: Set expiration
  expires_after = {
    days = 90
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| `name` | The name of the vector store | `string` | n/a | yes |
| `file_ids` | A list of File IDs that the vector store should use | `list(string)` | `[]` | no |
| `metadata` | Set of key-value pairs that can be attached to the vector store | `map(string)` | `{}` | no |
| `chunking_strategy` | The chunking strategy used to chunk the files | `object` | `{type = "auto"}` | no |
| `expires_after` | The expiration policy for the vector store | `object` | `null` | no |
| `use_file_batches` | Whether to use file batches for adding files | `bool` | `false` | no |
| `file_attributes` | Set of key-value pairs for file objects | `map(string)` | `{}` | no |

### Chunking Strategy Options

The `chunking_strategy` variable accepts the following formats:

```hcl
# Auto (default)
chunking_strategy = {
  type = "auto"
}

# Fixed size chunks
chunking_strategy = {
  type = "fixed"
  size = 1000  # Characters per chunk
}

# Semantic chunking
chunking_strategy = {
  type       = "semantic"
  max_tokens = 300  # Maximum tokens per chunk
}
```

### Expiration Policy Options

The `expires_after` variable accepts the following formats:

```hcl
# Expire after a number of days
expires_after = {
  days = 30
}

# Never expire
expires_after = {
  never = true
}
```

## Outputs

| Name | Description |
|------|-------------|
| `id` | The ID of the created vector store |
| `name` | The name of the vector store |
| `created_at` | The timestamp when the vector store was created |
| `file_count` | The number of files in the vector store |
| `object` | The type of object (always 'vector_store') |
| `status` | The current status of the vector store |

## Implementation Details

The module creates the following resources:

1. An `openai_vector_store` resource to manage the vector store
2. Either:
   - Multiple `openai_vector_store_file` resources (when `use_file_batches = false`)
   - A single `openai_vector_store_file_batch` resource (when `use_file_batches = true`)

## Notes

- File IDs must refer to files that have already been uploaded to OpenAI
- Files must have the appropriate purpose (e.g., "assistants")
- Vector stores are in beta and require the `assistants=v2` beta header
- When working with many files, using file batches (`use_file_batches = true`) is more efficient
- The maximum number of key-value pairs in metadata and attributes is 16

## Common API Errors

### Invalid File ID

**Error:** `File not found or not accessible`

**Solution:**
- Verify that the file ID is correct
- Ensure the file has been uploaded to OpenAI
- Check that the file has the appropriate purpose

### Rate Limiting

**Error:** `Rate limit exceeded`

**Solution:**
- Implement exponential backoff and retry logic
- Reduce the frequency of requests to the API

## Related Resources

- [OpenAI Vector Store API Documentation](https://platform.openai.com/docs/api-reference/vector-stores)
- [OpenAI File API Documentation](https://platform.openai.com/docs/api-reference/files) 