# OpenAI Runs Example

This example demonstrates how to use the OpenAI Terraform provider to create and manage Runs in the OpenAI Assistants API.

## What are Runs?

Runs represent the execution of an Assistant on a Thread. When you create a Run, the Assistant processes all the Messages in the Thread, uses its configuration and tools to generate a response, and then adds the response as a new Message to the Thread.

## Example Use Cases

This example covers:

1. Creating a basic run with default settings
2. Creating a run with custom parameters (model override, instructions, temperature)
3. Using the run module for simplified run creation and management
4. Creating a thread and run in a single operation
5. Using data sources to retrieve information about existing runs

## Prerequisites

- An OpenAI account with access to the Assistants API
- An API key with appropriate permissions

## Usage

1. Set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key"
```

2. Initialize Terraform:

```bash
terraform init
```

3. Apply the Terraform configuration:

```bash
terraform apply
```

## Understanding the Example

This example demonstrates:

- `openai_run`: The resource for creating a run on an existing thread
- `openai_thread_run`: The resource for creating a thread and starting a run in one operation
- `data.openai_run`: Data source for retrieving information about an existing run
- `data.openai_thread_run`: Data source for retrieving information about an existing thread run
- The OpenAI run module, which simplifies run creation

## Resource Types

### openai_run

Creates a run on an existing thread.

```hcl
resource "openai_run" "example" {
  thread_id    = "thread_abc123"
  assistant_id = "asst_abc123"
  model        = "gpt-4o"  # Optional, overrides the assistant's model
  instructions = "Additional instructions for this run only"
}
```

### openai_thread_run

Creates a thread and a run in a single operation.

```hcl
resource "openai_thread_run" "example" {
  assistant_id = "asst_abc123"
  
  thread {
    messages {
      role    = "user"
      content = "Hello, can you help me with a question?"
    }
  }
}
```

## Outputs

This example provides the following outputs:

- `assistant_id`: The ID of the created assistant
- `thread_id`: The ID of the thread used for runs
- `basic_run`: Details of the basic run
- `custom_run`: Details of the run with custom parameters
- `module_run`: Details of the run created using the module
- `combined_thread_run`: Details of the combined thread and run operation
- `run_info_from_data_source`: Run information retrieved via data source
- `thread_run_info_from_data_source`: Thread run information retrieved via data source 