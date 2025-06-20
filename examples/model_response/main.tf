terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

provider "openai" {
  # Configuration options
  # API Key can also be provided via environment variable OPENAI_API_KEY
  # api_key = "your-api-key"
}

# This resource will be recreated when parameters change
resource "openai_model_response" "bedtime_story" {
  model              = "gpt-4o-2024-08-06"
  input              = "Tell me a three-sentence bedtime story about a robot and his dog."
  temperature        = 0.5
  max_output_tokens  = 100
  instructions       = "The story should be suitable for young children and have a clear moral lesson."
  preserve_on_change = true
  imported           = true // Set as imported for testing purposes
}

# Output the bedtime story text
output "robot_story" {
  value = lookup(openai_model_response.bedtime_story.output, "text", "No output available")
}

# Output token usage statistics
output "token_usage" {
  value = openai_model_response.bedtime_story.usage
}

# This resource will NOT be recreated when parameters change
resource "openai_model_response" "detailed_response" {
  model              = "gpt-4o-2024-08-06"
  input              = "Explain how rainbows form in simple terms."
  temperature        = 1
  user               = "example-user-id"
  preserve_on_change = true
}

# Output the detailed explanation text
output "rainbow_explanation" {
  value = lookup(openai_model_response.detailed_response.output, "text", "No output available")
}

# Example with simpler parameters 
resource "openai_model_response" "creative_story" {
  model              = "gpt-4o-2024-08-06"
  input              = "Write a short poem about terraforming Mars."
  temperature        = 1
  max_output_tokens  = 150
  preserve_on_change = true
}

# Output the creative story text
output "mars_poem" {
  value = lookup(openai_model_response.creative_story.output, "text", "No output available")
}

# Example of a configuration for an imported resource
# When using terraform import, this configuration is the minimum needed.
# The provider will automatically fetch the input, output, and other details from the API.
#
# First create this configuration:
# resource "openai_model_response" "imported_example" {
#   # No need to specify input or most parameters, they'll be retrieved from the API
#   # You can specify any attributes you want to override from the imported state
#   
#   # These are automatically set by the import process, but you can set them explicitly
#   preserve_on_change = true
# }
#
# Then import with:
# terraform import openai_model_response.imported_example resp_YOUR_RESPONSE_ID
#
# After importing, the resource will have all fields populated from the API,
# including the original input, model, and response output.

# DATA SOURCE EXAMPLES

# Example 1: Retrieve a specific model response using its ID
data "openai_model_response" "bedtime_story_data" {
  # Using the ID of the bedtime_story resource we created above
  response_id = openai_model_response.bedtime_story.id
}

# Output showing the data source works the same as the resource
output "story_from_data_source" {
  value = lookup(data.openai_model_response.bedtime_story_data.output, "text", "No output available")
}

# Example 2: Retrieve input items for a specific model response
data "openai_model_response_input_items" "creative_story_inputs" {
  # Using the ID of the creative_story resource we created above
  response_id = openai_model_response.creative_story.id
}

# Output showing the input items from the data source
output "creative_story_input_items" {
  value = data.openai_model_response_input_items.creative_story_inputs.input_items
}
