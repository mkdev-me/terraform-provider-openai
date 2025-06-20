# OpenAI Files Data Source

The `openai_files` data source allows you to retrieve a list of files available in your OpenAI account, with optional filtering by purpose.

## Example Usage

```hcl
# Retrieve all files
data "openai_files" "all" {}

# List files with a specific purpose
data "openai_files" "fine_tune_files" {
  purpose = "fine-tune"
}

# Access file details
output "file_count" {
  value = length(data.openai_files.all.files)
}

output "file_names" {
  value = [for file in data.openai_files.all.files : file.filename]
}
```

### Using with Projects

For organizational purposes, you can associate the data source with a project ID (this is only for Terraform state reference and not used in the OpenAI API call):

```hcl
data "openai_files" "project_files" {
  project_id = "proj_123"
}
```

## Filtering

You can filter files by their purpose:

```hcl
data "openai_files" "assistant_files" {
  purpose = "assistants"
}

data "openai_files" "batch_files" {
  purpose = "batch"
}
```

## Grouping Files by Purpose

```hcl
data "openai_files" "all" {}

output "files_by_purpose" {
  value = {
    for file in data.openai_files.all.files :
    file.purpose => file.filename...
  }
}
```

## Using in Assistants

```hcl
data "openai_files" "assistant_files" {
  purpose = "assistants"
}

resource "openai_assistant" "my_assistant" {
  name       = "Knowledge Assistant"
  model      = "gpt-4-turbo"
  file_ids   = [for file in data.openai_files.assistant_files.files : file.id]
}
```

## Argument Reference

* `purpose` - (Optional) Filter files by purpose. Can be "fine-tune", "assistants", "batch", etc.
* `project_id` - (Optional) The project ID to associate with this file lookup (for Terraform reference only, not sent to OpenAI API).

## Attribute Reference

* `files` - A list of files with the following attributes:
  * `id` - The ID of the file.
  * `filename` - The name of the file.
  * `bytes` - The size of the file in bytes.
  * `created_at` - A timestamp of when the file was created, formatted as an RFC3339 string.
  * `purpose` - The purpose of the file (e.g., "fine-tune", "assistants").
  * `object` - The object type, which is always "file".

## Related Resources

* [`openai_file` Resource](../resources/file.md)
* [`openai_file` Data Source](../data-sources/file.md)

## Notes

* Files are specific to your OpenAI account or organization.
* File filtering is performed by the OpenAI API, not client-side. 