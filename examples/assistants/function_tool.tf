# Example of an assistant with a function tool
resource "openai_assistant" "function_assistant" {
  name         = "Weather Assistant"
  model        = "gpt-4o"
  instructions = "You are a weather assistant that can help users check the weather."
  description  = "Assistant that can provide weather information"

  # Configure a function tool with proper syntax
  tools {
    type = "function"

    function {
      name        = "get_weather"
      description = "Get the current weather in a given location"
      parameters = jsonencode({
        type = "object",
        properties = {
          location = {
            type        = "string",
            description = "The city and state, e.g., San Francisco, CA"
          },
          unit = {
            type        = "string",
            enum        = ["celsius", "fahrenheit"],
            description = "The unit of temperature to use"
          }
        },
        required = ["location"]
      })
    }
  }

  metadata = {
    "created_by" = "terraform",
    "purpose"    = "weather_info"
  }
}

# Output the function assistant ID
output "function_assistant_id" {
  description = "The ID of the created function assistant"
  value       = openai_assistant.function_assistant.id
} 