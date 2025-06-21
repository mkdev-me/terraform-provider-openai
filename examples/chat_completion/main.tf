terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
    }
  }
}

# Variable to control whether to attempt to use Chat Completions Store features
# Set this to true only if you have the feature enabled in your OpenAI account
variable "enable_chat_completions_store" {
  description = "Set to true only if you have the Chat Completions Store feature enabled in your OpenAI account"
  type        = bool
  default     = true
}

provider "openai" {
  # API key is set from the OPENAI_API_KEY environment variable by default
}

# Example 1: Basic conversation with GPT
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
      content = "Explain the concept of neural networks to me like I'm 10 years old."
    }
  ]

  temperature = 0.7
  max_tokens  = 300
  imported    = false
}

# Example 2: More advanced conversation with multiple turns and GPT-4
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
      content = "Climate change refers to long-term shifts in temperatures and weather patterns. These changes may be natural, such as through variations in the solar cycle. But since the 1800s, human activities have been the main driver of climate change, primarily due to burning fossil fuels like coal, oil, and gas, which produces heat-trapping gases."
    },
    {
      role    = "user"
      content = "What can I do to help reduce climate change?"
    }
  ]

  temperature = 0.6
  max_tokens  = 500
  imported    = false
}

# Example 3: Chat with function calling
# This example doesn't create a new completion when we make changes due to our imported flag
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
      parameters = jsonencode({
        type = "object",
        properties = {
          departure_city = {
            type        = "string",
            description = "The city to depart from"
          },
          arrival_city = {
            type        = "string",
            description = "The city to arrive at"
          },
          departure_date = {
            type        = "string",
            description = "The date of departure in YYYY-MM-DD format"
          },
          return_date = {
            type        = "string",
            description = "Optional: The return date in YYYY-MM-DD format for round trips"
          },
          num_passengers = {
            type        = "integer",
            description = "The number of passengers"
          },
          class = {
            type        = "string",
            enum        = ["economy", "premium_economy", "business", "first"],
            description = "The class of travel"
          }
        },
        required = ["departure_city", "arrival_city", "departure_date", "num_passengers"]
      })
    }
  ]

  function_call = "auto"
  temperature   = 0.2
  imported      = false
}

# Example 4: Using the newest model with specific settings
module "advanced_chat" {
  source = "../../modules/chat_completion"

  model = "gpt-4o"

  messages = [
    {
      role    = "system"
      content = "You are a creative writing assistant that specializes in crafting compelling story beginnings."
    },
    {
      role    = "user"
      content = "Write the opening paragraph for a science fiction story set 200 years in the future where humanity has colonized Mars."
    }
  ]

  temperature       = 0.9
  max_tokens        = 1000
  top_p             = 0.95
  frequency_penalty = 0.5
  presence_penalty  = 0.5
  imported          = false
}

# Note: These data source examples require the Chat Completions Store feature to be enabled
# on your OpenAI account. This is a relatively new and experimental feature.
# Additionally, when creating chat completions that you want to retrieve later, you must:
# 1. Use a compatible model (e.g., gpt-4o)
# 2. Set the `store` parameter to true
# 3. Have the Chat Completions Store feature enabled on your account

#--------------------------------------------------------------
# Example 1: Creating a chat completion with storage enabled
#--------------------------------------------------------------

# First, create a chat completion with storage enabled for later retrieval
resource "openai_chat_completion" "stored_completion" {
  model    = "gpt-4o" # Use a compatible model
  store    = true     # Required for storage and later retrieval
  imported = false    # Prevent recreation when attributes change

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

# Output the completion ID for reference
output "completion_id" {
  value = openai_chat_completion.stored_completion.id
}

#--------------------------------------------------------------
# Example 2: Retrieving a chat completion by ID
#--------------------------------------------------------------

# Important: This will only work if Chat Completions Store is enabled on your account
data "openai_chat_completion" "retrieved_completion" {
  count = var.enable_chat_completions_store ? 1 : 0

  completion_id = openai_chat_completion.stored_completion.id
}

output "retrieved_assistant_response" {
  value = var.enable_chat_completions_store ? try(
    data.openai_chat_completion.retrieved_completion[0].choices[0].message[0].content,
    "No content available - chat completion may have expired"
  ) : "Chat Completions Store feature not enabled"
}

#--------------------------------------------------------------
# Example 3: Retrieving messages from a chat completion
#--------------------------------------------------------------

# Important: This will only work if Chat Completions Store is enabled on your account
data "openai_chat_completion_messages" "completion_messages" {
  count = var.enable_chat_completions_store ? 1 : 0

  completion_id = openai_chat_completion.stored_completion.id
  limit         = 10
  order         = "asc" # Oldest messages first
}

output "retrieved_messages" {
  value = var.enable_chat_completions_store ? (
    length(jsondecode(jsonencode(try(data.openai_chat_completion_messages.completion_messages[0].messages, [])))) > 0 ?
    jsonencode(try(data.openai_chat_completion_messages.completion_messages[0].messages, [])) :
    "No messages available - chat completion may have expired"
  ) : "Chat Completions Store feature not enabled"
}

#--------------------------------------------------------------
# Example 4: Listing multiple chat completions with filters
#--------------------------------------------------------------

# Important: This will only work if Chat Completions Store is enabled on your account
data "openai_chat_completions" "recent_completions" {
  count = var.enable_chat_completions_store ? 1 : 0

  limit = 5
  order = "desc" # Most recent first

  # Optional: Filter by metadata
  metadata = {
    category = "terraform_info"
  }
}

output "recent_completions_list" {
  value = var.enable_chat_completions_store ? (
    length(jsondecode(jsonencode(try(data.openai_chat_completions.recent_completions[0].chat_completions, [])))) > 0 ?
    jsonencode(try(data.openai_chat_completions.recent_completions[0].chat_completions, [])) :
    "No chat completions available - they may have expired"
  ) : "Chat Completions Store feature not enabled"
}

#--------------------------------------------------------------
# Example 5: Using output attributes from stored chat completion resource
#--------------------------------------------------------------

# This approach works regardless of Chat Completions Store availability
# as it uses Terraform state, not the OpenAI API
output "immediate_assistant_response" {
  value = openai_chat_completion.stored_completion.choices[0].message[0].content
}

output "token_usage" {
  value = openai_chat_completion.stored_completion.usage
}

