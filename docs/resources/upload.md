# openai_file

This resource allows you to upload and manage files in OpenAI. Files are used for various operations like fine-tuning models, assistants, and vision applications.

## Example Usage

```hcl
# Upload a file for fine-tuning
resource "openai_file" "fine_tune_file" {
  file     = "./training_data.jsonl"
  purpose  = "fine-tune"
}

# Use the uploaded file for fine-tuning
resource "openai_fine_tuned_model" "my_model" {
  training_file = openai_file.fine_tune_file.id
  model         = "gpt-3.5-turbo"
}
```

## Using the Upload Module

For a more flexible approach, you can use the upload module:

```hcl
module "fine_tune_upload" {
  source    = "../../modules/upload"
  purpose   = "fine-tune"
  file_path = "./training_data.jsonl"
}

output "file_id" {
  value = module.fine_tune_upload.file_id
}
```

## Argument Reference

* `file` - (Required) Path to the file to upload.
* `purpose` - (Required) The intended purpose of the uploaded file. Must be one of: "fine-tune", "assistants", or "vision".
* `project_id` - (Optional) The project ID to associate this file with (for Terraform reference only, not sent to OpenAI API).

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the file
* `filename` - The name of the uploaded file
* `bytes` - The size of the file in bytes
* `status` - The current status of the file. Possible values include:
   * `uploaded` - The file has been uploaded but not yet processed
   * `processed` - The file has been processed and is ready for use
   * `error` - There was an error processing the file
* `created_at` - The Unix timestamp when the file was created

## Supported File Types

Different upload purposes require different file types:

* **fine-tune**: JSONL files with prompt-completion pairs
* **assistants**: Various document formats (PDF, DOCX, CSV, etc.)
* **vision**: Image files (PNG, JPEG, GIF, WEBP)

For a complete list of supported file types for each purpose, see the [OpenAI API documentation](https://platform.openai.com/docs/api-reference/files).

## Import

Files can be imported using their ID:

```
$ terraform import openai_file.fine_tune_file file-abc123
``` 