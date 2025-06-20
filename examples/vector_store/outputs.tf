output "basic_vector_store_id" {
  description = "The ID of the basic vector store."
  value       = module.basic_vector_store.id
}

output "support_vector_store_details" {
  description = "Details about the support vector store."
  value = {
    id         = module.support_vector_store.id
    name       = module.support_vector_store.name
    created_at = module.support_vector_store.created_at
    file_count = module.support_vector_store.file_count
    status     = module.support_vector_store.status
  }
}

output "api_vector_store_id" {
  description = "The ID of the API documentation vector store."
  value       = module.api_vector_store.id
}

output "custom_vector_store_id" {
  description = "The ID of the custom vector store created without using the module."
  value       = openai_vector_store.custom_store.id
}

output "all_vector_store_ids" {
  description = "All vector store IDs created in this example."
  value = [
    module.basic_vector_store.id,
    module.support_vector_store.id,
    module.api_vector_store.id,
    openai_vector_store.custom_store.id
  ]
} 