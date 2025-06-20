# OpenAI Thread Run Module

This module simplifies the process of creating a thread and starting a run in the OpenAI Assistants API. It handles the configuration of both thread creation and run execution in a single module.

## Usage

```hcl
module "openai_thread_run" {
  source = "./modules/run"

  assistant_id = openai_assistant.example.id
  
  # Thread configuration
  create_new_thread = true
  messages = [
    {
      role    = "user"
      content = "What are the benefits of using Terraform modules?"
    },
    {
      role    = "user"
      content = "Can you provide examples of modular Terraform architecture?"
    }
  ]
  thread_metadata = {
    project = "terraform-documentation"
  }
  
  # Run configuration
  model        = "gpt-4o"
  instructions = "Provide detailed technical examples in your responses."
  tools = [
    {
      type = "code_interpreter"
    },
    {
      type     = "function"
      function = {
        name        = "fetch_examples"
        description = "Fetch examples of Terraform modules"
        parameters  = jsonencode({
          type       = "object"
          properties = {
            category = {
              type        = "string"
              description = "The category of modules to fetch"
            }
          }
          required = ["category"]
        })
      }
    }
  ]
  
  temperature = 0.7
  max_completion_tokens = 2048
}
```

## Using an Existing Thread

```hcl
module "openai_thread_run" {
  source = "./modules/run"

  assistant_id     = openai_assistant.example.id
  create_new_thread = false
  existing_thread_id = openai_thread.existing.id
  
  instructions = "Please summarize the conversation so far."
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| assistant_id | The ID of the assistant to use for the run | `string` | n/a | yes |
| create_new_thread | Whether to create a new thread | `bool` | `true` | no |
| existing_thread_id | ID of an existing thread to use (required if create_new_thread is false) | `string` | `null` | no |
| messages | Messages to add to the thread | `list(object)` | `[]` | no |
| thread_metadata | Metadata to attach to the thread | `map(string)` | `{}` | no |
| model | Model to use for the run (overrides assistant's model) | `string` | `null` | no |
| instructions | Instructions for this specific run | `string` | `null` | no |
| tools | Tools the assistant can use for this run | `list(object)` | `[]` | no |
| metadata | Metadata to attach to the run | `map(string)` | `{}` | no |
| temperature | Sampling temperature (0.0 to 2.0) | `number` | `null` | no |
| top_p | Nucleus sampling parameter (0.0 to 1.0) | `number` | `null` | no |
| max_completion_tokens | Maximum number of tokens for completion | `number` | `null` | no |
| stream | Whether to stream the response (not supported in Terraform) | `bool` | `false` | no |
| response_format | Format for the assistant's response | `object` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| id | The ID of the created run |
| thread_id | The ID of the thread used for the run |
| status | The status of the run |
| created_at | When the run was created |
| started_at | When the run was started |
| completed_at | When the run was completed |
| usage | Usage statistics for the run |

## Implementation Notes

This module wraps the `openai_thread_run` resource to provide a simplified interface for creating threads and starting runs. It handles the conditional creation of a new thread or the use of an existing thread based on the `create_new_thread` parameter.

For complex configuration needs, consider using the underlying resource directly. 