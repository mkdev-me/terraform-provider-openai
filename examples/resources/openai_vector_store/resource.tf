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
