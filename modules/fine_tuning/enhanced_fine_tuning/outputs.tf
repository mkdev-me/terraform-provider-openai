# These outputs match the ones defined in main.tf
# They're duplicated here for backward compatibility

# Basic outputs for the fine-tuned model
output "id" {
  description = "The ID of the fine-tuned model created"
  value       = length(openai_fine_tuned_model.model) > 0 ? openai_fine_tuned_model.model[0].id : "No model created"
}

output "fine_tuned_model" {
  description = "The fine-tuned model created"
  value       = length(openai_fine_tuned_model.model) > 0 ? openai_fine_tuned_model.model[0].fine_tuned_model : "No model created"
}

output "status" {
  description = "The current status of the fine-tuning job"
  value       = length(openai_fine_tuned_model.model) > 0 ? openai_fine_tuned_model.model[0].status : "No model created"
}

output "created_at" {
  description = "The timestamp when the fine-tuning job was created"
  value       = length(openai_fine_tuned_model.model) > 0 ? openai_fine_tuned_model.model[0].created_at : "No model created"
}

# Commenting out since the model doesn't have fine_tuning_job_id attribute
# output "fine_tuning_job_id" {
#   description = "The ID of the fine-tuning job"
#   value       = openai_fine_tuned_model.model.fine_tuning_job_id
# }

# Commenting out since we commented out the job_details data source
# output "status" {
#   description = "The current status of the fine-tuning job"
#   value       = data.openai_fine_tuning_job.job_details.status
# }

# output "created_at" {
#   description = "The timestamp when the fine-tuning job was created"
#   value       = data.openai_fine_tuning_job.job_details.created_at
# }

# output "finished_at" {
#   description = "The timestamp when the fine-tuning job was completed"
#   value       = data.openai_fine_tuning_job.job_details.finished_at
# }

# output "events" {
#   description = "Events from the fine-tuning job (if monitoring is enabled)"
#   value       = var.enable_monitoring && length(data.openai_fine_tuning_events.job_events) > 0 ? data.openai_fine_tuning_events.job_events[0].events : null
# }

# output "checkpoints" {
#   description = "Checkpoints from the fine-tuning job (if checkpoint access is enabled and job is finished)"
#   value       = local.should_create_checkpoints && length(data.openai_fine_tuning_checkpoints.job_checkpoints) > 0 ? data.openai_fine_tuning_checkpoints.job_checkpoints[0].checkpoints : null
# }

# Commenting out since we commented out the checkpoint permissions resource
/*
output "checkpoint_permissions" {
  description = "Permissions created for sharing checkpoints with other organizations"
  value       = local.shares_count > 0 ? openai_fine_tuning_checkpoint_permission.checkpoint_permissions[*].id : null
}
*/ 