# Fetch a specific file by ID
data "openai_file" "training_data" {
  file_id = "file-abc123"
}

# Output file ID
output "file_id" {
  value = data.openai_file.training_data.id
}
