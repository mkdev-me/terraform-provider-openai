# OpenAI Thread Run

This resource allows you to create a thread and start a run in a single operation. This simplifies the process of using the OpenAI Assistants API by combining two steps into one.

## Example Usage

```hcl
resource "openai_thread_run" "example" {
  assistant_id = openai_assistant.example.id
  
  thread {
    messages {
      role    = "user"
      content = "Hello, can you help me with a question about Terraform?"
    }
    
    messages {
      role    = "user"
      content = "What are the best practices for organizing Terraform modules?"
      attachments {
        file_id = openai_file.documentation.id
      }
    }
  }
  
  instructions = "You are a helpful Terraform expert assistant."
  model        = "gpt-4o"
  
  tools {
    type = "retrieval"
  }
  
  tools {
    type = "function"
    function {
      name        = "get_weather"
      description = "Get the current weather in a given location"
      parameters  = jsonencode({
        type = "object",
        properties = {
          location = {
            type = "string",
            description = "The city and state, e.g., San Francisco, CA"
          }
        },
        required = ["location"]
      })
    }
  }
  
  temperature = 0.7
  max_completion_tokens = 1024
}
```

## Creating a Run on an Existing Thread

```hcl
resource "openai_thread_run" "on_existing" {
  assistant_id = openai_assistant.expert.id
  existing_thread_id = openai_thread.existing.id
  
  instructions = "Please analyze the data in this thread."
  model = "gpt-4o-mini"
}
```

## Argument Reference

* `assistant_id` - (Required) The ID of the assistant to use for the run.
* `thread` - (Optional) Configuration block for creating a new thread. Conflicts with `existing_thread_id`.
* `existing_thread_id` - (Optional) The ID of an existing thread to use for this run. Conflicts with `thread`.
* `model` - (Optional) The ID of the model to use for the run. If not provided, the assistant's default model will be used.
* `instructions` - (Optional) Instructions that override the assistant's instructions for this run only.
* `tools` - (Optional) Override the tools the assistant can use for this run.
* `metadata` - (Optional) Set of key-value pairs that can be attached to the run.
* `stream` - (Optional) Whether to stream the run results. Not currently supported through the Terraform provider. Default is `false`.
* `max_completion_tokens` - (Optional) The maximum number of tokens that can be generated in the run completion.
* `temperature` - (Optional) What sampling temperature to use, between 0 and 2. Higher values make output more random, lower values more deterministic.
* `top_p` - (Optional) An alternative to sampling with temperature, where the model considers the results of the tokens with top_p probability mass.
* `response_format` - (Optional) Specifies the format of the response.

### Nested `thread` block

* `messages` - (Optional) Messages to create on the new thread.
* `metadata` - (Optional) Set of key-value pairs that can be attached to the thread.

### Nested `thread.messages` block

* `role` - (Required) The role of the entity that is creating the message. Currently only 'user' is supported.
* `content` - (Required) The content of the message.
* `attachments` - (Optional) A list of attachments to include in the message.
* `metadata` - (Optional) Set of key-value pairs that can be attached to the message.

### Nested `thread.messages.attachments` block

* `file_id` - (Required) The ID of the file to attach to the message.

### Nested `tools` block

* `type` - (Required) The type of tool: code_interpreter, retrieval, or function.
* `function` - (Optional) Defines a function that can be called by the assistant. Required when type is function.

### Nested `tools.function` block

* `name` - (Required) The name of the function.
* `description` - (Optional) A description of what the function does.
* `parameters` - (Optional) The parameters the function accepts, described as a JSON Schema object.

### Nested `response_format` block

* `type` - (Required) Must be one of 'text' or 'json_object'.

## Attribute Reference

* `id` - The identifier of the run.
* `thread_id` - The ID of the thread that was created and associated with this run.
* `created_at` - The Unix timestamp (in seconds) of when the run was created.
* `completed_at` - The Unix timestamp (in seconds) of when the run was completed.
* `started_at` - The Unix timestamp (in seconds) of when the run was started.
* `object` - The object type, which is always 'thread.run'.
* `status` - The status of the run.
* `file_ids` - A list of file IDs that the run has access to.
* `usage` - Usage statistics for the run, including prompt_tokens, completion_tokens, and total_tokens.

## Import

Thread runs can be imported using the format `thread_id:run_id`, e.g.,

```bash
terraform import openai_thread_run.example thread_abc123:run_xyz789
``` 