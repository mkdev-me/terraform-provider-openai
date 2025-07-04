---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "openai_batches Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  
---

# openai_batches (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `project_id` (String) The ID of the project associated with the batch jobs. If not specified, the API key's default project will be used.

### Read-Only

- `batches` (List of Object) (see [below for nested schema](#nestedatt--batches))
- `id` (String) The ID of this resource.

<a id="nestedatt--batches"></a>
### Nested Schema for `batches`

Read-Only:

- `completed_at` (Number)
- `completion_window` (String)
- `created_at` (Number)
- `endpoint` (String)
- `error_file_id` (String)
- `expires_at` (Number)
- `id` (String)
- `in_progress_at` (Number)
- `input_file_id` (String)
- `metadata` (Map of String)
- `output_file_id` (String)
- `request_counts` (Map of Number)
- `status` (String)
