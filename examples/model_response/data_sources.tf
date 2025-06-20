# NOTE: The data source examples are commented out
# because the example response ID is not available or valid.
# To use this example, replace the response_id with a valid ID from your OpenAI account.

# Example of retrieving a specific model response by ID
# data "openai_model_response" "existing_response" {
#   response_id = "resp_67ebbaa0adf88191a710dfcae1ea529b0a8163aa6dff32fa" # Replace with a valid response ID
# }
# 
# # Output the retrieved response details
# output "retrieved_response" {
#   value = data.openai_model_response.existing_response.output.text
# }
# 
# # Output the usage information
# output "retrieved_response_usage" {
#   value = data.openai_model_response.existing_response.usage
# }
# 
# # Output the input items (prompts) that generated this response
# output "retrieved_response_input" {
#   value = data.openai_model_response.existing_response.input_items
# }

# NOTE: The openai_model_responses data source is commented out because
# OpenAI requires browser session authentication for the responses list endpoint.
# This endpoint can't be accessed with a regular API key.
#
# Example of listing multiple model responses (requires browser session auth)
# data "openai_model_responses" "recent_responses" {
#   limit = 5
#   order = "desc" # Most recent first
#   
#   # Optional filters
#   # filter_by_user = "user123"
#   # after = "resp_abc123"
#   # before = "resp_xyz789"
# }
# 
# # Output the list of recent responses
# output "recent_responses" {
#   value = data.openai_model_responses.recent_responses.responses
# }
# 
# # Check if there are more responses to fetch
# output "has_more_responses" {
#   value = data.openai_model_responses.recent_responses.has_more
# }

# Example of retrieving a specific model response by ID
# To use this example, replace the response_id with a valid ID from your OpenAI account
/*
data "openai_model_response" "existing_response" {
  response_id = "resp_67ebbaa0adf88191a710dfcae1ea529b0a8163aa6dff32fa" # Replace with your actual response ID

  # Optionally include additional fields
  # include = ["usage.input_tokens_details", "usage.output_tokens_details"]
}

# Output the retrieved response details
output "retrieved_response" {
  value = data.openai_model_response.existing_response.output.text
}

# Output the usage information
output "retrieved_response_usage" {
  value = data.openai_model_response.existing_response.usage
}

# Output the input items
output "retrieved_response_input_items" {
  value = data.openai_model_response.existing_response.input_items
}
*/ 