# OpenAI Runs Data Source Example
# This example demonstrates various ways to use the openai_runs data source
# to retrieve and analyze information about runs in the OpenAI Assistants API.

# First, create some resources that we'll query with the data source
resource "openai_assistant" "example" {
  name         = "Run Testing Assistant"
  description  = "Assistant used for testing the runs data source"
  model        = "gpt-4o"
  instructions = "You are an assistant that helps with testing. Provide concise answers."

  tools {
    type = "code_interpreter"
  }
}

resource "openai_thread" "example" {
  messages {
    role    = "user"
    content = "This is a test message for the runs data source example."
  }
}

# Create multiple runs on the same thread
resource "openai_run" "first" {
  thread_id    = openai_thread.example.id
  assistant_id = openai_assistant.example.id

  metadata = {
    purpose = "test-run-1"
    type    = "example"
  }
}

resource "openai_run" "second" {
  thread_id    = openai_thread.example.id
  assistant_id = openai_assistant.example.id

  # This depends on the first run to ensure sequential creation
  depends_on = [openai_run.first]

  metadata = {
    purpose = "test-run-2"
    type    = "example"
  }
}

# Example 1: Basic retrieval of runs for a thread
data "openai_runs" "basic_example" {
  thread_id = openai_thread.example.id
  limit     = 10

  # Only query after both runs are created
  depends_on = [openai_run.first, openai_run.second]
}

# Example 2: This would filter runs by status if the feature was supported
# For now, we'll just retrieve all runs and filter in the output
data "openai_runs" "all_runs" {
  thread_id = openai_thread.example.id
  limit     = 5

  depends_on = [openai_run.first, openai_run.second]
}

# Example 3: Descending order (newest first)
data "openai_runs" "newest_first" {
  thread_id = openai_thread.example.id
  limit     = 5
  order     = "desc"

  depends_on = [openai_run.first, openai_run.second]
}

# Example 4: Ascending order (oldest first)
data "openai_runs" "oldest_first" {
  thread_id = openai_thread.example.id
  limit     = 5
  order     = "asc"

  depends_on = [openai_run.first, openai_run.second]
}

# Example 5: Pagination demonstration
data "openai_runs" "page_one" {
  thread_id = openai_thread.example.id
  limit     = 1
  order     = "desc"

  depends_on = [openai_run.first, openai_run.second]
}

data "openai_runs" "page_two" {
  thread_id = openai_thread.example.id
  limit     = 1
  order     = "desc"
  after     = data.openai_runs.page_one.last_id

  depends_on = [data.openai_runs.page_one]
}

# Outputs to demonstrate the data source results

output "basic_example_count" {
  value       = length(data.openai_runs.basic_example.runs)
  description = "Count of runs retrieved in the basic example"
}

output "basic_example_runs" {
  value = [
    for run in data.openai_runs.basic_example.runs : {
      id           = run.id
      status       = run.status
      created_at   = run.created_at
      assistant_id = run.assistant_id
    }
  ]
  description = "Basic information about each run"
}

# Filter completed runs in the output instead
output "completed_runs" {
  value = [
    for run in data.openai_runs.all_runs.runs : run
    if run.status == "completed"
  ]
  description = "Only the completed runs"
}

output "completed_runs_count" {
  value = length([
    for run in data.openai_runs.all_runs.runs : run
    if run.status == "completed"
  ])
  description = "Count of completed runs"
}

output "newest_run" {
  value = length(data.openai_runs.newest_first.runs) > 0 ? {
    id         = data.openai_runs.newest_first.runs[0].id
    created_at = data.openai_runs.newest_first.runs[0].created_at
    status     = data.openai_runs.newest_first.runs[0].status
  } : null
  description = "Information about the newest run"
}

output "oldest_run" {
  value = length(data.openai_runs.oldest_first.runs) > 0 ? {
    id         = data.openai_runs.oldest_first.runs[0].id
    created_at = data.openai_runs.oldest_first.runs[0].created_at
    status     = data.openai_runs.oldest_first.runs[0].status
  } : null
  description = "Information about the oldest run"
}

output "paginated_runs" {
  value = concat(
    [for run in data.openai_runs.page_one.runs : run.id],
    [for run in data.openai_runs.page_two.runs : run.id]
  )
  description = "IDs of runs retrieved through pagination"
}

output "has_more_runs" {
  value       = data.openai_runs.page_two.has_more
  description = "Whether there are more runs available beyond the pages we retrieved"
} 