# openai_batch (Data Source)

Retrieves information about an existing OpenAI batch job. This data source allows you to monitor the status of a batch job, check for the availability of results, and handle errors if they occur.

## Example Usage

```hcl
# Reference an existing batch job
data "openai_batch" "existing_batch" {
  batch_id = "batch_abc123"
}

# Check the status of the batch job
output "batch_status" {
  value = data.openai_batch.existing_batch.status
}

# Get the output file ID when available
output "output_file_id" {
  value = data.openai_batch.existing_batch.output_file_id
}
```

### Retrieving a Specific Batch Job by ID

If you have a batch job ID from a previous operation or from outside your Terraform workflow, you can retrieve its information:

```hcl
data "openai_batch" "specific_batch" {
  batch_id = "batch_67dc4576e9cc8190b8a169f65a805996"
}

output "batch_details" {
  value = {
    status        = data.openai_batch.specific_batch.status
    input_file    = data.openai_batch.specific_batch.input_file_id
    output_file   = data.openai_batch.specific_batch.output_file_id
    created_at    = data.openai_batch.specific_batch.created_at
    endpoint      = data.openai_batch.specific_batch.endpoint
    request_count = data.openai_batch.specific_batch.request_counts
  }
}
```

### Monitoring Resources Created by Terraform

You can monitor batch jobs that were created using Terraform by referencing their resource IDs:

```hcl
# Create a batch job
resource "openai_batch" "embeddings_batch" {
  input_file_id     = openai_file.embeddings_file.id
  endpoint          = "/v1/embeddings"
  model             = "text-embedding-ada-002"
  completion_window = "24h"
}

# Monitor the batch job status
data "openai_batch" "monitor_batch" {
  batch_id = openai_batch.embeddings_batch.id
}

# Output that conditionally uses the output file ID when available
locals {
  output_file_id = data.openai_batch.monitor_batch.status == "completed" ? data.openai_batch.monitor_batch.output_file_id : ""
}

output "batch_results_available" {
  value = data.openai_batch.monitor_batch.status == "completed"
}

output "batch_output_file_id" {
  value = local.output_file_id
}
```

### Obtaining Batch Job IDs

You can find your batch job IDs through several methods:

1. From the OpenAI dashboard
2. From a previous Terraform run (stored in your state)
3. By querying the OpenAI API directly:
   ```bash
   curl -X GET https://api.openai.com/v1/batches \
     -H "Authorization: Bearer $OPENAI_API_KEY"
   ```

## Argument Reference

* `batch_id` - (Required) The ID of the batch job to retrieve.
* `project_id` - (Optional) The ID of the OpenAI project associated with the batch job. If not specified, the API key's default project will be used. For project-specific API keys (starting with `sk-proj-`), this parameter is typically not needed.

## Attribute Reference

* `id` - The ID of the batch job.
* `status` - The current status of the batch job. Possible values include:
   * `validating` - The input file is being validated
   * `validation_failed` - Validation found issues with the input file
   * `processing` - The batch is being processed
   * `processing_failed` - Processing encountered errors
   * `completed` - The batch has been successfully processed
* `input_file_id` - The ID of the input file used for the batch.
* `endpoint` - The endpoint used for the batch requests.
* `completion_window` - The time window specified for batch completion.
* `output_file_id` - The ID of the output file (once processing completes).
* `error_file_id` - The ID of the error file (if errors occurred during processing).
* `error` - Information about any error that occurred during processing.
* `created_at` - The Unix timestamp when the batch was created.
* `in_progress_at` - The Unix timestamp when the batch job began processing.
* `expires_at` - The Unix timestamp when the batch expires.
* `completed_at` - The Unix timestamp when the batch job completed.
* `request_counts` - Statistics about request processing, including total, completed, and failed counts.
* `metadata` - Any custom metadata that was attached to the batch job.

## Monitoring Batch Jobs

You can use this data source with Terraform's `depends_on` feature to create a dependency chain that waits for a batch job to complete before proceeding with subsequent resources:

```hcl
# Create a batch job
resource "openai_batch" "embeddings_batch" {
  input_file_id     = openai_file.embeddings_file.id
  endpoint          = "/v1/embeddings"
  model             = "text-embedding-ada-002"
  completion_window = "24h"
}

# Reference the batch job to monitor its status
data "openai_batch" "monitor_batch" {
  batch_id = openai_batch.embeddings_batch.id
  # Force Terraform to check the status on every apply
  depends_on = [openai_batch.embeddings_batch]
}

# Process the results when the batch completes
# This will only succeed when the batch status is "completed"
data "openai_file" "batch_results" {
  file_id = data.openai_batch.monitor_batch.output_file_id
  count   = data.openai_batch.monitor_batch.status == "completed" ? 1 : 0
}
```

## Real-time Monitoring

For real-time monitoring without changing infrastructure, use:

```bash
# Run this command periodically to check progress
terraform refresh

# Or, on Unix systems, automate monitoring with:
watch -n 30 "terraform refresh && terraform output batch_status"
```

## Troubleshooting

If you encounter a 404 Not Found error when retrieving batch information, check:

1. The batch ID is correct and the batch job exists
2. Your API key has appropriate permissions to access the batch
3. The API URL is correctly formatted (no duplicate /v1/ prefixes)

If you receive HTML responses instead of JSON, this typically indicates:
1. An authentication issue with your API key
2. An incorrectly formatted URL
3. A network issue or proxy interference

For more detailed diagnostics, enable debug logging:

```bash
export TF_LOG=DEBUG
terraform apply
```

## Notes

- Batch jobs can take from minutes to hours to complete, so plan your Terraform workflow accordingly.
- For long-running batch jobs, consider using an external monitoring system or script rather than relying on Terraform apply operations.
- If a batch job has already been completed or canceled, you can still retrieve its information using this data source. 