# Fetch a specific fine-tuning job by ID
data "openai_fine_tuning_job" "custom_model" {
  fine_tuning_job_id = "ftjob-1bkb8XBdqNaTI1p5MAq6JxVV"
}

# Output fine-tuning job status
output "job_status" {
  value = data.openai_fine_tuning_job.custom_model.status
}
