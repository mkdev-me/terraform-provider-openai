---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "openai_vector_store Resource - terraform-provider-openai"
subcategory: ""
description: |-
  
---

# openai_vector_store (Resource)



## Example Usage

```terraform
# Create a vector store for document retrieval
resource "openai_vector_store" "knowledge_base" {
  name = "Company Knowledge Base"

  # Optional: Add metadata for organization
  metadata = {
    department   = "documentation"
    version      = "2.0"
    last_updated = "2024-01-15"
    access_level = "internal"
  }
}

# Create a vector store for customer support documents
resource "openai_vector_store" "support_docs" {
  name = "Customer Support Documentation"

  # Optional: Configure file handling
  file_ids = [
    # Reference existing files to include in the vector store
    # These would typically be created with openai_file resources
    # "file-abc123",
    # "file-def456"
  ]

  metadata = {
    category     = "support"
    language     = "en-US"
    indexed_date = "2024-01-15"
  }
}

# Create a vector store for code documentation
resource "openai_vector_store" "code_docs" {
  name = "API and Code Documentation"

  metadata = {
    repository  = "github.com/company/api-docs"
    doc_type    = "technical"
    api_version = "v3"
    team        = "platform-engineering"
  }
}

# Create a vector store for product manuals
resource "openai_vector_store" "product_manuals" {
  name = "Product User Manuals"

  metadata = {
    product_line = "enterprise"
    format       = "pdf"
    languages    = "en,es,fr,de"
    compliance   = "ISO-9001"
  }
}

# Output the vector store ID
output "knowledge_base_id" {
  value       = openai_vector_store.knowledge_base.id
  description = "The ID of the knowledge base vector store"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the vector store.

### Optional

- `chunking_strategy` (Block List, Max: 1) The chunking strategy used to chunk the files. (see [below for nested schema](#nestedblock--chunking_strategy))
- `expires_after` (Block List, Max: 1) The expiration policy for the vector store. (see [below for nested schema](#nestedblock--expires_after))
- `file_ids` (List of String) The list of file IDs to use in the vector store.
- `metadata` (Map of String) Set of key-value pairs that can be attached to the vector store.

### Read-Only

- `created_at` (Number) The timestamp for when the vector store was created.
- `file_count` (Number) The number of files in the vector store.
- `id` (String) The ID of the vector store.
- `object` (String) The object type (always 'vector_store').
- `status` (String) The current status of the vector store.

<a id="nestedblock--chunking_strategy"></a>
### Nested Schema for `chunking_strategy`

Required:

- `type` (String) The type of chunking strategy (auto or static).


<a id="nestedblock--expires_after"></a>
### Nested Schema for `expires_after`

Required:

- `anchor` (String) The anchor time for the expiration (usually 'now').

Optional:

- `days` (Number) Number of days after which the vector store should expire.

## Import

Import is supported using the following syntax:

```shell
#!/bin/bash
# Import existing OpenAI vector store
terraform import openai_vector_store.example vs_abc123def456
```
