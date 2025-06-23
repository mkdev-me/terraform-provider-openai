# List all fine-tuning jobs
data "openai_fine_tuning_jobs" "all" {
  # Optional: Limit the number of jobs returned
  # limit = 20
}

# Output total job count
output "total_jobs" {
  value = length(data.openai_fine_tuning_jobs.all.jobs)
}

