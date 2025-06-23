resource "openai_file" "training_data" {
  file    = "training_data.jsonl"
  purpose = "fine-tune"
}

output "file_id" {
  value = openai_file.training_data.id
}
