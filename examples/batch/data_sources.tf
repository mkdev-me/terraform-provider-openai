# Example usage of openai_batch and openai_batches data sources
# NOTE: These data sources require proper permissions to read batch job information

# Optional: If you want to try the data sources (set to false by default to avoid errors if you don't have batches)
variable "try_batch_data_sources" {
  description = "Whether to try using batch data sources"
  type        = bool
  default     = false
}

# Retrieve a specific batch by ID - using count to make it optional
data "openai_batch" "specific_batch" {
  count    = var.try_batch_data_sources ? 1 : 0
  batch_id = "batch_abc123456789" # Replace with an actual batch ID

  # Optional: Specify the project ID if needed
  # project_id = var.project_id
}

# Retrieve all batches - using count to make it optional
data "openai_batches" "all_batches" {
  count = var.try_batch_data_sources ? 1 : 0

  # Optional: Specify the project ID if needed
  # project_id = var.project_id
}

# Output examples that work regardless of whether data sources succeeded
output "specific_batch_details" {
  description = "Details of a specific batch job (if access granted)"
  value = var.try_batch_data_sources && length(data.openai_batch.specific_batch) > 0 ? {
    id          = data.openai_batch.specific_batch[0].id
    status      = data.openai_batch.specific_batch[0].status
    input_file  = data.openai_batch.specific_batch[0].input_file_id
    output_file = data.openai_batch.specific_batch[0].output_file_id
    created_at  = data.openai_batch.specific_batch[0].created_at
    } : {
    id          = "batch_abc123456789 (example ID)"
    status      = "unknown (enable try_batch_data_sources to fetch actual status)"
    input_file  = "unknown"
    output_file = "unknown"
    created_at  = 0
  }
}

output "all_batch_jobs" {
  description = "List of all batch jobs (if access granted)"
  value = var.try_batch_data_sources && length(data.openai_batches.all_batches) > 0 ? try(
    [for batch in data.openai_batches.all_batches[0].batches : {
      id      = batch.id
      status  = batch.status
      created = batch.created_at
    }],
    ["Error accessing batches data"]
  ) : ["Enable try_batch_data_sources to fetch actual batch jobs"]
}

# Example: Find all completed batch jobs
output "completed_batch_jobs" {
  description = "List of all completed batch jobs (if access granted)"
  value = var.try_batch_data_sources && length(data.openai_batches.all_batches) > 0 ? try(
    [for batch in data.openai_batches.all_batches[0].batches : {
      id          = batch.id
      input_file  = batch.input_file_id
      output_file = batch.output_file_id
    } if batch.status == "completed"],
    ["Error accessing batches data"]
  ) : ["Enable try_batch_data_sources to fetch completed batch jobs"]
} 