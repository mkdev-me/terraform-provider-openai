---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "openai_model_response_input_items Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Data source for retrieving input items for an OpenAI model response
---

# openai_model_response_input_items (Data Source)

Data source for retrieving input items for an OpenAI model response



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `response_id` (String) The ID of the model response to retrieve input items for

### Optional

- `after` (String) An item ID to list items after, used in pagination
- `before` (String) An item ID to list items before, used in pagination
- `include` (List of String) Additional fields to include in the response
- `limit` (Number) A limit on the number of objects to be returned (1-100, default: 20)
- `order` (String) The order to return items in (asc or desc, default: asc)

### Read-Only

- `first_id` (String) The ID of the first item in the list
- `has_more` (Boolean) Whether there are more items to fetch
- `id` (String) The ID of this resource.
- `input_items` (List of Object) The input items for the model response (see [below for nested schema](#nestedatt--input_items))
- `last_id` (String) The ID of the last item in the list

<a id="nestedatt--input_items"></a>
### Nested Schema for `input_items`

Read-Only:

- `content` (String)
- `id` (String)
- `role` (String)
- `status` (String)
- `type` (String)
