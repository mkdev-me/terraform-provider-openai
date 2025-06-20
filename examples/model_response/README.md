# OpenAI Model Response Examples

This directory contains examples of using the OpenAI Model Response resources and data sources with Terraform.

## Resources

- `openai_model_response` - Create a response using OpenAI's generative models

## Data Sources

- `openai_model_response` - Retrieve information about a specific model response by ID
- `openai_model_responses` - List multiple model responses with filtering and pagination options
- `openai_model_response_input_items` - Retrieve the input items from a specific model response

## Examples

### Resource Examples

- `main.tf` - Basic example of creating model responses
- `input_items.tf` - Example of using structured input items with model responses

### Data Source Examples

- `model_response_data_source.tf` - Examples of retrieving a specific model response
- `model_responses_data_source.tf` - Examples of listing multiple model responses
- `model_response_input_items_data_source.tf` - Examples of retrieving input items for a response

## Data Source Usage Details

### Retrieving a Single Model Response

The `openai_model_response` data source allows you to retrieve a specific model response by its ID:

```hcl
data "openai_model_response" "specific_response" {
  response_id = "resp_67ed97f0cb58819194b0cea77d2b675403a2738d77e56696" # Replace with actual ID
  
  # Optionally include additional details
  # include = ["usage.input_tokens_details", "usage.output_tokens_details"]
}

# Access the generated text
output "response_text" {
  value = data.openai_model_response.specific_response.output.text
}

# Access token usage statistics
output "response_usage" {
  value = data.openai_model_response.specific_response.usage
}

# Access input prompts
output "response_input" {
  value = data.openai_model_response.specific_response.input_items
}
```

You can also use the ID from an existing resource:

```hcl
# First create a resource
resource "openai_model_response" "bedtime_story" {
  model  = "gpt-4o"
  input  = "Tell me a bedtime story"
  # ... other parameters
}

# Then reference it with a data source
data "openai_model_response" "bedtime_story_data" {
  response_id = openai_model_response.bedtime_story.id
}

# Both the resource and data source provide the same output
output "story_from_resource" {
  value = openai_model_response.bedtime_story.output.text
}

output "story_from_data_source" {
  value = data.openai_model_response.bedtime_story_data.output.text
}
```

### Retrieving Input Items for a Response

The `openai_model_response_input_items` data source allows you to retrieve just the input items used to generate a response:

```hcl
data "openai_model_response_input_items" "example_inputs" {
  response_id = "resp_67ed97f0cb58819194b0cea77d2b675403a2738d77e56696" # Replace with actual ID
  
  # Optional pagination parameters
  # limit = 10
  # order = "desc" # Most recent first (default is "asc")
  # after = "msg_abc123"  # For pagination
  # before = "msg_xyz789" # For pagination
}

# Access all input items
output "all_input_items" {
  value = data.openai_model_response_input_items.example_inputs.input_items
}

# Access first input content
output "first_input_content" {
  value = length(data.openai_model_response_input_items.example_inputs.input_items) > 0 ? 
    data.openai_model_response_input_items.example_inputs.input_items[0].content : 
    "no input items"
}

# Check if there are more items to fetch
output "has_more_items" {
  value = data.openai_model_response_input_items.example_inputs.has_more
}
```

### Listing Multiple Model Responses

The `openai_model_responses` data source allows you to list multiple responses with filtering options:

```hcl
data "openai_model_responses" "recent_responses" {
  limit = 5
  order = "desc" # Most recent first
  
  # Optional filtering
  # filter_by_user = "user123"
  
  # Optional pagination
  # after = "resp_abc123"
  # before = "resp_xyz789"
}

# Access list of responses
output "recent_responses_list" {
  value = data.openai_model_responses.recent_responses.responses
}

# Check if there are more responses
output "has_more_responses" {
  value = data.openai_model_responses.recent_responses.has_more
}
```

**Note**: This data source requires browser session authentication and cannot be used with an API key.

## Authentication Notes

The `openai_model_responses` data source requires browser session authentication, as it cannot be accessed with a regular API key.

## Usage

To run these examples:

1. Set your OpenAI API key as an environment variable:
   ```
   export OPENAI_API_KEY="your-api-key"
   ```

2. Initialize Terraform:
   ```
   terraform init
   ```

3. Apply the configuration:
   ```
   terraform apply
   ```

Note: Replace any placeholder response IDs in the examples with actual IDs from your OpenAI account.

## Response ID Format

Response IDs typically follow the format: `resp_67ebbaa0adf88191a710dfcae1ea529b0a8163aa6dff32fa`

## OpenAI API Endpoints

This Terraform provider uses the following OpenAI API endpoints:

```
POST https://api.openai.com/v1/responses                  # Create a model response
GET https://api.openai.com/v1/responses/{id}              # Retrieve a model response
GET https://api.openai.com/v1/responses/{id}/input_items  # Retrieve input items for a response
GET https://api.openai.com/v1/responses                   # List model responses (browser auth)
```

## Run the Example

Set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

Then execute the Terraform commands:

```bash
terraform init
terraform plan
terraform apply
```

## Importing Existing Model Responses

You can import existing OpenAI model responses into your Terraform state with:

```bash
terraform import openai_model_response.resource_name response_id
```

For example:
```bash
terraform import openai_model_response.bedtime_story resp_67edb65336b881919b5703f0a37064fa07d30f4d397146c6
```

The import process automatically:

1. Retrieves all response data from the OpenAI API, including the original input prompt
2. Preserves the original response output, preventing accidental recreation
3. Sets `imported = true` and `preserve_on_change = true` to maintain the resource state
4. Allows you to manage the resource without needing to know the exact inputs used to create it

After importing, you can verify with:
```bash
terraform state show openai_model_response.resource_name
```

If you want to make changes to your configuration file after importing, the provider will detect drift but won't recreate the resource, preserving your original output.

## Output

After applying the configuration, Terraform will output the generated text responses and any data from queried data sources.

Example outputs:
- `robot_story` - A bedtime story about a robot and his dog
- `rainbow_explanation` - An explanation of how rainbows form
- `mars_poem` - A poem about terraforming Mars
- `response_text` - Text retrieved from a data source
- `all_input_items` - Input items retrieved from a data source
- `token_usage` - Token usage statistics for the responses