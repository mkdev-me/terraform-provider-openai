# List all files with optional filtering
data "openai_files" "all" {
  # Optional: filter by purpose
  purpose = "fine-tune"
}

# List files for assistants
data "openai_files" "assistant_files" {
  purpose = "assistants"
}

# Output total files count
output "total_files" {
  value = length(data.openai_files.all.files)
}
