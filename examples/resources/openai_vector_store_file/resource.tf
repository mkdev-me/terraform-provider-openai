# Example: Adding files to vector stores for use with Assistants
# Vector store files enable semantic search over document contents

# First, create a vector store
resource "openai_vector_store" "knowledge_base" {
  name = "Company Knowledge Base"
  metadata = {
    department = "engineering"
    version    = "1.0"
  }
}

# Upload files that will be added to the vector store
resource "openai_file" "technical_docs" {
  file    = "technical_documentation.pdf"
  purpose = "assistants"
}

resource "openai_file" "api_reference" {
  file    = "api_reference.md"
  purpose = "assistants"
}

resource "openai_file" "faq_document" {
  file    = "frequently_asked_questions.txt"
  purpose = "assistants"
}

# Add a file to the vector store
resource "openai_vector_store_file" "add_tech_docs" {
  # The vector store to add the file to
  vector_store_id = openai_vector_store.knowledge_base.id

  # The file to add
  file_id = openai_file.technical_docs.id

  # Optional: Chunking strategy
  chunking_strategy {
    type = "fixed"
    size = 800
  }
}

# Add API reference with auto chunking
resource "openai_vector_store_file" "add_api_ref" {
  vector_store_id = openai_vector_store.knowledge_base.id
  file_id         = openai_file.api_reference.id

  # Use automatic chunking (default)
  chunking_strategy {
    type = "auto"
  }
}

# Add FAQ document with custom chunking
resource "openai_vector_store_file" "add_faq" {
  vector_store_id = openai_vector_store.knowledge_base.id
  file_id         = openai_file.faq_document.id

  chunking_strategy {
    type = "fixed"
    size = 200 # Smaller chunks for Q&A format
  }
}

# Example: Multiple vector stores for different purposes
resource "openai_vector_store" "customer_support" {
  name = "Customer Support Database"
}

resource "openai_file" "support_tickets" {
  file    = "resolved_tickets_2024.jsonl"
  purpose = "assistants"
}

resource "openai_vector_store_file" "support_knowledge" {
  vector_store_id = openai_vector_store.customer_support.id
  file_id         = openai_file.support_tickets.id

  # Larger chunks for conversation context
  chunking_strategy {
    type = "fixed"
    size = 1200
  }
}

# Example: Code repository vector store
resource "openai_vector_store" "code_search" {
  name = "Codebase Search"
  metadata = {
    language = "python"
    project  = "backend-api"
  }
}

resource "openai_file" "source_code" {
  file    = "backend_source.zip"
  purpose = "assistants"
}

resource "openai_vector_store_file" "code_index" {
  vector_store_id = openai_vector_store.code_search.id
  file_id         = openai_file.source_code.id

  # Specific chunking for code
  chunking_strategy {
    type = "fixed"
    size = 500
  }
}

# Output file status
output "tech_docs_status" {
  value = openai_vector_store_file.add_tech_docs.status
}
