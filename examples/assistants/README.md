# OpenAI Assistants Example

This example demonstrates how to use the OpenAI Terraform Provider to create and manage OpenAI Assistants.

## What are OpenAI Assistants?

Assistants are AI systems that can be configured to perform specific tasks or respond in particular ways. The OpenAI Assistants API allows you to create assistants with various configurations, including:

- Different base models (like GPT-4o)
- Custom instructions
- Tools for extending capabilities (code interpreter, file search, function calling)
- File attachments for reference during conversations

## Example Components

This example demonstrates:

1. Creating a Math Tutor assistant with the code interpreter tool
2. Creating a Writing Assistant for content creation
3. Using the Assistants data source to list all assistants
4. Using a data source to retrieve a single assistant by ID
5. Importing an existing assistant from OpenAI into Terraform
6. Outputting assistant properties and attributes

## Usage

To run this example:

1. Set your OpenAI API key:
```bash
export OPENAI_API_KEY="your-api-key"
```

2. Initialize Terraform:
```bash
terraform init
```

3. Apply the configuration:
```bash
terraform apply
```

## Terraform Resources Used

| Resource/Data Source | Description |
|----------------------|-------------|
| `openai_assistant` | Create and manage OpenAI assistants |
| `openai_assistants` | List and retrieve information about all assistants |
| `openai_assistant` (data source) | Retrieve a single assistant by ID |

## Assistant Properties

The `openai_assistant` resource supports the following properties:

- `name` - The name of the assistant
- `model` - The model to use (e.g., gpt-4o)
- `instructions` - System instructions for the assistant
- `description` - A description of the assistant
- `tools` - List of tools the assistant can use
- `metadata` - Key-value pairs for organizing assistants
- `file_ids` - IDs of files the assistant can access

## Advanced Usage

For more advanced usage, including:
- Creating assistants with function calling
- Attaching files to assistants
- Setting up assistants with the file search tool
- Importing existing assistants

See the [OpenAI Assistants API documentation](https://platform.openai.com/docs/api-reference/assistants) and the provider documentation.

### Importing an Existing Assistant

If you have assistants that were created outside of Terraform (e.g., through the OpenAI API or dashboard), you can import them into Terraform management:

1. Create a placeholder resource in your Terraform configuration:
```hcl
resource "openai_assistant" "imported_assistant" {
  name  = "Placeholder Name"  # Will be overwritten on import
  model = "gpt-4o"            # Will be overwritten on import
}
```

2. Run the import command:
```bash
terraform import openai_assistant.imported_assistant asst_abc123
```

3. After importing, you can view the imported state and modify properties as needed:
```bash
terraform state show openai_assistant.imported_assistant
```

## Output Example

After applying this configuration, you'll get outputs similar to:

```
math_tutor_id = "asst_abc123..."
math_tutor_created_at = 1699009709
writing_assistant_id = "asst_xyz456..."
all_assistants_count = 2
all_assistants_ids = [
  "asst_abc123...",
  "asst_xyz456..."
]
data_math_tutor_name = "Math Tutor"
data_math_tutor_model = "gpt-4o"
data_math_tutor_instructions = "You are a personal math tutor..."
``` 