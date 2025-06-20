# OpenAI Chat Completion Examples

This directory contains examples demonstrating how to use OpenAI chat completion resources and data sources with Terraform.

## Prerequisites

- An OpenAI API key with an active billing account
- Terraform installed

## Setup

1. Set your OpenAI API key as an environment variable:
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

   If you want to try the data sources (requires Chat Completions Store feature):
   ```bash
   terraform apply -var="enable_chat_completions_store=true"
   ```

## Resources Examples

### 1. Basic Chat

A simple conversation with GPT-3.5-turbo:

```hcl
module "basic_chat" {
  source = "../../modules/chat_completion"
  
  model = "gpt-3.5-turbo"
  
  messages = [
    {
      role    = "system"
      content = "You are a helpful assistant specialized in explaining complex topics in simple terms."
    },
    {
      role    = "user"
      content = "Explain the concept of quantum computing to me like I'm 10 years old."
    }
  ]
  
  temperature = 0.7
  max_tokens  = 300
}
```

### 2. Multi-Turn Conversation

A more complex conversation with multiple turns using GPT-4:

```hcl
module "multi_turn_chat" {
  source = "../../modules/chat_completion"
  
  model = "gpt-4"
  
  messages = [
    {
      role    = "system"
      content = "You are an expert on climate change who provides educational information."
    },
    {
      role    = "user"
      content = "What is climate change?"
    },
    {
      role    = "assistant"
      content = "Climate change refers to long-term shifts in temperatures and weather patterns..."
    },
    {
      role    = "user"
      content = "What can I do to help reduce climate change?"
    }
  ]
  
  temperature = 0.6
  max_tokens  = 500
}
```

### 3. Function Calling

Demonstrating the function calling capability with GPT-4:

```hcl
module "function_calling_chat" {
  source = "../../modules/chat_completion"
  
  model = "gpt-4"
  
  messages = [
    {
      role    = "system"
      content = "You are an assistant that helps people book flight tickets."
    },
    {
      role    = "user"
      content = "I want to book a flight from New York to London next week."
    }
  ]
  
  functions = [
    {
      name        = "search_flights"
      description = "Search for available flights between two locations"
      parameters  = jsonencode({
        type = "object",
        properties = {
          departure_city = {
            type        = "string",
            description = "The city to depart from"
          },
          # ... other parameters
        },
        required = ["departure_city", "arrival_city", "departure_date", "num_passengers"]
      })
    }
  ]
  
  function_call = "auto"
  temperature   = 0.2
}
```

### 4. Using the Store Parameter

Create a chat completion that can be stored for later retrieval:

```hcl
resource "openai_chat_completion" "stored_completion" {
  model = "gpt-4o"  # Use a compatible model
  store = true      # Required for storage and later retrieval

  messages {
    role    = "system"
    content = "You are a helpful assistant."
  }

  messages {
    role    = "user"
    content = "What are the top 3 benefits of using Terraform?"
  }

  # Optional metadata for filtering
  metadata = {
    category = "terraform_info"
    user_id  = "example_user"
  }
}
```

## Data Sources Examples

> **Important Note About Chat Completions Store**
>
> The OpenAI Chat Completions Store is a relatively new feature with these requirements:
>
> 1. **Account Activation**: The Chat Completions Store feature must be enabled on your OpenAI account. This is not available by default for all accounts.
>
> 2. **Model Compatibility**: You must use a compatible model such as `gpt-4o` or `gpt-4-1106-preview`.
>
> 3. **Store Parameter**: When creating chat completions, you must explicitly set `store = true`.
>
> 4. **API Access**: You must use the same API key for retrieving as you used for creation.
>
> Without meeting all these conditions, the data source endpoints will return errors.
>
> **Example Usage with Conditional Loading:**
> This example directory uses the `enable_chat_completions_store` variable to conditionally include the data sources:
>
> ```bash
> # To try the data sources (if you have the feature):
> terraform apply -var="enable_chat_completions_store=true"
> ```

### 1. Retrieving a Chat Completion by ID

```hcl
data "openai_chat_completion" "retrieved_completion" {
  completion_id = openai_chat_completion.stored_completion.id
}

output "retrieved_assistant_response" {
  value = data.openai_chat_completion.retrieved_completion.choices[0].message[0].content
}
```

### 2. Retrieving Messages from a Chat Completion

```hcl
data "openai_chat_completion_messages" "completion_messages" {
  completion_id = openai_chat_completion.stored_completion.id
  limit         = 10
  order         = "asc"  # Oldest messages first
}

output "retrieved_messages" {
  value = data.openai_chat_completion_messages.completion_messages.messages
}
```

### 3. Listing Multiple Chat Completions with Filters

```hcl
data "openai_chat_completions" "recent_completions" {
  limit = 5
  order = "desc"  # Most recent first
  
  # Optional: Filter by metadata
  metadata = {
    category = "terraform_info"
  }
}

output "recent_completions_list" {
  value = data.openai_chat_completions.recent_completions.chat_completions
}
```

## Alternative Approach Without Chat Completions Store

If you don't have access to the Chat Completions Store feature, you can use Terraform's state to access the completion data immediately after creation:

```hcl
output "immediate_assistant_response" {
  value = openai_chat_completion.stored_completion.choices[0].message[0].content
}

output "token_usage" {
  value = openai_chat_completion.stored_completion.usage
}
```

## Common API Errors

### Chat Completions Store Not Enabled

**Error:** `The OpenAI API endpoint for retrieving a specific chat completion by ID requires the "Chat Completions Store" feature...`

**Solution:**
- Check if your OpenAI account has the Chat Completions Store feature enabled
- Ensure you're using a compatible model (like gpt-4o)
- Make sure you've set `store = true` in your chat completion resource
- Use the same API key for retrieval as you used for creation

### Other Common Errors

- Invalid API Key: Ensure your API key is correct and active
- Token Limit Exceeded: Reduce message length or use a model with higher limits
- Model Not Available: Verify you have access to the specified model
- Rate Limiting: Implement proper backoff and retry logic

## Production Recommendations

For production use:

1. **Error Handling**: Implement robust error handling for API errors
2. **Alternative Storage**: Consider storing important completions in your own database
3. **Fallback Mechanisms**: Have a plan for when the Chat Completions Store is unavailable
4. **Monitoring**: Track token usage and costs
5. **Security**: Properly secure your API keys and sensitive data

## Additional Resources

- [OpenAI API Documentation](https://platform.openai.com/docs/api-reference/chat)
- [OpenAI Models Overview](https://platform.openai.com/docs/models)
- [Function Calling Guide](https://platform.openai.com/docs/guides/function-calling) 