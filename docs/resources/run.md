---
page_title: "OpenAI: openai_run Resource"
subcategory: ""
description: |-
  Executes an OpenAI assistant on a thread to generate responses.
---

# OpenAI Run

This resource allows you to create a run on an existing thread in the OpenAI Assistants API. A run represents the execution of an assistant on a thread, where the assistant processes messages, generates responses using tools, and adds messages back to the thread.

## Example Usage

```hcl
resource "openai_run" "example" {
  thread_id    = openai_thread.example.id
  assistant_id = openai_assistant.example.id
  
  model        = "gpt-4o"            # Optional, overrides the assistant's model
  instructions = "You are a helpful terraform expert focusing on AWS resources" # Optional, overrides assistant instructions
  
  tools {
    type = "code_interpreter"
  }
  
  tools {
    type = "retrieval"
  }
  
  tools {
    type = "function"
    function {
      name        = "get_resource_documentation"
      description = "Get documentation for an AWS resource"
      parameters  = jsonencode({
        type = "object",
        properties = {
          resource = {
            type = "string",
            description = "The AWS resource type"
          }
        },
        required = ["resource"]
      })
    }
  }
  
  metadata = {
    purpose = "infrastructure-planning"
    priority = "high"
  }
  
  temperature = 0.7
  max_tokens  = 2048
  top_p       = 0.95
}
```

## Run with Completion Window

This example creates a run and waits for it to complete within a specified time window:

```hcl
resource "openai_run" "with_completion" {
  thread_id    = openai_thread.conversation.id
  assistant_id = openai_assistant.expert.id
  
  # Wait up to 60 seconds for the run to complete
  completion_window = 60
}
```

## Argument Reference

* `thread_id` - (Required) The ID of the thread to run the assistant on.
* `assistant_id` - (Required) The ID of the assistant to use for the run.
* `model` - (Optional) The model to use for the run. If not provided, the assistant's model will be used.
* `instructions` - (Optional) Override the default instructions of the assistant for this specific run.
* `tools` - (Optional) Override the tools the assistant can use for this run.
* `metadata` - (Optional) Set of key-value pairs that can be attached to the run.
* `temperature` - (Optional) What sampling temperature to use, between 0 and 2. Higher values make output more random, lower values more deterministic.
* `max_tokens` - (Optional) The maximum number of tokens to generate in the run.
* `top_p` - (Optional) An alternative to sampling with temperature, where the model considers the results of the tokens with top_p probability mass.
* `stream_for_tool` - (Optional) Whether to stream tool outputs. Only available in the Chat Completions API.
* `completion_window` - (Optional) The maximum amount of time to wait for the run to complete, in seconds. If not provided, the run will be created but not waited for.

### Nested `tools` block

* `type` - (Required) The type of tool: code_interpreter, retrieval, or function.
* `function` - (Optional) Defines a function that can be called by the assistant. Required when type is function.

### Nested `tools.function` block

* `name` - (Required) The name of the function.
* `description` - (Optional) A description of what the function does.
* `parameters` - (Optional) The parameters the function accepts, described as a JSON Schema object.

## Attribute Reference

* `id` - The identifier of the run, which can be referenced in API endpoints.
* `object` - The object type, which is always 'thread.run'.
* `created_at` - The Unix timestamp (in seconds) of when the run was created.
* `started_at` - The Unix timestamp (in seconds) of when the run was started.
* `completed_at` - The Unix timestamp (in seconds) of when the run was completed.
* `status` - The status of the run (queued, in_progress, completed, failed, etc.).
* `file_ids` - The IDs of the files used in the run.
* `usage` - Usage statistics for the run (a map with prompt_tokens, completion_tokens, and total_tokens).
* `steps` - The steps of the run, each with its own ID, type, status, and details.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 10 minutes) Used when creating the run.
* `read` - (Defaults to 5 minutes) Used when retrieving the run.
* `delete` - (Defaults to 5 minutes) Used when deleting the run.

## Import

Runs can be imported using the format `thread_id:run_id`, e.g.,

```bash
terraform import openai_run.example thread_abc123:run_xyz789
``` 