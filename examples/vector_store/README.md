# OpenAI Vector Store Examples

This directory contains examples demonstrating how to use the OpenAI Vector Store resources and modules with Terraform.

## Prerequisites

Before using these examples, ensure you have:

1. Installed Terraform (v1.0.0 or later)
2. Set up the OpenAI provider with a valid API key:
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```
3. Files uploaded to OpenAI with the purpose="vector_store_file" (for examples using files)

## Examples Overview

The examples in this directory demonstrate various approaches to creating and managing OpenAI Vector Stores:

1. **Basic Vector Store** - Creates a simple vector store without any files
2. **Support Knowledge Base** - Creates a vector store with individually added files and auto chunking strategy
3. **API Documentation Store** - Creates a vector store using file batches with auto chunking strategy
4. **Custom Vector Store** - Creates a vector store directly using the provider resource (without the module)

## Configuration Requirements

### Chunking Strategy

The chunking strategy determines how documents are divided into chunks for embedding:

```hcl
chunking_strategy {
  type = "auto"  # Only "auto" or "static" are supported
}
```

**Important notes:**
- Only `type` parameter is supported
- Valid values are only "auto" or "static"
- Previous parameters like "size" and "max_tokens" are no longer supported by the API

### Expiration Policy

The expiration policy controls when the vector store will be deleted:

```hcl
expires_after {
  days   = 90    # Number of days until expiration
  anchor = "last_active_at"
}
```

**Important notes:**
- Both parameters are required
- `days` must be 1 or greater
- `anchor` must be set to "last_active_at" (the only supported value)

## File Structure

- `main.tf` - The main configuration file containing all examples
- `outputs.tf` - Defines outputs to demonstrate the values returned by vector stores
- `files/` - Contains sample files to be uploaded to vector stores

## Running the Examples

1. Navigate to this directory:
   ```bash
   cd examples/vector_store
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

4. To clean up resources:
   ```bash
   terraform destroy
   ```

## Module vs Direct Resource Usage

The examples showcase two approaches:

### Using the Module (Recommended)

```hcl
module "basic_vector_store" {
  source = "../../modules/vector_store"
  
  name = "Basic Knowledge Base"
  
  # For expires_after, both parameters are required
  expires_after = {
    days = 365,
    anchor = "last_active_at"
  }
}
```

### Using Direct Resources

```hcl
resource "openai_vector_store" "custom_store" {
  name = "Custom Vector Store"
  
  # For expires_after as a block, both parameters are required
  expires_after {
    days = 180
    anchor = "last_active_at"
  }
  
  chunking_strategy {
    type = "auto"
  }
}
```

## Error Handling and Troubleshooting

Common errors and their solutions:

1. **Unknown parameter errors**:
   - Error: `Unknown parameter: 'chunking_strategy.size'`
   - Solution: Only use the `type` parameter in chunking_strategy

2. **Invalid value errors**:
   - Error: `Invalid value: 'now'. Value must be 'last_active_at'`
   - Solution: Only use "last_active_at" for the anchor parameter

3. **Integer below minimum value**:
   - Error: `Invalid 'expires_after.days': integer below minimum value`
   - Solution: The days parameter must be 1 or greater

## Notes

- Vector Store is currently in Beta and subject to changes in the API
- The OpenAI API may return different errors as the service evolves
- For large numbers of files, the batch approach is more efficient 