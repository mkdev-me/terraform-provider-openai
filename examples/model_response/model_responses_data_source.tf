# NOTE: The openai_model_responses data source requires browser session authentication
# This endpoint can't be accessed with a regular API key.
# The examples below are for reference only and will NOT work with a standard API key.

# Example 1: List recent model responses
# THIS EXAMPLE CANNOT BE EXECUTED with a regular API key - shown for reference only
/*
data "openai_model_responses" "recent_responses" {
  limit = 5
  order = "desc" # Most recent first
}

# Output the list of recent responses
output "recent_responses_list" {
  value = data.openai_model_responses.recent_responses.responses
}

# Check if there are more responses to fetch
output "has_more_responses" {
  value = data.openai_model_responses.recent_responses.has_more
}
*/

# Example 2: Filter responses by user and pagination
# THIS EXAMPLE CANNOT BE EXECUTED with a regular API key - shown for reference only
/*
data "openai_model_responses" "filtered_responses" {
  limit = 10
  order = "desc" # Most recent first
  
  # Filter responses by user
  filter_by_user = "user123" # Replace with actual user ID
  
  # Pagination options
  # after = "resp_abc123" # Get responses after this ID
  # before = "resp_xyz789" # Get responses before this ID
}

# Output the first response ID
output "first_response_id" {
  value = length(data.openai_model_responses.filtered_responses.responses) > 0 ? data.openai_model_responses.filtered_responses.responses[0].id : "no responses"
}

# Extract model IDs from responses
output "response_models" {
  value = [for resp in data.openai_model_responses.filtered_responses.responses : resp.model]
}
*/ 