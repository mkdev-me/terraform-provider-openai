# OpenAI File Upload and Import Example

This example demonstrates how to use the OpenAI Provider to upload files for fine-tuning and how to import existing files into Terraform.

## Overview

The example shows how to:
1. Create a new file upload resource for fine-tuning
2. Import existing OpenAI files into Terraform state
3. Manage the lifecycle of file resources properly

## Requirements

- Terraform 1.0.0 or higher
- OpenAI Provider (fjcorp/openai) 1.0.0 or higher
- OpenAI API key with appropriate permissions (set via OPENAI_API_KEY environment variable)

## Usage

### Creating New Files

To create a new file upload:

```bash
# Initialize Terraform
terraform init

# Plan the deployment
terraform plan

# Apply the configuration to create the file
terraform apply
```

### Importing Existing Files

To import an existing file that was created outside of Terraform:

```bash
# Import the file using its ID (replace with your actual file ID)
terraform import module.fine_tune_upload.openai_file.file file-abc123
```

The import process will:
1. Set a placeholder for the local file path
2. Retrieve all other file properties from the OpenAI API
3. Add the file to Terraform state for management

## Implementation Details

The example uses the `modules/upload` module which:

- Handles both creation of new files and importing existing ones
- Uses a lifecycle configuration to ignore changes to the file path for imported files
- Detects import mode automatically and adjusts behavior accordingly

## Configuration

```hcl
module "fine_tune_upload" {
  source = "../../modules/upload"

  purpose   = "fine-tune"  # Purpose of the file (fine-tune, assistants, etc.)
  file_path = "./training_data.jsonl"  # Path to local file for upload
}
```

## Outputs

- `file_id` - The ID of the uploaded file
- `filename` - The filename of the uploaded file
- `bytes` - The size of the file in bytes
- `created_at` - The timestamp when the file was created

## Notes

- Files uploaded for fine-tuning must be in JSONL format
- The OpenAI API requires appropriate permissions to upload files
- You can use imported files for fine-tuning jobs just like newly created files
- Changes to the `file_path` attribute will create a new file rather than updating the existing one

## Related Resources

- [OpenAI Files API Documentation](https://platform.openai.com/docs/api-reference/files)
- [OpenAI File Upload Module](../../modules/upload/README.md)
- [Terraform Import Documentation](https://developer.hashicorp.com/terraform/cli/commands/import)

## Prerequisites

- An OpenAI API key
- Terraform installed
- The OpenAI Terraform Provider installed
- jq installed (for file processing)

## File Structure

- `main.tf` - The main Terraform configuration file
- `training_data.jsonl` - Sample training data for fine-tuning

## Usage

1. Set your OpenAI API key:
   ```
   export OPENAI_API_KEY=your_api_key_here
   ```

2. Initialize Terraform:
   ```
   terraform init
   ```

3. Apply the Terraform configuration:
   ```
   terraform apply
   ```

4. After the file is uploaded, you can get its ID:
   ```
   terraform output file_id
   ```

## File Upload Process

The OpenAI file upload process is handled by our module which:

1. Uses curl to upload the file to the OpenAI API
2. Captures the response with the file ID and metadata
3. Makes this information available as Terraform outputs

## Supported File Types and Purposes

Different upload purposes require different file types:

- **fine-tune**: JSONL files with prompt-completion pairs
- **assistants**: Various document formats (PDF, DOCX, CSV, etc.)
- **vision**: Image files (PNG, JPEG, GIF, WEBP)

## Using the Uploaded File

Once you've uploaded a file, you can use its ID for other operations:

```hcl
# Use the uploaded file for fine-tuning
resource "openai_fine_tuned_model" "my_model" {
  training_file = module.fine_tune_upload.file_id
  model         = "gpt-3.5-turbo"
}
```

## Clean Up

To destroy the resources created by Terraform:
```
terraform destroy
```

Note that this will not delete the file from OpenAI. To delete the file, you need to call the OpenAI API:

```
curl -X DELETE https://api.openai.com/v1/files/$(terraform output -raw file_id) \
  -H "Authorization: Bearer $OPENAI_API_KEY"
``` 