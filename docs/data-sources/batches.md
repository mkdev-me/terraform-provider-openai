---
page_title: "OpenAI: openai_batches"
subcategory: "Batches"
description: |-
  Retrieves a list of all batch jobs for an OpenAI project.
---

# Data Source: openai_batches

Use this data source to retrieve a list of all batch jobs for a specific OpenAI project.

## Example Usage

```hcl
data "openai_batches" "project_batches" {
  project_id = "proj_abc123456789"
  
  # Optional: use a project-specific API key
  api_key    = var.project_api_key
}

output "total_batch_jobs" {
  value = length(data.openai_batches.project_batches.batches)
}

# Find all completed batch jobs
output "completed_batches" {
  value = [
    for batch in data.openai_batches.project_batches.batches :
    batch if batch.status == "completed"
  ]
}

# Get all batch job IDs
output "all_batch_ids" {
  value = [
    for batch in data.openai_batches.project_batches.batches :
    batch.id
  ]
}
```

## Argument Reference

* `project_id` - (Optional) The ID of the project to retrieve batch jobs for (format: `proj_abc123456789`). If not specified, the API key's default project will be used.
* `api_key` - (Optional) A project-specific API key to use for authentication. If not provided, the provider's default API key will be used.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `batches` - A list of batch jobs for the project. Each batch job contains:
  * `id` - The unique identifier for the batch job.
  * `input_file_id` - The ID of the input file used for the batch.
  * `endpoint` - The endpoint used for the batch requests.
  * `completion_window` - The time window specified for batch completion.
  * `output_file_id` - The ID of the output file (if processing has completed).
  * `error_file_id` - The ID of the error file (if errors occurred during processing).
  * `status` - The current status of the batch job.
  * `created_at` - The Unix timestamp when the batch job was created.
  * `in_progress_at` - The Unix timestamp when the batch job began processing.
  * `expires_at` - The Unix timestamp when the batch job expires.
  * `completed_at` - The Unix timestamp when the batch job completed.
  * `request_counts` - Statistics about request processing, including total, completed, and failed counts.
  * `metadata` - Any custom metadata that was attached to the batch job.

## Permissions Required

This data source requires an OpenAI API key with appropriate permissions to list batch jobs for the project. 