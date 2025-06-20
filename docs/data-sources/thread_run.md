# openai_thread_run Data Source

Retrieves information about an existing OpenAI thread run. This data source allows you to access details about a run that was initiated on a thread, including its status, metadata, and usage statistics.

## Example Usage

```hcl
data "openai_thread_run" "example" {
  thread_id = "thread_abc123"
  run_id    = "run_xyz789"
}

output "run_status" {
  value = data.openai_thread_run.example.status
}

output "run_usage" {
  value = data.openai_thread_run.example.usage
}
```

## Argument Reference

* `thread_id` - (Required) The ID of the thread that the run belongs to.
* `run_id` - (Required) The ID of the run to retrieve.

## Attribute Reference

* `id` - The identifier of the run.
* `thread_id` - The ID of the thread that the run belongs to.
* `assistant_id` - The ID of the assistant used for the run.
* `model` - The ID of the model used for the run.
* `instructions` - The instructions provided for the run, if any were specified.
* `status` - The status of the run, which can be: queued, in_progress, requires_action, cancelling, cancelled, failed, completed, or expired.
* `tools` - The tools that were available to the assistant during the run.
* `created_at` - The Unix timestamp (in seconds) of when the run was created.
* `started_at` - The Unix timestamp (in seconds) of when the run was started.
* `completed_at` - The Unix timestamp (in seconds) of when the run was completed, if applicable.
* `object` - The object type, which is always 'thread.run'.
* `file_ids` - A list of file IDs that the run had access to.
* `metadata` - Set of key-value pairs that were attached to the run.
* `usage` - Usage statistics for the run, including prompt_tokens, completion_tokens, and total_tokens.

## Usage with Resources

You can use this data source to retrieve information about runs that were created using the `openai_thread_run` resource:

```hcl
resource "openai_thread_run" "example" {
  assistant_id = openai_assistant.example.id
  
  thread {
    messages {
      role    = "user"
      content = "What are the key benefits of infrastructure as code?"
    }
  }
}

data "openai_thread_run" "details" {
  thread_id = openai_thread_run.example.thread_id
  run_id    = openai_thread_run.example.id
  
  depends_on = [openai_thread_run.example]
}

output "run_details" {
  value = {
    status       = data.openai_thread_run.details.status
    started_at   = data.openai_thread_run.details.started_at
    completed_at = data.openai_thread_run.details.completed_at
    usage        = data.openai_thread_run.details.usage
  }
}
```

## Notes

* This data source is useful for retrieving the current status and details of a run, especially when you need to check if a run has completed or to get usage statistics.
* The run must exist in the OpenAI API to be retrieved. If it has been deleted, the data source will return an error.
* Run statuses may change over time. For example, a run might be in the "in_progress" status when first queried, but later move to "completed". 