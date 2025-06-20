# Example of using the OpenAI files data source
# This example demonstrates retrieving and filtering OpenAI files

# List all files without filtering
data "openai_files" "all" {}

# List files filtered by purpose
data "openai_files" "fine_tune_files" {
  purpose = "fine-tune"
}

data "openai_files" "assistant_files" {
  purpose = "assistants"
}

# Output count of all files
output "total_files_count" {
  description = "Total number of files in the account"
  value       = length(data.openai_files.all.files)
}

# Output file information for all files
output "all_files" {
  description = "Information about all files"
  value       = data.openai_files.all.files
}

# Output just the file IDs and names for easier viewing
output "file_names" {
  description = "Names of all files"
  value       = [for file in data.openai_files.all.files : "${file.id}: ${file.filename}"]
}

# Output fine-tune files count
output "fine_tune_files_count" {
  description = "Number of fine-tune files"
  value       = length(data.openai_files.fine_tune_files.files)
}

# Output assistant files count
output "assistant_files_count" {
  description = "Number of assistant files"
  value       = length(data.openai_files.assistant_files.files)
}

# Calculate file sizes
output "total_bytes_used" {
  description = "Total storage used by all files (in bytes)"
  value       = length(data.openai_files.all.files) > 0 ? sum([for file in data.openai_files.all.files : tonumber(file.bytes)]) : 0
}

# Group files by purpose
output "files_by_purpose" {
  description = "Files grouped by purpose"
  value = {
    for file in data.openai_files.all.files :
    file.purpose => file.filename...
  }
} 