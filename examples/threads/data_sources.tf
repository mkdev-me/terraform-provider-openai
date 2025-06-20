# Data source for retrieving a single thread by ID
data "openai_thread" "existing_thread" {
  # This will reference the ID of one of the created threads
  thread_id = openai_thread.with_messages.id
}

# Outputs for the thread data source
output "data_thread_id" {
  description = "The ID of the thread retrieved from the data source"
  value       = data.openai_thread.existing_thread.id
}

output "data_thread_created_at" {
  description = "The timestamp when the thread was created"
  value       = data.openai_thread.existing_thread.created_at
}

output "data_thread_metadata" {
  description = "The metadata attached to the thread"
  value       = data.openai_thread.existing_thread.metadata
} 