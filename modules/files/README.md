# OpenAI File Management Module

This Terraform module provides a simplified interface for uploading and managing files with the OpenAI API.
Files are a fundamental component for many OpenAI features including fine-tuning, assistants, and batch processing.

## Features

- Upload files to OpenAI with proper purpose designation
- Retrieve existing files using the data source mode
- List all files in your OpenAI account with optional filtering
- Track file metadata including size, creation time, and purpose
- Associate files with specific OpenAI projects (for organizational purposes)
- Handle various file types for different OpenAI services

## Usage

### Resource Mode (Creating New Files)

```hcl
module "openai_fine_tune_file" {
  source = "../../modules/files"

  file_path  = "path/to/training_data.jsonl"
  purpose    = "fine-tune"
}

# The file ID can be used in other resources
resource "openai_fine_tuning" "custom_model" {
  training_file = module.openai_fine_tune_file.file_id
  model         = "gpt-3.5-turbo"
}
```

### Data Source Mode (Using Existing Files)

```hcl
module "existing_file" {
  source = "../../modules/files"

  use_data_source = true
  file_id         = "file-abc123def456"
}

# Use the retrieved file's details
output "file_details" {
  value = {
    name        = module.existing_file.filename
    size        = module.existing_file.bytes
    created_at  = module.existing_file.created_at
    purpose     = module.existing_file.purpose
  }
}
```

### List Mode (Retrieving All Files)

```hcl
module "all_files" {
  source = "../../modules/files"
  
  list_files = true
}

output "file_count" {
  value = module.all_files.file_count
}

output "files_by_purpose" {
  value = module.all_files.files_by_purpose
}
```

### List Mode with Filtering

```hcl
module "fine_tune_files" {
  source = "../../modules/files"
  
  list_files = true
  list_files_purpose = "fine-tune"
}

output "fine_tune_file_count" {
  value = module.fine_tune_files.file_count
}
```

## Examples

### Uploading a File for Fine-Tuning

```hcl
module "training_data" {
  source = "../../modules/files"

  file_path = "./data/training_samples.jsonl"
  purpose   = "fine-tune"
}
```

### Retrieving an Existing Assistant File

```hcl
module "knowledge_file" {
  source = "../../modules/files"

  use_data_source = true
  file_id         = "file-abc123def456"
}

# Use the file with an assistant
resource "openai_assistant" "research_assistant" {
  name             = "Research Assistant"
  model            = "gpt-4-turbo"
  file_ids         = [module.knowledge_file.file_id]
  instructions     = "Use the provided document to answer research questions"
}
```

### Uploading a Document for Assistants

```hcl
module "knowledge_base" {
  source = "../../modules/files"

  file_path = "./documents/knowledge_base.pdf"
  purpose   = "assistants"
}
```

### Uploading Batch Processing Requests

```hcl
module "batch_requests" {
  source = "../../modules/files"

  file_path = "./data/embedding_requests.jsonl"
  purpose   = "batch"
}
```

### Listing and Using All Assistant Files

```hcl
module "assistant_files" {
  source = "../../modules/files"
  
  list_files = true
  list_files_purpose = "assistants"
}

# Use the retrieved assistant files with a new assistant
resource "openai_assistant" "knowledge_assistant" {
  name         = "Knowledge Assistant"
  model        = "gpt-4-turbo"
  file_ids     = [for file in module.assistant_files.all_files : file.id]
  instructions = "Use the provided documents to answer questions"
}
```

## Input Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| file_path | Path to the file to upload (resource mode) | `string` | `null` | yes, when use_data_source is false |
| file_id | ID of an existing file to retrieve (data source mode) | `string` | `null` | yes, when use_data_source is true |
| purpose | Purpose of the file (fine-tune, assistants, batch, vision) | `string` | `null` | yes, when use_data_source is false |
| project_id | The ID of the OpenAI project to associate the file with | `string` | `""` | no |
| use_data_source | Whether to use data source mode to retrieve an existing file | `bool` | `false` | no |
| list_files | Whether to list all files | `bool` | `false` | no |
| list_files_purpose | Purpose filter when listing files | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| file_id | The ID of the file (resource/data source mode) |
| filename | The name of the file (resource/data source mode) |
| bytes | The size of the file in bytes (resource/data source mode) |
| created_at | Timestamp when the file was created (resource/data source mode) |
| purpose | The purpose of the file (resource/data source mode) |
| all_files | List of all files (list mode) |
| file_count | Number of files retrieved (list mode) |
| files_by_purpose | Files grouped by purpose (list mode) |

## Supported File Purposes

- `fine-tune`: For training and fine-tuning AI models
- `assistants`: For use with the Assistants API to provide knowledge bases
- `batch`: For batch processing of requests
- `vision`: For images used in vision fine-tuning

## Module Modes

This module supports three operational modes:

1. **Resource Mode** (default): Creates and uploads a new file to OpenAI
2. **Data Source Mode**: Retrieves an existing file from OpenAI
3. **List Mode**: Lists all files in your OpenAI account, with optional filtering

To switch between modes, use the appropriate variables:
- Resource Mode: Default mode, set `use_data_source=false` (requires `file_path` and `purpose`)
- Data Source Mode: Set `use_data_source=true` (requires `file_id`)
- List Mode: Set `list_files=true`, optionally set `list_files_purpose` for filtering

You can combine List Mode with either Resource or Data Source mode to perform multiple operations.

## Importing Existing Files

You can import files that already exist in your OpenAI account into Terraform state for management:

```bash
# The import syntax for modules requires the module path plus the internal resource path
terraform import module.knowledge_file.openai_file.this[0] file-abc123def456
```

The import process:

1. Retrieves complete file metadata from the OpenAI API
2. Sets a reasonable default for the `file` attribute based on the filename
3. Populates all attributes like `filename`, `bytes`, `created_at`, and `purpose` from the API

### Import Workflow Example

```bash
# First, define the module in your configuration
module "knowledge_file" {
  source = "../../modules/files"

  # Placeholder values that will be overridden by the import
  file_path = "./some/placeholder.pdf"
  purpose   = "assistants"
}

# Then import the existing file
terraform import module.knowledge_file.openai_file.this[0] file-abc123def456

# Verify the import by checking the state
terraform state show module.knowledge_file.openai_file.this[0]
```

After importing, update your configuration to match the imported state:

```hcl
module "knowledge_file" {
  source = "../../modules/files"

  file_path = "./data/knowledge_base.pdf"  # Should match what was imported
  purpose   = "assistants"                 # Should match what was imported
}
```

### Data Source Mode vs. Import

While both data source mode and import allow you to work with existing files:

- **Data Source Mode**: Reads file metadata for reference without managing the resource
- **Import**: Brings the file under Terraform management for its lifecycle

Choose import when you want Terraform to manage the file's lifecycle (especially deletion).

## Notes

- File sizes and formats are limited based on the purpose. Refer to OpenAI's documentation for the latest limits.
- Files may take time to process depending on their size and purpose.
- The file must be accessible to Terraform during the apply operation when using resource mode.
- You need appropriate permissions to access files when using data source mode. 
