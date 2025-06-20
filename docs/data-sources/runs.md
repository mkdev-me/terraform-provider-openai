# openai_runs Data Source

Retrieves a list of runs for a given thread in the OpenAI Assistants API. This data source allows you to query multiple runs at once to analyze patterns, monitor status, or gather historical data about assistant interactions.

## Example Usage

Retrieve all runs for a specific thread:

```hcl
data "openai_runs" "thread_history" {
  thread_id = "thread_abc123"
  limit     = 10
}

output "recent_runs" {
  value = data.openai_runs.thread_history.runs
}

output "total_runs_count" {
  value = length(data.openai_runs.thread_history.runs)
}
```

Filter runs by run status:

```hcl
data "openai_runs" "completed_runs" {
  thread_id = openai_thread.example.id
  limit     = 25
  status    = "completed"
  
  depends_on = [openai_run.example]
}

output "completed_runs_count" {
  value = length(data.openai_runs.completed_runs.runs)
}
```

Use pagination to retrieve runs in batches:

```hcl
data "openai_runs" "first_page" {
  thread_id = openai_thread.conversation.id
  limit     = 5
}

data "openai_runs" "second_page" {
  thread_id = openai_thread.conversation.id
  limit     = 5
  after     = data.openai_runs.first_page.last_id
}

output "all_runs" {
  value = concat(
    data.openai_runs.first_page.runs,
    data.openai_runs.second_page.runs
  )
}
```

## Argument Reference

* `thread_id` - (Required) The ID of the thread to list runs for.
* `limit` - (Optional) A limit on the number of runs to retrieve. Defaults to 20.
* `order` - (Optional) Sort order by the created_at timestamp. One of "asc" or "desc". Defaults to "desc".
* `after` - (Optional) A cursor for pagination. This should be the ID of the last run from a previous request.
* `before` - (Optional) A cursor for pagination. This should be the ID of the first run from a previous request.
* `status` - (Optional) Filter runs by status. One of: queued, in_progress, requires_action, cancelling, cancelled, failed, completed, expired.

## Attribute Reference

* `id` - An identifier for this data source.
* `thread_id` - The ID of the thread the runs belong to.
* `runs` - A list of run objects, each containing:
  * `id` - The identifier of the run.
  * `object` - The object type, which is always 'thread.run'.
  * `created_at` - The Unix timestamp (in seconds) of when the run was created.
  * `thread_id` - The ID of the thread the run belongs to.
  * `assistant_id` - The ID of the assistant used for the run.
  * `status` - The status of the run.
  * `started_at` - The Unix timestamp (in seconds) of when the run was started, if it has started.
  * `completed_at` - The Unix timestamp (in seconds) of when the run was completed, if it has completed.
  * `model` - The model used for the run.
  * `instructions` - The instructions used for the run.
  * `tools` - The tools available to the assistant during the run.
  * `file_ids` - The file IDs available to the assistant during the run.
  * `metadata` - Any metadata attached to the run.
* `first_id` - The ID of the first run in the response.
* `last_id` - The ID of the last run in the response.
* `has_more` - Whether there are more runs available beyond the current page.

## Usage Notes

- Use this data source when you need to analyze or report on multiple runs for a thread.
- The `limit` parameter can be adjusted based on your needs, but be aware that retrieving a large number of runs may impact performance.
- For efficient pagination, use the `after` parameter with the `last_id` from previous queries.
- For complete details about a specific run, use the `openai_run` data source with the run ID obtained from this list.
- When filtering by status, be aware that run statuses change over time, so results will vary based on when the data source is queried. 