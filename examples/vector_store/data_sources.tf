# List all Vector Stores with optional pagination
data "openai_vector_stores" "all" {
  limit = 20
  order = "desc"
}

# Get details about a specific vector store
data "openai_vector_store" "existing" {
  id = openai_vector_store.custom_store.id
}

# List files in a vector store
data "openai_vector_store_files" "store_files" {
  vector_store_id = openai_vector_store.custom_store.id
  limit           = 10
  order           = "desc"
}

# Get details about a specific file in a vector store
data "openai_vector_store_file" "specific_file" {
  vector_store_id = openai_vector_store.custom_store.id
  file_id         = openai_file.test_file.id
}

# Get content of a specific file in a vector store
data "openai_vector_store_file_content" "file_content" {
  vector_store_id = openai_vector_store.custom_store.id
  file_id         = openai_file.test_file.id
}

# Get details about a file batch in a vector store
data "openai_vector_store_file_batch" "specific_batch" {
  vector_store_id = openai_vector_store.custom_store.id
  batch_id        = openai_vector_store_file_batch.batch_upload.id
}

# List files in a file batch in a vector store
data "openai_vector_store_file_batch_files" "batch_files" {
  vector_store_id = openai_vector_store.custom_store.id
  batch_id        = openai_vector_store_file_batch.batch_upload.id
  limit           = 10
  order           = "desc"
}

# Show all vector stores in your account
output "vector_stores_list" {
  value = data.openai_vector_stores.all.vector_stores
}

# Show detailed information about a specific vector store
output "vector_store_details" {
  value = {
    id         = data.openai_vector_store.existing.id
    name       = data.openai_vector_store.existing.name
    file_count = data.openai_vector_store.existing.file_count
    status     = data.openai_vector_store.existing.status
    created_at = data.openai_vector_store.existing.created_at
    file_ids   = data.openai_vector_store.existing.file_ids
  }
}

# Show list of files in a vector store
output "vector_store_files_list" {
  value = data.openai_vector_store_files.store_files.files
}

# Show detailed information about a specific file in a vector store
output "vector_store_file_details" {
  value = {
    id         = data.openai_vector_store_file.specific_file.id
    created_at = data.openai_vector_store_file.specific_file.created_at
    status     = data.openai_vector_store_file.specific_file.status
    attributes = data.openai_vector_store_file.specific_file.attributes
  }
}

# Show file content from a vector store
output "vector_store_file_content" {
  value = substr(data.openai_vector_store_file_content.file_content.content, 0, 100)
}

# Show detailed information about a file batch in a vector store
output "vector_store_file_batch_details" {
  value = {
    id         = data.openai_vector_store_file_batch.specific_batch.id
    created_at = data.openai_vector_store_file_batch.specific_batch.created_at
    status     = data.openai_vector_store_file_batch.specific_batch.status
    file_ids   = data.openai_vector_store_file_batch.specific_batch.file_ids
    batch_type = data.openai_vector_store_file_batch.specific_batch.batch_type
    purpose    = data.openai_vector_store_file_batch.specific_batch.purpose
  }
}

# Show list of files in a batch
output "vector_store_file_batch_files_list" {
  value = data.openai_vector_store_file_batch_files.batch_files.files
} 