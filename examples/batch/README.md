# OpenAI Batch Processing Example

This example demonstrates how to use the OpenAI batch processing API with Terraform, including both resources and data sources.

## Background

OpenAI's API supports batch processing, which allows for efficient processing of multiple requests asynchronously. This is particularly useful for processing large datasets without making thousands of individual API calls.

## Example Implementation

The example shows how to:

1. Create JSONL files with properly formatted batch requests
2. Upload these files to OpenAI with the required `purpose = "batch"` setting 
3. Create batch processing jobs for different endpoints (embeddings and chat completions)
4. Add metadata to batch jobs for better organization
5. Monitor batch job status using the data source
6. Conditionally create resources based on batch job status

## Data Source Usage

This example demonstrates two key patterns with the `openai_batch` data source:

1. **Batch Job Monitoring**: The data source continuously checks the status of a batch job on each apply, providing the latest information about its progress.

2. **Conditional Resource Creation**: Resources that depend on batch completion (like retrieving the output file) can be conditionally created based on the batch job status.

3. **Implicit Dependencies**: When a data source references a resource attribute (like `batch_id = openai_batch.embedding_batch.id`), Terraform automatically creates an implicit dependency. No explicit `depends_on` is needed.

## Input File Format

Each batch job requires a JSONL (JSON Lines) file where each line contains a properly formatted request. The file format for batch requests is specific:

```jsonl
{"custom_id": "request-1", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Explain quantum computing in simple terms"}]}}
```

Key elements:
- `custom_id`: A unique identifier for the request
- `method`: The HTTP method (usually "POST")
- `url`: The API endpoint path (must match the endpoint specified in the batch job)
- `body`: The actual request parameters including the model

## Usage

To run this example:

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-api-key"

# Initialize Terraform
terraform init

# Apply the configuration
terraform apply
```

## Batch Processing Flow

1. **File Preparation**: Create a JSONL file with properly formatted requests
2. **File Upload**: Upload the file to OpenAI with `purpose = "batch"`
3. **Batch Creation**: Create a batch job specifying the input file, endpoint, and model
4. **Job Monitoring**: Monitor the batch job status using the data source
5. **Results Retrieval**: Once completed, retrieve the output file using the `output_file_id`

## Using Metadata

This example demonstrates how to add metadata to batch jobs for easier organization and tracking:

```hcl
resource "openai_batch" "chat_batch_with_metadata" {
  input_file_id     = openai_file.chat_requests.id
  endpoint          = "/v1/chat/completions"
  model             = "gpt-3.5-turbo"
  completion_window = "24h"
  
  metadata = {
    environment = "test"
    purpose     = "demo"
    source      = "terraform-example"
  }
}
```

Metadata can be used to tag your batch jobs with information useful for:
- Environment identification (production, staging, dev)
- Project association
- Versioning
- Purpose or use case tracking

## Resource Outputs

The example produces several outputs from the resources:

- `embedding_batch_id`: ID of the embeddings batch job
- `embedding_batch_status`: Status of the embeddings batch job (validating, processing, completed, etc.)
- `embedding_output_file`: Output file ID for the embeddings job (once completed)
- `chat_batch_id`: ID of the chat completions batch job
- `chat_batch_status`: Status of the chat completions batch job
- `chat_output_file_id`: Output file ID for the chat completions job (once completed)

## Data Source Outputs

Additional outputs from the data source:

- `monitored_batch_status`: Current status of the monitored batch job
- `monitored_batch_request_counts`: Request processing statistics showing progress
- `batch_results_available`: Boolean indicating whether results are ready to retrieve

## Real-time Monitoring

The data source approach enables real-time monitoring of batch jobs:

```bash
# Run this command repeatedly to monitor job progress
terraform refresh

# Or to automate with watch (on Unix systems)
watch -n 30 "terraform refresh"
```

This allows you to monitor long-running batch jobs without manually checking the OpenAI dashboard.

## Supported Endpoints

The batch processing supports the following endpoints:
- `/v1/chat/completions`: For chat-based completions
- `/v1/completions`: For text completions
- `/v1/embeddings`: For generating embeddings

When specifying the endpoint in the `openai_batch` resource, you should use the endpoint path (including the `/v1/` prefix) like this:

```hcl
endpoint = "/v1/embeddings"
```

The provider will automatically construct the full URL correctly without duplicating the `/v1/` prefix.

## Limitations

- Batch jobs can take from minutes to hours to complete
- Input files have a maximum size of 200 MB
- Each file can contain up to 50,000 requests
- For embeddings batches, there's a limit of 50,000 embedding inputs across all requests
- Currently only a 24-hour completion window is supported
- Files with purpose "batch" cannot be deleted until all associated batch jobs are complete

## Troubleshooting

If you encounter errors when using the batch resources:

1. **404 Not Found errors**: Make sure your API key has the correct permissions and the path is correctly formatted.

2. **HTML responses instead of JSON**: This typically indicates an authentication issue or incorrectly constructed URL. The provider has been updated to give more detailed error messages in these cases.

3. **Authentication issues**: Verify your API key is valid and has appropriate permissions.

For more detailed debugging, you can enable Terraform's debug logging:

```bash
export TF_LOG=DEBUG
terraform apply
```

## Notes on Reading Results

Once a batch job completes, you'll need to download the output file using the OpenAI API. This is typically done outside of Terraform with a script that:

1. Retrieves the `output_file_id` from the Terraform state
2. Downloads the file using the OpenAI API
3. Processes the results

For an actual implementation, you might use the output values from Terraform to feed into a script that downloads and processes the batch results. 