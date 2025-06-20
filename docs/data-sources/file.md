# openai_file (Data Source)

Retrieves information about an existing OpenAI file. This data source allows you to access metadata about files that have been uploaded to OpenAI, such as their status, purpose, and other attributes.

## Example Usage

```hcl
# Reference an existing file by ID
data "openai_file" "existing_file" {
  file_id = "file_abc123def456"
}

# Output file details
output "file_name" {
  value = data.openai_file.existing_file.filename
}

output "file_status" {
  value = data.openai_file.existing_file.status
}

output "file_purpose" {
  value = data.openai_file.existing_file.purpose
}
```

## Argument Reference

* `file_id` - (Required) The ID of the file to retrieve.
* `project_id` - (Optional) The ID of the OpenAI project associated with the file. This is for Terraform reference only and is not sent to the OpenAI API.

## Attribute Reference

* `id` - The ID of the file (same as `file_id`).
* `filename` - The original name of the file.
* `bytes` - The size of the file in bytes.
* `created_at` - The Unix timestamp when the file was created.
* `purpose` - The intended use of the file (e.g., `fine-tune`, `assistants`, `batch`).
* `status` - The current status of the file (e.g., `processed`, `processing`).
* `status_details` - Additional details about the status of the file, if available.

## Using with Other Resources

The file data source is commonly used to reference existing files for use with other OpenAI resources:

```hcl
# Reference an existing file
data "openai_file" "training_data" {
  file_id = "file_abc123def456"
}

# Use the file for fine-tuning
resource "openai_fine_tuning" "custom_model" {
  training_file = data.openai_file.training_data.id
  model         = "gpt-3.5-turbo"
}
```

## Note

Files have different access controls depending on their purpose and the API key used to upload them. Make sure you're using an API key with appropriate permissions to access the file. 