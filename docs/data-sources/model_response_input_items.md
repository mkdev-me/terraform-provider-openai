---
page_title: "OpenAI: openai_model_response_input_items Data Source"
subcategory: "Model Responses"
description: |-
  Retrieves the input items associated with a specific model response from the OpenAI API.
---

# openai_model_response_input_items Data Source

This data source allows you to retrieve the input items associated with a specific model response from the OpenAI API. Input items contain the original prompts or instructions that were provided to generate the response. This is useful for understanding the context of a model response or for auditing purposes.

## Example Usage

```hcl
# Retrieve input items for a specific model response
data "openai_model_response_input_items" "example" {
  response_id = "resp_67ebc9caabf48191bad495f24084ca270d4ded209c82d6c7"
  limit = 10
}

# Output the content of the first input item
output "first_input" {
  value = length(data.openai_model_response_input_items.example.input_items) > 0 ? data.openai_model_response_input_items.example.input_items[0].content : null
}

# Check if there are more input items
output "has_more_input_items" {
  value = data.openai_model_response_input_items.example.has_more
}
```

## Argument Reference

* `response_id` - (Required) The unique identifier of the model response whose input items you want to retrieve.
* `limit` - (Optional) Maximum number of input items to return. Defaults to 20, maximum is 100.
* `order` - (Optional) Sort order for input items. Can be "asc" (oldest first) or "desc" (newest first). Defaults to "asc".
* `after` - (Optional) Return input items after this input item ID (for pagination).
* `before` - (Optional) Return input items before this input item ID (for pagination).
* `include` - (Optional) List of fields to include in the response. Available options include "input".
* `first_id` - (Optional) ID of the first input item to include in the results.
* `last_id` - (Optional) ID of the last input item to include in the results.

## Attributes Reference

* `input_items` - A list of input items. Each input item contains:
  * `id` - The unique identifier of the input item.
  * `type` - The type of the input item (usually "message").
  * `role` - The role of the entity providing the input (usually "user").
  * `content` - The actual input content or prompt text.
  * `status` - The status of the input item (e.g., "completed").
* `has_more` - Boolean indicating if there are more input items available beyond the current page.

## Related Resources

* [openai_model_response Resource](../resources/model_response)
* [openai_model_response Data Source](./model_response)
* [openai_model_responses Data Source](./model_responses) 