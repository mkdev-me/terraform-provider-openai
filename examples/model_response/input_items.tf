# Example of retrieving input items for a specific model response
# To use this example, replace the response_id with a valid ID from your OpenAI account
# Note: This is commented out because it requires a valid response ID from your account
/*
data "openai_model_response_input_items" "response_inputs" {
  response_id = "resp_67ebbaa0adf88191a710dfcae1ea529b0a8163aa6dff32fa" # Replace with your actual response ID

  # Optional pagination parameters
  # limit = 10
  # order = "desc" # Most recent first (default is "asc")
  # after = "msg_abc123"  # For pagination
  # before = "msg_xyz789" # For pagination

  # Optionally include additional fields
  # include = ["usage"]
}

# Output all the input items
output "response_input_items" {
  value = data.openai_model_response_input_items.response_inputs.input_items
}

# Output the first item ID
output "first_input_item_id" {
  value = data.openai_model_response_input_items.response_inputs.first_id
}

# Output whether there are more items to fetch
output "has_more_input_items" {
  value = data.openai_model_response_input_items.response_inputs.has_more
}
*/ 