output "id" {
  description = "The ID of the vector store"
  value       = openai_vector_store.this.id
}

output "name" {
  description = "The name of the vector store"
  value       = openai_vector_store.this.name
}

output "file_count" {
  description = "The number of files in the vector store"
  value       = openai_vector_store.this.file_count
}

output "status" {
  description = "The status of the vector store"
  value       = openai_vector_store.this.status
}

output "created_at" {
  description = "The creation timestamp of the vector store"
  value       = openai_vector_store.this.created_at
}

output "details" {
  description = "A map of vector store details"
  value = {
    id         = openai_vector_store.this.id
    name       = openai_vector_store.this.name
    status     = openai_vector_store.this.status
    file_count = openai_vector_store.this.file_count
    created_at = openai_vector_store.this.created_at
  }
} 