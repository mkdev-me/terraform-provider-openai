# openai_run Data Source

Retrieves information about an existing run in the OpenAI Assistants API. This data source allows you to access details about a run that was executed on a thread, including its status, results, and usage statistics.

## Example Usage

```hcl
data "openai_run" "example" {
  thread_id = "thread_abc123"
  run_id    = "run_xyz789"
}

output "run_status" {
  value = data.openai_run.example.status
}

output "run_usage" {
  value = data.openai_run.example.usage
}

output "run_completed_at" {
  value = data.openai_run.example.completed_at
}
```

## Using with Resources

This example shows how to retrieve details about a run that was created with the `openai_run` resource:

```hcl
resource "openai_run" "my_run" {
  thread_id    = openai_thread.my_thread.id
  assistant_id = openai_assistant.my_assistant.id
}

data "openai_run" "details" {
  thread_id = openai_run.my_run.thread_id
  run_id    = openai_run.my_run.id
  
  depends_on = [openai_run.my_run]
}

output "detailed_status" {
  value = data.openai_run.details.status
}

output "steps_information" {
  value = data.openai_run.details.steps
}
```

## Accessing Run Steps

You can access specific steps and their details:

```hcl
data "openai_run" "with_steps" {
  thread_id = "thread_abc123"
  run_id    = "run_xyz789"
}

output "step_details" {
  value = [
    for step in data.openai_run.with_steps.steps : {
      id     = step.id
      type   = step.type
      status = step.status
      details = step.details
    }
  ]
}
```

## Argument Reference

* `thread_id` - (Required) The ID of the thread the run belongs to.
* `run_id` - (Required) The ID of the run to retrieve.

## Attribute Reference

* `id` - The identifier of the run.
* `thread_id` - The ID of the thread the run belongs to.
* `assistant_id` - The ID of the assistant used for the run.
* `model` - The model used for the run.
* `instructions` - The instructions used for the run, which may be different from the assistant's default instructions if they were overridden.
* `tools` - The tools available to the assistant during the run.
* `file_ids` - The file IDs available to the assistant during the run.
* `metadata` - Any metadata attached to the run.
* `status` - The status of the run, one of: queued, in_progress, requires_action, cancelling, cancelled, failed, completed, expired.
* `created_at` - The Unix timestamp (in seconds) of when the run was created.
* `started_at` - The Unix timestamp (in seconds) of when the run was started, if it has started.
* `completed_at` - The Unix timestamp (in seconds) of when the run was completed, if it has completed.
* `object` - The object type, which is always 'thread.run'.
* `usage` - Token usage statistics for the run:
  * `prompt_tokens` - Number of tokens in the prompt.
  * `completion_tokens` - Number of tokens in the completion.
  * `total_tokens` - Total tokens used.
* `steps` - A list of steps in the run, each with these attributes:
  * `id` - The ID of the step.
  * `object` - The object type, which is always 'thread.run.step'.
  * `created_at` - The Unix timestamp (in seconds) of when the step was created.
  * `type` - The type of the step.
  * `status` - The status of the step.
  * `details` - Details specific to the step type, encoded as a JSON string.
* `last_error` - Information about the last error that occurred during the run, if any:
  * `code` - The error code.
  * `message` - The error message.
* `required_action` - If the run requires action, this contains information about what action is needed.

## Usage Notes

- This data source is particularly useful for checking the status of long-running operations or retrieving the final results of a run.
- Since runs progress through different states (queued, in_progress, etc.), you may need to query the data source multiple times to check for completion.
- Use the `depends_on` attribute when retrieving information about a run that was just created in the same configuration to ensure proper dependency ordering. 