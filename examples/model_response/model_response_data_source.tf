# Example 1: Retrieve a specific model response by ID
# Using actual response ID from our current resources
# data "openai_model_response" "specific_response" {
#   response_id = "resp_67ed97f0cb58819194b0cea77d2b675403a2738d77e56696"
#   include     = ["usage.input_tokens_details", "usage.output_tokens_details"]
# }
# 
# # Output the retrieved response details
# output "response_text" {
#   description = "The text of the specific response"
#   value       = data.openai_model_response.specific_response.output.text
# }
# 
# # Output the usage information
# output "response_usage" {
#   description = "Usage stats for the specific response"
#   value       = data.openai_model_response.specific_response.usage
# }

# Example 2: Retrieve a model response by ID with included fields
# Note: This is commented out because it requires a valid response ID from your account
/*
data "openai_model_response" "response_with_details" {
  response_id = "resp_67ebbaa0adf88191a710dfcae1ea529b0a8163aa6dff32fa" # Replace with a valid response ID
  
  # Include additional usage details (using validated values)
  include = ["usage.input_tokens_details", "usage.output_tokens_details"]
}

# Output the model used
output "response_model" {
  value = data.openai_model_response.response_with_details.model
}

# Output the creation timestamp
output "response_created_at" {
  value = data.openai_model_response.response_with_details.created_at
}

# Output the status
output "response_status" {
  value = data.openai_model_response.response_with_details.status
}
*/ 