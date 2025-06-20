# OpenAI File Upload Module

This module provides a reusable way to manage file uploads with OpenAI. It supports both creating new files and importing existing ones into Terraform.

## Features

- Upload files to OpenAI with various purposes (fine-tune, assistants, vision, etc.)
- Import existing files into Terraform state
- Lifecycle management to avoid recreating imported files
- Automatic detection of import mode

## Usage

### Creating a New File

```hcl
module "fine_tune_upload" {
  source = "../modules/upload"

  purpose   = "fine-tune"
  file_path = "./training_data.jsonl"
}
```

### Importing an Existing File

1. Define the resource in your Terraform configuration (as above)
2. Import the file using the Terraform CLI:

```bash
terraform import module.fine_tune_upload.openai_file.file file-abc123
```

The module will automatically detect that this is an imported file and handle it appropriately.

## Input Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| purpose | The intended purpose of the uploaded file | `string` | n/a | yes |
| file_path | Path to the file to upload | `string` | `""` | yes for new files, ignored for imports |
| project_id | The ID of the OpenAI project to associate this upload with | `string` | `""` | no |

## Outputs

| Name | Description |
|------|-------------|
| file_id | The unique identifier for the uploaded file |
| filename | The name of the uploaded file |
| bytes | The size of the file in bytes |
| created_at | The timestamp when the upload was created |
| purpose | The purpose of the file |

## Import Mechanism

The module implements a special handling mechanism for imports:

1. It uses a placeholder file path for imported files
2. The lifecycle configuration ignores changes to the file path attribute
3. Import detection works by checking if the file path exists or is the placeholder

This allows seamless management of both new and existing files without requiring recreation.

## Example

```hcl
# Create a new file upload
module "new_file" {
  source = "../modules/upload"

  purpose   = "assistants"
  file_path = "./assistant_data.jsonl"
}

# Reference an imported file (after terraform import)
module "existing_file" {
  source = "../modules/upload"

  purpose   = "fine-tune"
  file_path = "./imported_file.jsonl" # This path will be ignored for imported files
}
```

## Notes

- The file path is required for new file creation but is ignored for imported files
- Changes to other attributes like purpose will require resource recreation
- When a file is imported, its metadata is fetched from the OpenAI API

## Related Resources

- [OpenAI Files API Documentation](https://platform.openai.com/docs/api-reference/files)
- [File Upload Example](../../examples/upload/README.md)
- [Terraform Import Documentation](https://developer.hashicorp.com/terraform/cli/commands/import)

## Supported File Types

Different upload purposes require different file types:

* **fine-tune**: `text/jsonl`