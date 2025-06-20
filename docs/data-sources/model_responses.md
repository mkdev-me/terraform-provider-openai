---
page_title: "OpenAI: openai_model_responses Data Source"
subcategory: "Model Responses"
description: |-
  Lists multiple model responses from the OpenAI API with pagination and filtering options.
---

# openai_model_responses Data Source

This data source allows you to retrieve a list of model responses from the OpenAI API. You can paginate through results and apply filters based on creation time or user. This is useful for monitoring or auditing the model responses generated in your account.

## Example Usage

```hcl
# Retrieve the 10 most recent model responses
data "openai_model_responses" "recent" {
  limit = 10
  order = "desc"  # Newest first
}

# Filter responses by a specific user
data "openai_model_responses" "user_responses" {
  limit = 20
  filter_by_user = "example-user-id"
}

# Paginate through responses
data "openai_model_responses" "paginated" {
  limit = 5
  after = "resp_67ebc9caabf48191bad495f24084ca270d4ded209c82d6c7"
}

# Output the response IDs
output "response_ids" {
  value = [for resp in data.openai_model_responses.recent.responses : resp.id]
}
```

## Argument Reference

* `limit` - (Optional) Maximum number of responses to return. Defaults to 20, maximum is 100.
* `order` - (Optional) Sort order for responses. Can be "asc" (oldest first) or "desc" (newest first). Defaults to "desc".
* `after` - (Optional) Return responses after this response ID (for pagination).
* `before` - (Optional) Return responses before this response ID (for pagination).
* `filter_by_user` - (Optional) Filter responses by user ID.

## Attributes Reference

* `responses` - A list of model responses. Each response contains:
  * `id` - The unique identifier of the model response.
  * `created_at` - The timestamp when the response was created.
  * `model` - The model used to generate the response.
  * `status` - The status of the response (e.g., "completed").
  * `output` - A map containing the generated output.
* `has_more` - Boolean indicating if there are more responses available beyond the current page.

## Related Resources

* [openai_model_response Resource](../resources/model_response)
* [openai_model_response Data Source](./model_response)
* [openai_model_response_input_items Data Source](./model_response_input_items) 