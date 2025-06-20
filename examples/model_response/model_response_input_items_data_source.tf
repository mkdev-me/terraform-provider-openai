# Example: Retrieve input items from a specific model response
# Using actual response ID from our current resources
# data "openai_model_response_input_items" "example_inputs" {
#   response_id = "resp_67ed97f0cb58819194b0cea77d2b675403a2738d77e56696" # bedtime_story response ID
# }
# 
# # Output all input items
# output "all_input_items" {
#   description = "All input items for the response"
#   value       = data.openai_model_response_input_items.example_inputs.input_items
# }
# 
# # Output the first input item content
# output "first_input_content" {
#   description = "The content of the first input item"
#   value       = data.openai_model_response_input_items.example_inputs.input_items[0].content
# }
# 
# # Output the first input item ID
# output "first_item_id" {
#   description = "The ID of the first input item"
#   value       = data.openai_model_response_input_items.example_inputs.input_items[0].id
# }
# 
# # Output has_more flag
# output "has_more_items" {
#   description = "Whether there are more items available"
#   value       = data.openai_model_response_input_items.example_inputs.has_more
# }
# 
# # Output roles for all input items
# output "input_roles" {
#   description = "The roles of all input items"
#   value       = [for item in data.openai_model_response_input_items.example_inputs.input_items : item.role]
# }
# 
# # Output response input as a structured object
# output "response_input" {
#   description = "The input used for the response, displayed as an object"
#   value       = data.openai_model_response_input_items.example_inputs.input_items
# } 