# Output values for the OpenAI Run module

output "run_id" {
  description = "The ID of the created run"
  value       = openai_run.run.id
}

output "status" {
  description = "The current status of the run"
  value       = openai_run.run.status
}

output "created_at" {
  description = "The timestamp when the run was created"
  value       = openai_run.run.created_at
}

output "started_at" {
  description = "The timestamp when the run was started"
  value       = openai_run.run.started_at
}

output "completed_at" {
  description = "The timestamp when the run was completed"
  value       = openai_run.run.completed_at
}

output "usage" {
  description = "Token usage statistics for the run"
  value       = openai_run.run.usage
} 