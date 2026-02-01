resource "openai_response" "full_example" {
  model = "gpt-5.2"
  input = "What is the weather in San Francisco?"

  # Context
  instructions = "You are a helpful assistant."
  # conversation_id = "conv_12345" # Optional: To continue a conversation

  # Tuning
  temperature       = 0.7
  top_p             = 1.0
  max_output_tokens = 100
  truncation        = "auto"
  reasoning_effort  = "medium"

  # Metadata
  metadata = {
    "environment" = "test"
    "user_id"     = "user_123"
  }

  # Tools
  tools = [
    {
      type = "function"
      function = {
        name        = "get_weather"
        description = "Get the current weather in a given location"
        parameters = jsonencode({
          type = "object"
          properties = {
            location = {
              type        = "string"
              description = "The city and state, e.g. San Francisco, CA"
            }
            unit = {
              type = "string"
              enum = ["celsius", "fahrenheit"]
            }
          }
          required = ["location"]
        })
      }
    },
    {
      type = "web_search"
    }
  ]
  tool_choice         = "auto"
  parallel_tool_calls = true

  # Output Format
  response_format = "text" # or json_object, json_schema
  include         = ["web_search_call.action.sources"]
}

output "full_example_output" {
  value = openai_response.full_example.content
}
