# Create a file with batch requests for chat completions
resource "openai_file" "batch_input" {
  file    = "${path.module}/batch_requests.jsonl"
  purpose = "batch"
}

# Create a batch job to process multiple chat completion requests
resource "openai_batch" "chat_completions" {
  input_file_id = openai_file.batch_input.id
  endpoint      = "/v1/chat/completions"

  # Optional: Set completion window (default is 24h)
  completion_window = "24h"

  # Optional: Add metadata for tracking
  metadata = {
    department = "customer-support"
    batch_type = "daily-analysis"
    version    = "1.0"
  }
}

# Output the batch ID
output "batch_id" {
  value = openai_batch.chat_completions.id
}
