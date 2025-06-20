---
page_title: "OpenAI: openai_model_response Data Source"
subcategory: "Model Responses"
description: |-
  Retrieves information about an existing model response from the OpenAI API.
---

# openai_model_response Data Source

This data source allows you to retrieve information about an existing model response from the OpenAI API. Model responses are generated text outputs from OpenAI's language models like GPT-4o. This is useful for fetching details of responses that have been created previously.

## Example Usage

```hcl
# Retrieve a specific model response by ID
data "openai_model_response" "example" {
  response_id = "resp_67ebc9caabf48191bad495f24084ca270d4ded209c82d6c7"
}

# Output the response text
output "response_text" {
  value = data.openai_model_response.example.output.text
}

# Output token usage
output "token_usage" {
  value = data.openai_model_response.example.usage
}
```

## Argument Reference

* `response_id` - (Required) The unique identifier of the model response to retrieve.

## Attributes Reference

* `id` - The unique identifier of the model response.
* `created_at` - The timestamp when the response was created.
* `model` - The model used to generate the response.
* `input_items` - The input provided to generate the response.
* `status` - The status of the response (e.g., "completed").
* `temperature` - The temperature setting used when generating the response.
* `top_p` - The top_p value used when generating the response.
* `output` - A map containing the generated output:
  * `text` - The generated text response.
  * `role` - The role of the entity providing the response (usually "assistant").
* `usage` - A map containing token usage statistics:
  * `prompt_tokens` - The number of tokens in the input prompt.
  * `completion_tokens` - The number of tokens in the generated completion.
  * `total_tokens` - The total number of tokens used.

## Related Resources

* [openai_model_response Resource](../resources/model_response)
* [openai_model_responses Data Source](./model_responses)
* [openai_model_response_input_items Data Source](./model_response_input_items) 