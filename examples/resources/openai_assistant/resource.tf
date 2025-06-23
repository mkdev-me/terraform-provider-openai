resource "openai_assistant" "example" {
  name         = "Math Tutor"
  instructions = "You are a personal math tutor. Write and run code to answer math questions."
  model        = "gpt-4o-mini"

  tools {
    type = "code_interpreter"
  }

  # Optional: Attach files to the assistant
  # file_ids = ["file-abc123"]

  # Optional: Add metadata
  # metadata = {
  #   team = "education"
  #   version = "1.0"
  # }
}

# Example with function tool
resource "openai_assistant" "function_example" {
  name         = "Weather Assistant"
  instructions = "You are a helpful assistant that can check the weather."
  model        = "gpt-4o-mini"

  tools {
    type = "function"
    function {
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
  }
}

output "assistant_id" {
  value = openai_assistant.example.id
}