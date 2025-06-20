# OpenAI Provider Examples

This directory contains examples demonstrating how to use the Terraform OpenAI Provider.

## Example Categories

| Category | Description |
|----------|-------------|
| [Upload](./upload/) | Examples for uploading and importing files using the OpenAI API |
| [Files](./files/) | Comprehensive file management examples |
| [Fine Tuning](./fine_tuning/) | Creating and managing fine-tuning jobs |
| [Chat Completion](./chat_completion/) | Working with chat completions |
| [Model Response](./model_response/) | Managing model responses |
| [Audio](./audio/) | Working with audio transcription and text-to-speech |
| [Embeddings](./embeddings/) | Creating and using embeddings |
| [Image](./image/) | Image generation and manipulation |
| [Moderation](./moderation/) | Content moderation examples |
| [Projects](./projects/) | Managing OpenAI projects |
| [Vector Store](./vector_store/) | Working with vector stores |
| [Batch](./batch/) | Batch processing examples |
| [Rate Limit](./rate_limit/) | Rate limiting examples |
| [System API](./system_api/) | Working with system APIs |
| [Service Account](./service_account/) | Managing service accounts |
| [Project API](./project_api/) | Working with project APIs |
| [Project User](./project_user/) | Managing project users |
| [Organization Users](./organization_users/) | Managing organization users |
| [Invite](./invite/) | Handling invites |

## Featured Examples

### File Upload & Import Example

The [upload](./upload/) example demonstrates how to create and import OpenAI files for fine-tuning and other purposes. It showcases:

- Creating new file uploads
- Importing existing files into Terraform
- Managing file lifecycle
- Handling file metadata

This is particularly useful when you need to manage files that were created outside of Terraform or when migrating existing OpenAI resources to Terraform management.

```bash
# Create a new file
cd examples/upload
terraform init
terraform apply

# Import an existing file
terraform import module.fine_tune_upload.openai_file.file file-abc123
```

### Other Key Examples

- **Fine Tuning Jobs**: Learn how to create fine-tuning jobs with custom models
- **Chat Completions**: Examples of chat completion API usage
- **Batch Processing**: Handling large-scale batch operations
- **Project Management**: Managing OpenAI projects and permissions

## Getting Started

Each example directory contains:
- A `README.md` with specific instructions
- Terraform configuration files (`.tf`)
- Sample data files where applicable

To run an example:

1. Navigate to the example directory
2. Set the `OPENAI_API_KEY` environment variable
3. Run `terraform init` and `terraform apply`

## Example Directory Structure

```
examples/
├── upload/           # File upload and import examples
├── files/            # File management examples
├── fine_tuning/      # Fine-tuning examples
└── ...               # Other example categories
```

## Notes

- Examples assume you have Terraform installed and an OpenAI API key
- Some examples require specific permissions on your OpenAI account
- All examples use environment variables for authentication

## Contributing

If you have additional examples or improvements, please submit a pull request!

## Examples

The following examples are available:

### Model

The `model` directory demonstrates how to use the `openai_model` and `openai_models` data sources to retrieve information about OpenAI models. It shows how to:

- Retrieve information about a specific model
- Retrieve a list of all available models
- Use Terraform outputs to display model information

## Invitation Workflow Issues and Solutions

The examples in the `invite/` directory demonstrate how to handle several challenges with the OpenAI invitation process:

### Known Issues

1. **Project Assignment Not Applied**: When sending invitations with project assignments, the OpenAI API **does not actually apply** the project assignments when users accept, even though the API accepts the projects block in the request.

2. **Deletion of Accepted Invitations**: The OpenAI API does not allow deleting invitations that have already been accepted. Attempting to delete an accepted invitation will result in an error.

3. **Finding User IDs After Acceptance**: After a user accepts an invitation, you need their user ID to add them to projects, but this requires an extra lookup step.

### Solutions

The `invite/` example implements several solutions:

1. **Two-Step User Addition Process**:
   - First invite the user to the organization (without relying on project assignments)
   - After they accept, use the `openai_organization_users` data source to find their ID
   - Add them to projects using the `openai_project_user` resource

2. **Handling Accepted Invitation Deletion**:
   - The provider code has been updated to handle "already accepted" errors during deletion
   - Alternatively, you can remove invitations from Terraform state when they're accepted

3. **Automating User ID Lookup**:
   - The example demonstrates using map lookups and `locals` to automatically find user IDs based on email addresses

For a complete workflow implementation, see the [invite example](/examples/invite/).

## API Key Configuration

All examples in this directory require an OpenAI API key to function properly. You have several options for providing this API key:

### 1. Using Environment Variables (Recommended for Local Development)

Set the `OPENAI_API_KEY` environment variable before running terraform commands:

```bash
export OPENAI_API_KEY="your-api-key-here"
terraform apply
```

### 2. Using Provider Configuration (Recommended for CI/CD)

Update the provider block in your configuration to accept an API key:

```hcl
provider "openai" {
  api_key = var.openai_api_key  # If not set, falls back to OPENAI_API_KEY environment variable
}

variable "openai_api_key" {
  description = "OpenAI API Key"
  type        = string
  sensitive   = true
  # Default to environment variable if not explicitly set
  default     = ""
}
```

Then you can pass the API key via command line or a .tfvars file:

```bash
terraform apply -var="openai_api_key=your-api-key-here"
```

Or with a terraform.tfvars file:
```
openai_api_key = "your-api-key-here"
```

### 3. Using Resource-Level API Keys (For Project-Specific Resources)

Some resources and data sources support setting an API key directly on the resource:

```hcl
data "openai_model" "gpt4" {
  model_id = "gpt-4"
  api_key  = var.project_api_key  # Override the provider's default API key
}
```

This is particularly useful when you need to access resources from different projects using different API keys.

### Troubleshooting API Key Issues

If you encounter errors like:
```
Error: Error retrieving model: API error: Incorrect API key provided: ''. You can find your API key at https://platform.openai.com/account/api-keys.
```

This indicates your API key is not being properly recognized. Try these solutions:

1. Verify the environment variable is correctly set: `echo $OPENAI_API_KEY`
2. Set the API key directly in your provider or resource configuration
3. Ensure your API key has the necessary permissions for the resources you're trying to access
