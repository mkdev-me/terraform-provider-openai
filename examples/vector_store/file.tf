resource "openai_file" "test_file" {
  file    = "test_file.txt"
  purpose = "assistants"
}

# Add the file to the custom vector store
resource "openai_vector_store_file" "custom_file" {
  vector_store_id = openai_vector_store.custom_store.id
  file_id         = openai_file.test_file.id

  attributes = {
    "category" = "test",
    "purpose"  = "demonstration"
  }
}

# Output the file information
output "file_id" {
  value = openai_file.test_file.id
}

output "vector_store_file_id" {
  value = openai_vector_store_file.custom_file.id
} 