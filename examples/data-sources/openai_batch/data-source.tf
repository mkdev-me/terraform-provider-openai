# Fetch a specific batch job by ID
data "openai_batch" "embeddings_batch" {
  batch_id = "batch_abc123"
}

# Output batch job details
output "batch_status" {
  value = data.openai_batch.embeddings_batch.status
}
