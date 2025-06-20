---
page_title: "OpenAI: openai_model_response Resource"
subcategory: "Model Responses"
description: |-
  Generates text responses from OpenAI language models using the new responses API.
---

# openai_model_response Resource

This resource allows you to generate text responses from OpenAI's language models (like GPT-4o, GPT-4, etc.) using the modern `/v1/responses` API. You provide an input prompt and the desired model, and OpenAI generates a text response based on your parameters.

## Example Usage

```hcl
# Basic usage
resource "openai_model_response" "simple" {
  input = "Tell me a three-sentence story about a robot."
  model = "gpt-4o"
}

# Advanced usage with parameters
resource "openai_model_response" "advanced" {
  input = "Explain how quantum computing works."
  model = "gpt-4o"
  
  # Control the response generation
  temperature       = 0.5
  max_output_tokens = 300
  instructions      = "Explain at a high school level, avoiding technical jargon."
  
  # Preserve the response when parameters change (prevent recreation)
  preserve_on_change = true
  
  # User identifier for tracking
  user = "example-user-id"
}

# Output the generated text
output "story" {
  value = openai_model_response.simple.output.text
}

# Output token usage statistics
output "token_usage" {
  value = openai_model_response.advanced.usage
}
```

## Immutability and the `preserve_on_change` Flag

By default, model responses are immutable - once created, they can't be modified. Any change to the input parameters will cause Terraform to destroy the existing resource and create a new one, generating a new response from the OpenAI API.

If you want to prevent recreation of the resource when parameters change (which would result in a new API call), you can set `preserve_on_change = true`. This has the following effects:

1. Changes to parameters will still be reflected in the Terraform state, but will not trigger recreation of the resource.
2. The OpenAI API will not be called again, so the actual response text will remain the same.
3. Terraform will show drift detection between the configuration and the actual resource.

This flag is useful when you want to maintain the same response text across configuration changes, or when you're experimenting with different parameters but don't want to generate a new response every time.

## Argument Reference

* `input` - (Required) The input text or prompt to generate a response for.
* `model` - (Required) ID of the model to use (e.g., "gpt-4o", "gpt-4-turbo").
* `max_output_tokens` - (Optional) The maximum number of tokens to generate.
* `temperature` - (Optional) Sampling temperature between 0 and 2. Higher values mean more randomness. Defaults to 0.7.
* `top_p` - (Optional) Nucleus sampling parameter between 0 and 1. Controls diversity.
* `top_k` - (Optional) Top-k sampling parameter. Only considers the top k tokens.
* `include` - (Optional) List of fields to include in the response.
* `instructions` - (Optional) Additional instructions to guide the model's response.
* `stop_sequences` - (Optional) List of sequences where the API will stop generating further tokens.
* `frequency_penalty` - (Optional) Penalty for token frequency between -2.0 and 2.0.
* `presence_penalty` - (Optional) Penalty for token presence between -2.0 and 2.0.
* `user` - (Optional) A unique identifier representing the end-user, to help track and detect abuse.
* `preserve_on_change` - (Optional) If true, prevents recreation when parameters change. Will show drift but preserve the existing response. Defaults to false.

## Attributes Reference

* `id` - The unique identifier for this response.
* `object` - Object type (usually "model_response").
* `created` - Unix timestamp when the response was created.
* `output` - The generated output containing:
  * `text` - The generated text response.
  * `role` - The role of the entity providing the response (usually "assistant").
* `usage` - Token usage statistics for the request:
  * `prompt_tokens` - Number of tokens in the input prompt.
  * `completion_tokens` - Number of tokens in the generated completion.
  * `total_tokens` - Total number of tokens used.
* `finish_reason` - Reason why the response finished (e.g., "stop", "length", "content").

## Relationship with Data Sources

The OpenAI Terraform provider offers both a resource (`openai_model_response`) and a data source (`data.openai_model_response`) for working with model responses. Here's how they differ:

1. **Resources vs Data Sources**:
   - **Resource**: Creates a new model response by making an API call to OpenAI. It's used when you want to generate new content.
   - **Data Source**: Retrieves an existing model response from OpenAI. It's read-only and doesn't create anything new.

2. **Use Cases**:
   - Use the resource when you need to generate new text with OpenAI models
   - Use the data source when you need to reference an existing response (created outside Terraform or by another resource)

3. **Common Pattern**: Create with resource, reference with data source:
   ```hcl
   # Create a response
   resource "openai_model_response" "story" {
     input = "Tell me a story"
     model = "gpt-4o"
   }
   
   # Reference it with a data source (if needed in another module)
   data "openai_model_response" "story_data" {
     response_id = openai_model_response.story.id
   }
   ```

4. **Related Data Sources**:
   - `openai_model_response` - Gets a single response by ID
   - `openai_model_responses` - Lists multiple responses with filtering
   - `openai_model_response_input_items` - Gets input items for a specific response

See the [Data Sources section](../data-sources/) for more details on using these data sources.

## Import

Model responses can be imported using their ID:

```shell
terraform import openai_model_response.example resp_67ebc9caabf48191bad495f24084ca270d4ded209c82d6c7
```

When importing a model response:

1. The provider automatically retrieves all fields from the OpenAI API, including the original `input` prompt
2. The imported resource is marked with `imported = true` and `preserve_on_change = true` to prevent accidental recreation
3. You don't need to specify the original input parameters in your configuration - they are preserved from the API
4. Any changes to the configuration will show as drift but won't cause the resource to be recreated (preserving the original output)

This allows you to easily import existing model responses without needing to remember or specify the exact prompts that were used to create them.

## Related Resources

* [openai_model_response Data Source](../data-sources/model_response)
* [openai_model_responses Data Source](../data-sources/model_responses)
* [openai_model_response_input_items Data Source](../data-sources/model_response_input_items) 