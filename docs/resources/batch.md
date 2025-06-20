# openai_batch

This resource allows you to create and manage batch processing jobs in OpenAI. Batch processing enables you to process multiple requests asynchronously, which is ideal for large-scale operations like generating embeddings or completions for datasets.

## Example Usage

```hcl
# Upload a file with batch requests
resource "openai_file" "embeddings_file" {
  file    = "embeddings_requests.jsonl"
  purpose = "batch"
}

# Create a batch job for processing embeddings
resource "openai_batch" "embeddings_batch" {
  input_file_id     = openai_file.embeddings_file.id
  endpoint          = "/v1/embeddings"
  model             = "text-embedding-ada-002"
  completion_window = "24h"
}

# Output the batch job details
output "batch_id" {
  value = openai_batch.embeddings_batch.id
}

output "batch_status" {
  value = openai_batch.embeddings_batch.status
}

output "output_file_id" {
  value = openai_batch.embeddings_batch.output_file_id
}
```

## With Metadata

You can attach metadata to your batch jobs for better organization:

```hcl
resource "openai_batch" "embeddings_batch_with_metadata" {
  input_file_id     = openai_file.embeddings_file.id
  endpoint          = "/v1/embeddings"
  model             = "text-embedding-ada-002"
  completion_window = "24h"
  
  metadata = {
    environment = "production"
    project     = "document-search"
    version     = "1.0"
  }
}
```

## Input File Format

The input file for batch processing must be a JSONL (JSON Lines) file with each line containing a properly formatted request. The exact format depends on the endpoint being used.

### Correct Format for Batch Requests

Each line in the JSONL file must include the following fields:

- `custom_id`: A unique identifier for each request
- `method`: The HTTP method (usually "POST")
- `url`: The API endpoint path (e.g., "/v1/chat/completions" or "/v1/embeddings")
- `body`: The actual request parameters

**Example for Chat Completions:**

```jsonl
{"custom_id": "request-1", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Explain quantum computing in simple terms"}]}}
{"custom_id": "request-2", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Write a short poem about artificial intelligence"}]}}
{"custom_id": "request-3", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "How can I improve my time management skills?"}]}}
```

**Example for Embeddings:**

```jsonl
{"custom_id": "embed-1", "method": "POST", "url": "/v1/embeddings", "body": {"model": "text-embedding-ada-002", "input": "The food was delicious and the service was excellent."}}
{"custom_id": "embed-2", "method": "POST", "url": "/v1/embeddings", "body": {"model": "text-embedding-ada-002", "input": "I had a terrible experience at the restaurant."}}
{"custom_id": "embed-3", "method": "POST", "url": "/v1/embeddings", "body": {"model": "text-embedding-ada-002", "input": "The price was reasonable for the quality of the meal."}}
```

### Common Formatting Errors

Incorrect formats that will **not** work:

1. Missing required fields:
```jsonl
{"model": "gpt-3.5-turbo", "custom_id": "request-1", "messages": [{"role": "user", "content": "Explain quantum computing in simple terms"}]}
```

2. Missing method field:
```jsonl
{"model": "gpt-3.5-turbo", "custom_id": "request-1", "messages": [{"role": "user", "content": "Explain quantum computing in simple terms"}]}
```

3. Not wrapping request body in a "body" field:
```jsonl
{"custom_id": "request-1", "method": "POST", "url": "/v1/chat/completions", "model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Explain quantum computing in simple terms"}]}
```

The OpenAI API requires the specific format shown in the correct examples above for batch processing to work properly.

## Argument Reference

* `input_file_id` - (Required) The ID of the file containing the batch requests. The file must have been uploaded with `purpose = "batch"`.
* `endpoint` - (Required) The endpoint to use for all requests in the batch. Must be one of: `/v1/chat/completions`, `/v1/completions`, `/v1/embeddings`, or another supported batch endpoint.
* `model` - (Required) The model to use for the batch. This is stored in the Terraform state but not sent to the OpenAI API. Each request in your JSONL file must include the `model` parameter.
* `completion_window` - (Optional) The time window within which the batch should be processed. Currently, only `"24h"` is supported. Defaults to `"24h"`.
* `project_id` - (Optional) The ID of the OpenAI project to use for this batch. If not specified, the API key's default project will be used. For project-specific API keys (starting with `sk-proj-`), this parameter is typically not needed.
* `metadata` - (Optional) A map of key-value pairs that can be attached to the batch object. Limited to 16 pairs maximum.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the batch job
* `status` - The current status of the batch job. Possible values include:
   * `validating` - The input file is being validated
   * `validation_failed` - Validation found issues with the input file
   * `processing` - The batch is being processed
   * `processing_failed` - Processing encountered errors
   * `completed` - The batch has been successfully processed
* `created_at` - The Unix timestamp when the batch was created
* `expires_at` - The Unix timestamp when the batch expires
* `output_file_id` - The ID of the output file (once processing completes)
* `error` - Information about any error that occurred during processing

## API Keys and Projects

For batch processing, you have two options:

1. **Standard API Key**: Use a regular API key (starting with `sk-`) and specify a `project_id` if you want to use a specific project.
2. **Project-Specific API Key**: Use a project-specific API key (starting with `sk-proj-`), which is automatically associated with a specific project. In this case, you don't need to specify a `project_id`.

Project-specific API keys are recommended for production use to ensure resource isolation and proper billing.

## Output Files

When a batch job completes successfully, the `output_file_id` attribute will contain the ID of the file with the results. You can retrieve the contents of this file using the OpenAI API or by using the `openai_file` data source.

The output file is a JSONL file where each line corresponds to the result of processing each line in your input file.

## Timeouts and Limitations

- Batch jobs can take from minutes to hours to complete, depending on the size of your input file and the current load on OpenAI's systems.
- The maximum input file size is 200 MB, containing up to 50,000 requests.
- For embeddings batches, there's a limit of 50,000 embedding inputs across all requests.
- Currently, only a 24-hour completion window is supported.
- Files uploaded with `purpose = "batch"` cannot be deleted until all associated batch jobs are complete or canceled.

## Import

Batch jobs can be imported using their ID:

```bash
terraform import openai_batch.embeddings_batch batch_abc123
```

The import process:

1. Retrieves all details about the batch job directly from the OpenAI API
2. Sets the required attributes (`input_file_id`, `endpoint`, `completion_window`) from the API response
3. Populates computed attributes like `status`, `created_at`, `expires_at`, `output_file_id`, etc.
4. Preserves any metadata attached to the batch job

### Import Example

```bash
# First, remove the resource from Terraform state if it exists
terraform state rm openai_batch.my_batch

# Then import it using the batch ID
terraform import openai_batch.my_batch batch_67dc4576e9cc8190b8a169f65a805996
```

After importing, your Terraform configuration should include at minimum:

```hcl
resource "openai_batch" "my_batch" {
  input_file_id     = "file-abc123"            # From the imported state
  endpoint          = "/v1/chat/completions"   # From the imported state
  completion_window = "24h"                    # From the imported state
}
```

### Import Considerations

1. After importing, run `terraform state show openai_batch.my_batch` to see all the imported attributes
2. Update your configuration to match the imported state to prevent Terraform from making unwanted changes
3. The input file resource must be imported separately if you want to manage it through Terraform

Note that importing the batch job does not create any input or output file resources in Terraform. If you need to manage these files as well, you'll need to import them separately using `openai_file`.

## Troubleshooting

If you encounter a 404 Not Found error when creating a batch job, make sure:

1. Your API key has the correct permissions
2. The API endpoint URL is correctly formed (don't add duplicate /v1/ prefixes)
3. The endpoint path in your request matches what's expected by the API

If you receive HTML responses instead of JSON, this typically indicates:
1. An authentication issue with your API key
2. An incorrectly formatted URL
3. A network issue or proxy that's interfering with the request

Enable debug logging for more detailed error information:

```
export TF_LOG=DEBUG
terraform apply
```