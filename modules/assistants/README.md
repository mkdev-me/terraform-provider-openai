# OpenAI Assistants Terraform Module

This module provides a simplified interface for working with OpenAI Assistants using the OpenAI Terraform Provider.

## Features

- Create and manage OpenAI Assistants
- Configure assistant properties including model, instructions, and tools
- Retrieve lists of assistants using data sources
- Fetch a specific assistant by ID
- Convenient outputs for all assistant properties

## Usage

```hcl
module "math_assistant" {
  source = "../../modules/assistants"

  # Assistant configuration
  assistant_name         = "Math Tutor"
  assistant_model        = "gpt-4o"
  assistant_instructions = "You are a personal math tutor. When asked a question, write and run Python code to answer the question."
  assistant_description  = "A math tutor that uses code interpreter to solve problems"
  
  # Enable code interpreter tool
  assistant_tools = [
    {
      type = "code_interpreter"
    }
  ]
  
  # Add metadata
  assistant_metadata = {
    "created_by" = "terraform",
    "version"    = "1.0"
  }
}

# Output the assistant ID
output "assistant_id" {
  value = module.math_assistant.assistant_id
}
```

## Advanced Usage: Function Calling

```hcl
module "function_assistant" {
  source = "../../modules/assistants"

  assistant_name         = "Weather Assistant"
  assistant_model        = "gpt-4o"
  assistant_instructions = "You are a weather assistant that can help users check the weather."
  
  # Configure a function tool
  assistant_tools = [
    {
      type = "function"
      function = {
        name        = "get_weather"
        description = "Get the current weather in a given location"
        parameters  = jsonencode({
          type = "object",
          properties = {
            location = {
              type = "string",
              description = "The city and state, e.g., San Francisco, CA"
            },
            unit = {
              type = "string",
              enum = ["celsius", "fahrenheit"],
              description = "The unit of temperature to use"
            }
          },
          required = ["location"]
        })
      }
    }
  ]
}
```

## Using the Data Source

```hcl
module "assistants_list" {
  source = "../../modules/assistants"
  
  # Disable creating an assistant
  enable_assistant = false
  
  # Enable the data source
  enable_assistants_data_source = true
  assistants_limit = 10
  assistants_order = "desc"
}

# Output all assistants
output "all_assistants" {
  value = module.assistants_list.all_assistants
}
```

## Fetching a Single Assistant

```hcl
module "fetch_assistant" {
  source = "../../modules/assistants"
  
  # Disable creating an assistant
  enable_assistant = false
  
  # Enable fetching a single assistant
  enable_single_assistant_data_source = true
  single_assistant_id = "asst_abc123"  # Replace with your assistant ID
}

# Output details of the specific assistant
output "assistant_name" {
  value = module.fetch_assistant.single_assistant_name
}

output "assistant_model" {
  value = module.fetch_assistant.single_assistant_model
}
```

## Importing an Existing Assistant

You can also use this module with assistants imported through Terraform:

1. First, create a placeholder resource in your main configuration:
```hcl
resource "openai_assistant" "imported" {
  name  = "Placeholder"  # Will be overwritten on import
  model = "gpt-4o"       # Will be overwritten on import
}
```

2. Import the assistant:
```bash
terraform import openai_assistant.imported asst_abc123
```

3. Then use the module with the imported resource:
```hcl
module "use_imported" {
  source = "../../modules/assistants"
  
  # Don't create a new assistant, we'll reference the imported one
  enable_assistant = false
  
  # Get details about the imported assistant
  enable_single_assistant_data_source = true
  single_assistant_id = openai_assistant.imported.id
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 0.13.0 |
| openai | >= 1.0.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| enable_assistant | Whether to create the assistant | `bool` | `true` | no |
| assistant_name | The name of the assistant | `string` | `null` | no |
| assistant_model | ID of the model to use for the assistant | `string` | n/a | yes |
| assistant_instructions | The system instructions that the assistant uses | `string` | `null` | no |
| assistant_description | The description of the assistant | `string` | `null` | no |
| assistant_tools | List of tools enabled on the assistant | `list(object)` | `[]` | no |
| assistant_file_ids | List of file IDs attached to the assistant | `list(string)` | `[]` | no |
| assistant_metadata | Metadata for the assistant | `map(string)` | `{}` | no |
| enable_assistants_data_source | Whether to fetch the assistants list | `bool` | `false` | no |
| assistants_limit | Limit on the number of assistants to fetch (1-100) | `number` | `20` | no |
| assistants_order | Sort order by created_at timestamp (asc or desc) | `string` | `"desc"` | no |
| assistants_after | Cursor for pagination (fetch after this assistant ID) | `string` | `null` | no |
| assistants_before | Cursor for pagination (fetch before this assistant ID) | `string` | `null` | no |
| enable_single_assistant_data_source | Whether to fetch a single assistant by ID | `bool` | `false` | no |
| single_assistant_id | ID of a specific assistant to fetch | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| assistant_id | The ID of the created assistant |
| assistant_name | The name of the created assistant |
| assistant_created_at | The timestamp when the assistant was created |
| assistant_model | The model used by the assistant |
| assistant_description | The description of the assistant |
| assistant_instructions | The system instructions of the assistant |
| assistant_tools | The tools enabled on the assistant |
| assistant_file_ids | The file IDs attached to the assistant |
| assistant_metadata | The metadata attached to the assistant |
| all_assistants | List of all assistants retrieved by the data source |
| all_assistants_count | Count of all assistants retrieved by the data source |
| first_assistant_id | The ID of the first assistant in the list |
| last_assistant_id | The ID of the last assistant in the list |
| has_more_assistants | Whether there are more assistants available beyond the current list |
| single_assistant | Details of a specific assistant retrieved by ID |
| single_assistant_id | ID of the specific assistant retrieved |
| single_assistant_name | Name of the specific assistant retrieved |
| single_assistant_model | Model of the specific assistant retrieved |
| single_assistant_instructions | Instructions of the specific assistant retrieved | 