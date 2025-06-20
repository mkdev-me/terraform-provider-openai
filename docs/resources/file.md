# openai_file

Provides a resource to upload a file to OpenAI's API.

## Example Usage

```hcl
resource "openai_file" "my_file" {
  file    = "./training-data.jsonl"
  purpose = "fine-tune"
}
```

## Argument Reference

* `file` - (Required) Path to the file to be uploaded.
* `purpose` - (Required) Purpose of the file. Supported purpose types are `fine-tune`, `assistants`, and `assistants_output`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier for this file.
* `filename` - Original filename as provided when uploaded.
* `bytes` - Size of this file in bytes.
* `created_at` - Timestamp for when this file was created.
* `status` - Status of the file (`uploaded`, `processed`, or `error`).
* `status_details` - Additional details about the status of the file (e.g., error messages).
* `object` - Type of object (always `file`).

## Import

OpenAI files can be imported using the file ID, e.g.,

```bash
terraform import openai_file.my_file file-abc123
```

The import process:

1. Retrieves the file metadata directly from the OpenAI API
2. Sets the `file` attribute to a best-guess path based on the filename (typically `./data/filename.ext`)
3. Populates all computed attributes like `filename`, `bytes`, `created_at`, and `purpose` from the API

### Import Example

```bash
# First, remove the resource from Terraform state if it exists
terraform state rm openai_file.assistants_file

# Then import it using the file ID
terraform import openai_file.assistants_file file-HrT9CHDFutQWhwkGoZFqH1
```

After importing, you may need to align your configuration with the imported state to prevent unwanted changes:

```hcl
resource "openai_file" "assistants_file" {
  file    = "./data/assistants_sample.txt" # Should match the path Terraform imported
  purpose = "assistants"                   # Will be set automatically from the API
}
```

### Import with Modules

If using a file module:

```bash
terraform import module.training_file.openai_file.this[0] file-abc123
```

### Import Limitations

* You cannot modify the content of an imported file
* Changing attributes like `purpose` will require recreation of the resource

## File Formats and Limitations

The supported file formats depend on the purpose:

* For `fine-tune`: JSONL files where each line is a valid JSON object with `prompt` and `completion` fields
* For `assistants`: PDF, docx, txt, csv, json, etc.
* For `batch`: JSONL files containing valid API requests
* For `vision`: Image files in formats supported by OpenAI's vision models

## Best Practices

1. For fine-tuning files, ensure your JSONL is properly formatted according to OpenAI's specifications
2. Files uploaded for assistants should be relevant to the assistant's purpose and optimized for understanding
3. For batch processing, each line in your JSONL file should be a complete and valid API request
4. Consider file size limitations: OpenAI typically limits file sizes based on purpose 
