---
page_title: "OpenAI: openai_chat_completion Resource"
subcategory: ""
description: |-
  Generates conversational responses using OpenAI's chat models.
---

# openai_chat_completion Resource

The `openai_chat_completion` resource generates conversational responses using OpenAI's chat models like GPT-4 and GPT-3.5-Turbo. This resource provides a more conversational interface than the legacy Completion API and is recommended for most use cases.

## Important Note on Persistence

By default, the OpenAI API does not provide endpoints to retrieve chat completions after they've been created. However, OpenAI does offer a "Chat Completions Store" feature with specific requirements:

### Requirements for using the OpenAI Chat Completions Store:

1. **Compatible model:** You must use a model that supports storage, such as `gpt-4-1106-preview` or `gpt-4o`.
2. **Explicit storage request:** You must set `store = true` in your configuration.
3. **Account feature activation:** The "Chat Completions Store" feature must be enabled on your OpenAI account (this may not be available on all accounts or visible in the dashboard).

If all three conditions are met, you can list your stored completions using the `GET /v1/chat/completions` endpoint.

Without these conditions, the endpoint will return an empty list: `{"object":"list","data":[],"first_id":null,"last_id":null,"has_more":false}`.

Even without the Chat Completions Store feature, the OpenAI Terraform provider stores all chat completion details in the Terraform state, including the full response content. Use outputs or data sources that reference the resource (not the OpenAI API) to access this information.

## Example Usage

```hcl
resource "openai_chat_completion" "example" {
  model = "gpt-4"
  
  message {
    role    = "system"
    content = "You are a helpful assistant that translates English to French."
  }
  
  message {
    role    = "user"
    content = "Translate the following: 'Hello, how are you?'"
  }
  
  temperature = 0.7
  max_tokens  = 100
}

output "translation" {
  value = openai_chat_completion.example.choices[0].message.content
}

# With function calling
resource "openai_chat_completion" "function_call" {
  model = "gpt-4"
  
  message {
    role    = "system"
    content = "You are an assistant that helps users find weather information."
  }
  
  message {
    role    = "user"
    content = "What's the weather like in Boston today?"
  }
  
  tool {
    type    = "function"
    function {
      name        = "get_weather"
      description = "Get the current weather in a given location"
      parameters  = jsonencode({
        type = "object",
        properties = {
          location = {
            type = "string",
            description = "The city and state, e.g. San Francisco, CA"
          },
          unit = {
            type = "string",
            enum = ["celsius", "fahrenheit"]
          }
        },
        required = ["location"]
      })
    }
  }
  
  tool_choice = jsonencode({
    type = "function",
    function = {
      name = "get_weather"
    }
  })
}

output "function_call" {
  value = openai_chat_completion.function_call.choices[0].message.tool_calls
}

# Chat completion with storage enabled
resource "openai_chat_completion" "stored_example" {
  model = "gpt-4"
  store = true  # Indicates to OpenAI to store this completion (for their internal use)
  
  message {
    role    = "system"
    content = "You are a helpful assistant."
  }
  
  message {
    role    = "user"
    content = "Explain the concept of quantum computing in simple terms."
  }
}

# Advanced chat completion with function calling
resource "openai_chat_completion" "function_example" {
  model        = "gpt-4"
  temperature  = 0.2
  function_call = "auto"
  
  message {
    role    = "system"
    content = "You are a helpful assistant."
  }
  
  message {
    role    = "user"
    content = "What's the weather like in San Francisco right now?"
  }
  
  functions {
    name        = "get_weather"
    description = "Get the current weather in a given location"
    parameters  = jsonencode({
      "type": "object",
      "properties": {
        "location": {
          "type": "string",
          "description": "The city and state, e.g. San Francisco, CA"
        },
        "unit": {
          "type": "string",
          "enum": ["celsius", "fahrenheit"]
        }
      },
      "required": ["location"]
    })
  }
}
```

## Argument Reference

* `model` - (Required) The ID of the model to use. Common models include:
  * `gpt-4`
  * `gpt-4-turbo`
  * `gpt-3.5-turbo`
* `message` - (Required) One or more message blocks representing the conversation. Each block supports:
  * `role` - (Required) The role of the message sender. Valid values:
    * `system` - Instructions that set the behavior of the assistant.
    * `user` - The content sent by the user.
    * `assistant` - A previous response from the assistant.
    * `tool` - A message from a tool call.
  * `content` - (Required) The content of the message.
  * `name` - (Optional) The name of the sender, used for user and assistant roles.
  * `tool_call_id` - (Optional) The ID of the tool call that this message is in response to (for tool role).
* `temperature` - (Optional) Controls randomness. Range: 0.0 (deterministic) to 2.0 (more random). Defaults to 1.0.
* `top_p` - (Optional) Controls diversity via nucleus sampling. Range: 0.0 to 1.0. Defaults to 1.0.
* `n` - (Optional) How many chat completion choices to generate. Defaults to 1.
* `stream` - (Optional) Whether to stream back partial progress. Not recommended for Terraform use, always set to false.
* `stop` - (Optional) Sequences where the API will stop generating further tokens.
* `max_tokens` - (Optional) The maximum number of tokens to generate. Defaults to infinity.
* `presence_penalty` - (Optional) Penalizes new tokens based on whether they appear in the text so far. Range: -2.0 to 2.0.
* `frequency_penalty` - (Optional) Penalizes new tokens based on their frequency in the text so far. Range: -2.0 to 2.0.
* `logit_bias` - (Optional) Map of token IDs to bias values from -100 to 100.
* `user` - (Optional) A unique identifier representing your end-user.
* `tool` - (Optional) One or more tool blocks representing tools the model may call. Each block supports:
  * `type` - (Required) The type of tool, currently only "function" is supported.
  * `function` - (Required) A function block representing a function the model may call:
    * `name` - (Required) The name of the function.
    * `description` - (Optional) A description of what the function does.
    * `parameters` - (Required) The parameters the function accepts, specified as a JSON Schema object.
* `tool_choice` - (Optional) Controls which (if any) tool is called by the model. Can be "none", "auto", or a specific tool specification provided as JSON.
* `response_format` - (Optional) Specifies the format of the response. Supports:
  * `type` - (Required) Valid options are "text" or "json_object".
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.
* `store` - (Optional) Whether to store the completion for OpenAI's model distillation or evaluation products. Note that even with this set to `true`, the completion is not retrievable via the API. Default is `false`.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A unique identifier for this chat completion resource.
* `choices` - A list of generated chat completion choices:
  * `index` - The index of this choice in the list.
  * `message` - The message generated by the model:
    * `role` - The role of the generated message, typically "assistant".
    * `content` - The content of the generated message.
    * `tool_calls` - Any tool calls the model decided to make:
      * `id` - The ID of the tool call.
      * `type` - The type of tool call, typically "function".
      * `function` - Details about the function call:
        * `name` - The name of the function being called.
        * `arguments` - The arguments provided to the function.
  * `finish_reason` - The reason why generation finished, can be "stop", "length", "tool_calls", etc.
* `usage` - Information about token usage:
  * `prompt_tokens` - The number of tokens in the prompt.
  * `completion_tokens` - The number of tokens in the generated completion.
  * `total_tokens` - The total number of tokens used (prompt + completion).
* `created` - The timestamp when the chat completion was created.
* `model` - The model used for the chat completion.
* `system_fingerprint` - A unique identifier for the system configuration used.
* `object` - The object type, always "chat.completion".

## Import

Chat completion resources cannot be imported because they represent one-time API calls rather than persistent resources. 