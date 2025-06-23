# Example: Adding multiple files to a vector store in a single batch operation
# File batches are more efficient than adding files individually

# Create a vector store for documentation
resource "openai_vector_store" "documentation" {
  name = "Product Documentation Hub"
  metadata = {
    project = "main-product"
    type    = "documentation"
    year    = "2024"
  }
}

# Upload multiple documentation files
resource "openai_file" "user_guide" {
  file    = "docs/user_guide.pdf"
  purpose = "assistants"
}

resource "openai_file" "admin_manual" {
  file    = "docs/admin_manual.pdf"
  purpose = "assistants"
}

resource "openai_file" "api_docs" {
  file    = "docs/api_documentation.md"
  purpose = "assistants"
}

resource "openai_file" "troubleshooting" {
  file    = "docs/troubleshooting_guide.txt"
  purpose = "assistants"
}

resource "openai_file" "release_notes" {
  file    = "docs/release_notes_2024.md"
  purpose = "assistants"
}

# Add all documentation files to the vector store in one batch
resource "openai_vector_store_file_batch" "documentation_batch" {
  # The vector store to add files to
  vector_store_id = openai_vector_store.documentation.id

  # List of file IDs to add
  file_ids = [
    openai_file.user_guide.id,
    openai_file.admin_manual.id,
    openai_file.api_docs.id,
    openai_file.troubleshooting.id,
    openai_file.release_notes.id
  ]

  # Optional: Default chunking strategy for all files
  chunking_strategy {
    type = "fixed"
    size = 600
  }
}

# Example: Knowledge base with mixed content types
resource "openai_vector_store" "knowledge_base" {
  name = "Company Knowledge Base"
}

# Upload various file types
resource "openai_file" "policies" {
  file    = "company_policies.pdf"
  purpose = "assistants"
}

resource "openai_file" "procedures" {
  file    = "standard_procedures.docx"
  purpose = "assistants"
}

resource "openai_file" "training" {
  file    = "training_materials.pptx"
  purpose = "assistants"
}

resource "openai_file" "faqs" {
  file    = "employee_faqs.txt"
  purpose = "assistants"
}

# Batch add with auto chunking
resource "openai_vector_store_file_batch" "knowledge_batch" {
  vector_store_id = openai_vector_store.knowledge_base.id

  file_ids = [
    openai_file.policies.id,
    openai_file.procedures.id,
    openai_file.training.id,
    openai_file.faqs.id
  ]

  # Use automatic chunking for mixed content
  chunking_strategy {
    type = "auto"
  }
}

# Example: Research papers batch
resource "openai_vector_store" "research_library" {
  name = "AI Research Papers"
  metadata = {
    field = "machine_learning"
    year  = "2024"
  }
}

# Dynamic file collection
locals {
  research_files = [
    "papers/transformer_architecture.pdf",
    "papers/attention_mechanisms.pdf",
    "papers/neural_networks_survey.pdf",
    "papers/deep_learning_advances.pdf",
    "papers/llm_evaluation.pdf"
  ]
}

# Upload research papers dynamically
resource "openai_file" "research_papers" {
  for_each = toset(local.research_files)

  file    = each.value
  purpose = "assistants"
}

# Batch add all research papers
resource "openai_vector_store_file_batch" "research_batch" {
  vector_store_id = openai_vector_store.research_library.id

  # Collect all file IDs dynamically
  file_ids = [for paper in openai_file.research_papers : paper.id]

  # Academic papers need larger chunks for context
  chunking_strategy {
    type = "fixed"
    size = 1000
  }
}

# Output batch status
output "documentation_batch_status" {
  value = openai_vector_store_file_batch.documentation_batch.status
}
