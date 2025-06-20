# OpenAI Threads Module Outputs

# Thread resource outputs
output "thread_id" {
  description = "The ID of the created thread"
  value       = var.enable_thread ? openai_thread.this[0].id : null
}

output "thread_created_at" {
  description = "The timestamp when the thread was created"
  value       = var.enable_thread ? openai_thread.this[0].created_at : null
}

output "thread_metadata" {
  description = "The metadata attached to the thread"
  value       = var.enable_thread ? openai_thread.this[0].metadata : null
}

# Thread data source outputs
output "single_thread" {
  description = "Details of a specific thread retrieved by ID"
  value       = var.enable_thread_data_source && var.thread_id != null ? data.openai_thread.single[0] : null
}

output "single_thread_id" {
  description = "ID of the specific thread retrieved"
  value       = var.enable_thread_data_source && var.thread_id != null ? data.openai_thread.single[0].id : null
}

output "single_thread_created_at" {
  description = "Timestamp when the specific thread was created"
  value       = var.enable_thread_data_source && var.thread_id != null ? data.openai_thread.single[0].created_at : null
}

output "single_thread_metadata" {
  description = "Metadata of the specific thread retrieved"
  value       = var.enable_thread_data_source && var.thread_id != null ? data.openai_thread.single[0].metadata : null
} 