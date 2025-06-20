---
page_title: "OpenAI: openai_completion Resource"
subcategory: ""
description: |-
  Generates text completions using OpenAI's legacy completion models.
---

# openai_completion Resource

The `openai_completion` resource generates text completions using OpenAI's legacy completion models. This resource is most useful for older models like `text-davinci-003` that don't use the chat interface. For newer models, the `openai_chat_completion` resource is recommended.

## Example Usage

```hcl
resource "openai_completion" "example" {
  model       = "text-davinci-003"
  prompt      = "Summarize the main principles of DevOps in a paragraph:"
  max_tokens  = 150
  temperature = 0.7
  n           = 1
}

output "completion_text" {
  value = openai_completion.example.choices[0].text
}

# Generate multiple completions from the same prompt
resource "openai_completion" "multiple" {
  model       = "text-davinci-003"
  prompt      = "Write a haiku about terraform:"
  max_tokens  = 50
  temperature = 0.9
  n           = 3
}

output "haikus" {
  value = openai_completion.multiple.choices[*].text
}
```

## Argument Reference

* `model` - (Required) The ID of the model to use for completion. Common models include:
  * `text-davinci-003`
  * `text-davinci-002`
  * `text-curie-001`
  * `text-babbage-001`
  * `text-ada-001`
* `prompt` - (Required) The prompt to generate completions for.
* `suffix` - (Optional) The suffix that comes after a completion.
* `max_tokens` - (Optional) The maximum number of tokens to generate. Defaults to 16.
* `temperature` - (Optional) Controls randomness: 0.0 means deterministic, 1.0 means more random. Range: 0.0 to 2.0. Defaults to 1.0.
* `top_p` - (Optional) Controls diversity via nucleus sampling. Range: 0.0 to 1.0. Defaults to 1.0.
* `n` - (Optional) How many completions to generate. Defaults to 1.
* `stream` - (Optional) Whether to stream back partial progress. Not usable with Terraform, always set to false.
* `logprobs` - (Optional) Include the log probabilities on the most likely tokens. Range: 0 to 5.
* `echo` - (Optional) Echo back the prompt in addition to the completion. Defaults to false.
* `stop` - (Optional) Sequences where the API will stop generating further tokens.
* `presence_penalty` - (Optional) Penalizes new tokens based on whether they appear in the text so far. Range: -2.0 to 2.0.
* `frequency_penalty` - (Optional) Penalizes new tokens based on their frequency in the text so far. Range: -2.0 to 2.0.
* `best_of` - (Optional) Generates best_of completions server-side and returns the best one. Defaults to 1.
* `logit_bias` - (Optional) Map of token IDs to bias values from -100 to 100.
* `user` - (Optional) A unique identifier representing your end-user.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A unique identifier for this completion resource.
* `choices` - A list of generated completions:
  * `text` - The generated completion text.
  * `index` - The index of this completion in the list.
  * `logprobs` - Log probability information if requested.
  * `finish_reason` - The reason why the completion finished, can be "stop", "length", etc.
* `usage` - Information about token usage:
  * `prompt_tokens` - The number of tokens in the prompt.
  * `completion_tokens` - The number of tokens in the generated completion.
  * `total_tokens` - The total number of tokens used (prompt + completion).
* `created` - The timestamp when the completion was created.
* `model` - The model used for the completion.
* `object` - The object type, always "text_completion".

## Import

Completion resources cannot be imported because they represent one-time API calls rather than persistent resources. 