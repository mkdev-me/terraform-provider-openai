# OpenAI File Management Examples

This example demonstrates how to use the OpenAI Terraform provider to manage files in your OpenAI account. Files are a fundamental component for many OpenAI features including fine-tuning, assistants, and batch processing.

## Features Demonstrated

This example showcases:

1. **Creating Files**: Uploading new files to OpenAI with various purposes
2. **Retrieving Files**: Accessing existing files using data source mode
3. **Multiple File Types**: Working with different file types for various purposes
4. **Module Patterns**: Using both the direct resource and the abstracted module

## Prerequisites

- OpenAI API key with appropriate permissions
- Terraform installed
- Basic understanding of the OpenAI API

## Usage

### Setting Up

1. Set your OpenAI API key as an environment variable:
   ```
   export OPENAI_API_KEY=your_api_key_here
   ```

2. Initialize Terraform:
   ```
   terraform init
   ```

3. Apply the configuration:
   ```
   terraform apply
   ```

## Example Overview

### 1. Fine-tuning Files

This example creates a JSONL file with prompt-completion pairs and uploads it for fine-tuning purposes.

```hcl
resource "openai_file" "fine_tune_file" {
  file    = local_file.fine_tune_sample.filename
  purpose = "fine-tune"
}
```

### 2. Batch Processing Files

Creates and uploads a file for batch processing, typically used for embedding or completion requests.

```hcl
resource "openai_file" "batch_file" {
  file    = local_file.batch_sample.filename
  purpose = "batch"
}
```

### 3. Using the Files Module

Demonstrates the abstracted module approach for fine-tuning files.

```hcl
module "training_file" {
  source = "../../modules/files"

  file_path = local_file.fine_tune_sample.filename
  purpose   = "fine-tune"
}
```

### 4. Assistant Files

Shows how to upload a text file for use with the Assistants API.

```hcl
resource "openai_file" "assistants_file" {
  file    = local_file.assistants_sample.filename
  purpose = "assistants"
}
```

### 5. Retrieving Existing Files

Demonstrates using the files module in data source mode to retrieve information about existing files.

```hcl
module "existing_file" {
  source = "../../modules/files"

  use_data_source = true
  file_id         = openai_file.assistants_file.id
}
```

### 6. Using a Custom API Key

Shows how to override the provider's default API key for specific file operations.

```hcl
module "custom_api_key_file" {
  source = "../../modules/files"

  file_path = "path/to/file.txt"
  purpose   = "assistants"
}
```

## Importing Existing Files

You can import files that already exist in your OpenAI account into Terraform state:

```bash
# Import an existing file
terraform import openai_file.assistants_file file-abc123defg

# Check the imported state
terraform state show openai_file.assistants_file
```

The import process:

1. Retrieves the file's metadata from the OpenAI API
2. Sets a reasonable default for the `file` attribute based on the filename
3. Updates the Terraform state with all retrieved properties

After importing, you may need to update your configuration to match the imported state:

```hcl
resource "openai_file" "assistants_file" {
  file    = "./data/your_document.pdf"  # Should match the path Terraform set
  purpose = "assistants"                # Will match what's in the API
}
```

### Importing Files for Modules

If you're using the files module, the import command is slightly different:

```bash
terraform import module.existing_file.openai_file.this[0] file-abc123defg
```

### Import Workflow Example

```bash
# First, remove the resource from state if it already exists
terraform state rm openai_file.assistants_file

# Import the file by its ID
terraform import openai_file.assistants_file file-HrT9CHDFutQWhwkGoZFqH1

# Verify the import was successful
terraform plan
```

If the `plan` shows no changes needed, the import was successful. If it shows changes, you may need to adjust your Terraform configuration to match the imported state.

## File Purposes

OpenAI supports different file purposes:

- `fine-tune`: For training and fine-tuning AI models
- `assistants`: For use with the Assistants API
- `batch`: For batch processing operations
- `vision`: For vision fine-tuning
- `user_data`: Flexible file type
- `evals`: For evaluation data sets

## Module Modes

The files module supports two modes:

1. **Resource Mode** (default): Creates new files
   - Requires `file_path` and `purpose`
   - Creates and uploads the file
   - Optionally accepts a custom API key via the `api_key` parameter

2. **Data Source Mode**: Retrieves existing files
   - Requires `use_data_source = true` and `file_id`
   - Retrieves information about an existing file
   - Optionally accepts a custom API key via the `api_key` parameter

## Outputs

The example includes outputs for:

- File IDs for all created files
- Detailed file information (filename, size, etc.)
- Retrieved file details when using data source mode

## Notes

- File uploads may take time depending on size
- Some file operations may require specific API permissions
- Different purposes have different supported file formats
- The data source mode allows you to work with files that were created outside of Terraform

## File Format Requirements

OpenAI requires different file formats depending on the purpose:

| Purpose | Required Format | Notes |
|---------|----------------|-------|
| `fine-tune` | `.jsonl` | Each line must be a valid JSON object with "prompt" and "completion" fields |
| `assistants` | Various formats | Supported formats: "c", "cpp", "css", "csv", "doc", "docx", "gif", "go", "html", "java", "jpeg", "jpg", "js", "json", "md", "pdf", "php", "pkl", "png", "pptx", "py", "rb", "tar", "tex", "ts", "txt", "webp", "xlsx", "xml", "zip" |
| `batch` | `.jsonl` | Each line must be a valid API request in JSON format |
| `vision` | Image formats | Supported formats include "jpg", "jpeg", "png", "webp" |

**Important**: Using the wrong file format for a purpose will result in an API error. For example, a `.jsonl` file cannot be used with the "assistants" purpose, and a `.txt` file cannot be used with the "fine-tune" purpose.

## Example Files Created

The configuration automatically generates sample files in the `data/` directory:

- `fine_tune_sample.jsonl` - A minimal sample for fine-tuning with prompt-completion pairs
- `batch_sample.jsonl` - A sample file for batch embedding generation

## Output Information

After applying the configuration, you'll see outputs including:

- File IDs for all uploaded files
- Detailed information about the referenced file, including:
  - Original filename
  - File size
  - Creation timestamp
  - Purpose

## Cleaning Up

To delete all created resources including the uploaded files:

```bash
terraform destroy
```

## Common Issues

- **File Processing Time**: Files may take time to process after uploading. Use `terraform refresh` to get the latest file information.
- **File Size Limits**: OpenAI has varying size limits based on file purpose. Check their documentation for current limits.
- **Permission Issues**: Ensure your API key has appropriate permissions for file operations.
- **File Retention**: Note that files uploaded for certain purposes may be automatically deleted after a period of time.

## Next Steps

- Use the uploaded file IDs in other OpenAI resources like fine-tuning jobs or assistants
- Explore complex file workflows by combining these examples with other OpenAI provider resources 
