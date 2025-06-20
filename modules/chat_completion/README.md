# OpenAI Chat Completion Module

This Terraform module provides a wrapper around the OpenAI Chat Completions API, allowing you to generate conversational text using OpenAI's state-of-the-art language models.

## Features

- Generate conversational responses using OpenAI's powerful models (GPT-4, GPT-3.5-turbo, etc.)
- Easily configure message sequences with different roles (system, user, assistant)
- Fine-tune generation parameters (temperature, max tokens, etc.)
- Support for function calling
- Proper error handling and state management

## Usage

### Basic Chat Completion

```hcl
module "chat" {
  source = "path/to/modules/chat_completion"
  
  model = "gpt-3.5-turbo"
  
  messages = [
    {
      role    = "system"
      content = "You are a helpful assistant."
    },
    {
      role    = "user"
      content = "Tell me about the solar system."
    }
  ]
  
  temperature = 0.7
  max_tokens  = 500
}

output "assistant_response" {
  value = module.chat.content
}
```

### Function Calling Example

```hcl
module "chat_with_functions" {
  source = "path/to/modules/chat_completion"
  
  model = "gpt-4"
  
  messages = [
    {
      role    = "system"
      content = "You are an assistant that can call functions to get information."
    },
    {
      role    = "user"
      content = "What's the weather in San Francisco?"
    }
  ]
  
  functions = [
    {
      name        = "get_weather"
      description = "Get the current weather in a location"
      parameters  = jsonencode({
        type = "object",
        properties = {
          location = {
            type        = "string",
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
  ]
  
  function_call = "auto"
}

output "function_call" {
  value = module.chat_with_functions.function_call
}
```

### Multi-Turn Conversation

```hcl
module "conversation" {
  source = "path/to/modules/chat_completion"
  
  model = "gpt-4o"
  
  messages = [
    {
      role    = "system"
      content = "You are a helpful assistant that provides concise responses."
    },
    {
      role    = "user"
      content = "Hello, how are you?"
    },
    {
      role    = "assistant"
      content = "I'm doing well, thank you for asking! How can I assist you today?"
    },
    {
      role    = "user"
      content = "Can you recommend a book?"
    }
  ]
  
  temperature = 0.8
}

output "recommendation" {
  value = module.conversation.content
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| `model` | ID of the model to use (e.g., gpt-4o, gpt-4, gpt-3.5-turbo) | `string` | n/a | yes |
| `messages` | A list of messages comprising the conversation context | `list(object)` | n/a | yes |
| `functions` | A list of functions the model may generate JSON inputs for | `list(object)` | `null` | no |
| `function_call` | Controls how the model responds to function calls | `string` | `null` | no |
| `temperature` | What sampling temperature to use (0-2) | `number` | `1.0` | no |
| `max_tokens` | Maximum tokens to generate | `number` | `null` | no |
| `top_p` | Nucleus sampling parameter (0-1) | `number` | `1.0` | no |
| `n` | Number of chat completion choices to generate | `number` | `1` | no |
| `stop` | Up to 4 sequences where the API will stop generating further tokens | `list(string)` | `null` | no |
| `presence_penalty` | Penalty for new tokens based on presence (-2.0 to 2.0) | `number` | `0` | no |
| `frequency_penalty` | Penalty for new tokens based on frequency (-2.0 to 2.0) | `number` | `0` | no |
| `user` | A unique identifier representing your end-user | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| `id` | The ID of the chat completion |
| `created` | The Unix timestamp of when the completion was created |
| `model` | The model used for the completion |
| `choices` | The generated completions |
| `content` | The content of the first generated message |
| `finish_reason` | The reason why the generation finished |
| `function_call` | The function call in the first generated message, if any |
| `usage` | Token usage statistics |

## Message Format

Each message in the `messages` list requires the following format:

```hcl
{
  role    = string # One of: "system", "user", "assistant", "function", "developer"
  content = string # The content of the message
  name    = string # Optional: The name of the author (required if role is "function")
  function_call = {  # Optional: If the message contains a function call
    name      = string # The name of the function to call
    arguments = string # The arguments to pass to the function as a JSON string
  }
}
```

## Function Format

Each function in the `functions` list requires the following format:

```hcl
{
  name        = string # The name of the function
  description = string # Optional: A description of what the function does
  parameters  = string # The parameters the function accepts, described as a JSON Schema object
}
```

## Chat Completions Store

OpenAI offers a "Chat Completions Store" feature that allows you to retrieve chat completions after they've been created. To use this feature:

1. **Check Feature Availability**: The Chat Completions Store feature must be enabled on your OpenAI account. This is a relatively new and experimental feature that is not available by default for all accounts.

2. **Use a Compatible Model**: You must use a model that supports storage, such as `gpt-4o` or `gpt-4-1106-preview`.

3. **Enable Storage**: Set the `store` parameter to `true` in your configuration:

```hcl
module "storable_chat" {
  source = "path/to/modules/chat_completion"
  
  model = "gpt-4o"
  store = true
  
  messages = [
    {
      role    = "system"
      content = "You are a helpful assistant."
    },
    {
      role    = "user"
      content = "What's the capital of France?"
    }
  ]
}
```

4. **Add Metadata (Optional)**: You can add metadata to your chat completions for filtering when listing them later:

```hcl
module "storable_chat_with_metadata" {
  source = "path/to/modules/chat_completion"
  
  model = "gpt-4o"
  store = true
  
  metadata = {
    category = "geography"
    user_id  = "user123"
  }
  
  messages = [
    # ... messages ...
  ]
}
```

If all these conditions are met, you can retrieve the chat completion later using the data sources provided by the OpenAI Terraform provider:

```hcl
data "openai_chat_completion" "retrieved" {
  completion_id = module.storable_chat.id
}

output "retrieved_content" {
  value = data.openai_chat_completion.retrieved.choices[0].message[0].content
}
```

## Notes

- Each chat completion request creates a new, stateless conversation. To simulate a continuing conversation, include the full message history in each request.
- The module automatically handles the conversion between Terraform data structures and the format required by the OpenAI API.
- For production use, implement proper error handling and retry logic in your application code.
- Be aware of token limits for different models. If you exceed the model's token limit, the request will fail.

## Common API Errors

### Token Limit Exceeded
If your messages are too long, you may encounter token limit errors. Consider using a model with higher token limits or reducing the length of your messages.

### Rate Limiting
If you're making too many requests, you may encounter rate limit errors. Implement exponential backoff and retry logic in your application code.

## Related Resources

- [OpenAI Chat Completions API Documentation](https://platform.openai.com/docs/api-reference/chat)
- [OpenAI Chat Models Documentation](https://platform.openai.com/docs/models/overview) 