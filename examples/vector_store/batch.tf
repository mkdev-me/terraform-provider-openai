# Create a second file for testing batch uploads
resource "openai_file" "test_file2" {
  file    = "test_file.txt" # Using the same file for simplicity
  purpose = "assistants"
}

# Create a batch file upload resource
resource "openai_vector_store_file_batch" "batch_upload" {
  vector_store_id = module.api_vector_store.id
  file_ids = [
    openai_file.test_file.id,
    openai_file.test_file2.id
  ]
}

# Output batch info
output "batch_id" {
  value = openai_vector_store_file_batch.batch_upload.id
}

output "batch_status" {
  value = openai_vector_store_file_batch.batch_upload.status
}
