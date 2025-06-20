# OpenAI Model Response Module

This Terraform module provides a simplified way to generate text responses from OpenAI's language models using the `openai_model_response` resource. It handles the configuration of model parameters and provides convenient outputs for accessing the generated responses.

## Usage

```hcl
module "bedtime_story" {
  source = "../../modules/model_response"

  input = "Generate a three-sentence bedtime story about a unicorn."
  model = "gpt-4o"
  
  temperature      = 0.7
  max_output_tokens = 100
  instructions     = "The story should be suitable for young children and have a positive message."
}

output "story" {
  value = module.bedtime_story.output_text
}

output "token_usage" {
  value = module.bedtime_story.usage
}
```

### Advanced Usage

```hcl
module "technical_explanation" {
  source = "../../modules/model_response"

  input        = "Explain how quantum computing works."
  model        = "gpt-4o"
  temperature  = 0.5
  top_p        = 0.9
  
  # Limit generation length
  max_output_tokens = 300
  
  # Add custom instructions
  instructions = "Explain at a high school level, avoiding overly technical jargon."
  
  # Control repetition and diversity
  frequency_penalty = 0.8
  presence_penalty  = 0.4
  
  # Stop sequences to terminate generation
  stop_sequences = [
    "In conclusion",
    "To summarize"
  ]
  
  # Prevent recreation when configuration changes
  preserve_on_change = true
}
```

## Importing Existing Responses

You can import existing OpenAI model responses to be managed by this module. This is particularly useful when you want to:

1. Bring existing responses under Terraform management
2. Preserve outputs you've already generated and like
3. Manage responses without having to remember the exact inputs that created them

### Import Process

1. Define a minimal module instance in your Terraform configuration:

```hcl
module "imported_response" {
  source = "../../modules/model_response"
  
  # You can leave input and model empty - they'll be populated from the API
  # The provider will automatically retrieve all details from the OpenAI API
  
  # Recommended to keep this set to true for imported resources
  preserve_on_change = true
}
```

2. Run the import command, replacing the placeholder ID with your actual response ID:

```bash
terraform import module.imported_response.openai_model_response.response resp_67edb65336b881919b5703f0a37064fa07d30f4d397146c6
```

3. Verify the import:

```bash
terraform state show module.imported_response.openai_model_response.response
```

The provider automatically:
- Retrieves all model parameters from the API, including the original input prompt
- Sets imported and preserve_on_change flags to prevent accidental recreation
- Preserves the existing output without making additional API calls

You can then use the module outputs normally:

```hcl
output "imported_text" {
  value = module.imported_response.output_text
}
```

## Immutability and Preservation

By default, model responses are immutable - changes to input parameters will cause Terraform to destroy the existing resource and create a new one, generating a fresh response from the OpenAI API.

Setting `preserve_on_change = true` will:
1. Prevent recreation of the resource when parameters change
2. Maintain the same response text across configuration changes
3. Show drift detection in Terraform between configuration and the actual resource

This is useful when:
- You want consistent responses across infrastructure updates
- You're experimenting with different parameters but don't want to generate new responses
- You want to avoid unnecessary API calls and costs

## Related Data Sources

The provider includes data sources to retrieve information about existing model responses:

1. **openai_model_response** - Retrieves a specific model response by its ID:
```hcl
data "openai_model_response" "existing_response" {
  response_id = "resp_abc123def456" # Replace with an actual response ID
}

output "response_text" {
  value = data.openai_model_response.existing_response.output.text
}
```

You can also use the module output directly as input to a data source:
```hcl
module "bedtime_story" {
  source = "../../modules/model_response"
  input  = "Tell me a bedtime story."
  model  = "gpt-4o"
}

data "openai_model_response" "bedtime_story_data" {
  response_id = module.bedtime_story.id
}

# Both provide the same information
output "from_module" {
  value = module.bedtime_story.output_text
}

output "from_data_source" {
  value = data.openai_model_response.bedtime_story_data.output.text
}
```

2. **openai_model_response_input_items** - Retrieves the input items (prompts) for a model response:
```hcl
data "openai_model_response_input_items" "response_inputs" {
  response_id = "resp_abc123def456" # Replace with an actual response ID
}

output "first_input" {
  value = length(data.openai_model_response_input_items.response_inputs.input_items) > 0 ? 
    data.openai_model_response_input_items.response_inputs.input_items[0].content : 
    "No input items found"
}
```

3. **openai_model_responses** - Lists multiple model responses (requires browser session authentication):
```hcl
data "openai_model_responses" "recent_responses" {
  limit = 5
  order = "desc" # Most recent first
}

output "recent_response_ids" {
  value = [for resp in data.openai_model_responses.recent_responses.responses : resp.id]
}
```

See the [provider documentation](../../docs/data-sources) for more details on these data sources.

## Required Inputs

| Name | Description | Type | Default |
|------|-------------|------|---------|
| `input` | The input text to generate a response for | `string` | - |
| `model` | ID of the model to use (e.g., "gpt-4o", "gpt-4-turbo") | `string` | - |

## Optional Inputs

| Name | Description | Type | Default |
|------|-------------|------|---------|
| `temperature` | Sampling temperature (0-2). Higher values mean more randomness | `number` | `0.7` |
| `max_output_tokens` | Maximum number of tokens to generate | `number` | `null` |
| `top_p` | Nucleus sampling parameter (0-1) | `number` | `null` |
| `top_k` | Top-k sampling parameter | `number` | `null` |
| `include` | Optional fields to include in the response | `list(string)` | `null` |
| `instructions` | Optional instructions to guide the model | `string` | `null` |
| `stop_sequences` | Sequences where generation stops | `list(string)` | `null` |
| `frequency_penalty` | Penalty for repeated tokens (-2 to 2) | `number` | `0` |
| `presence_penalty` | Penalty for new tokens (-2 to 2) | `number` | `0` |
| `user` | User identifier for tracking | `string` | `null` |
| `preserve_on_change` | If true, prevents recreation when parameters change | `bool` | `false` |

## Outputs

| Name | Description |
|------|-------------|
| `id` | Unique identifier for the response |
| `created` | Unix timestamp when the response was created |
| `object` | Object type (usually "model_response") |
| `output` | The generated output (map containing text and token_count) |
| `output_text` | The generated response text |
| `token_count` | Number of tokens in the output |
| `usage` | Token usage statistics (input_tokens, output_tokens, total_tokens) |
| `finish_reason` | Reason why the generation stopped |

## Notes

- The model response is generated when Terraform applies the configuration and doesn't change unless the input parameters change and `preserve_on_change` is false
- This is a stateless resource - deleting it only removes it from Terraform state and doesn't affect anything in OpenAI
- The OpenAI API key is managed by the provider configuration, not this module
- Models like GPT-4o are optimized for instruction-following, so consider using the `instructions` parameter for more control 