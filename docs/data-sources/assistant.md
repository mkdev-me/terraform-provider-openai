---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "openai_assistant Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  
---

# openai_assistant (Data Source)



## Example Usage

```terraform
# Fetch a specific assistant by ID
data "openai_assistant" "code_reviewer" {
  assistant_id = "asst_8sPATZ7dVbBL1m1Yve97j2BM"
}

# Output the assistant ID
output "assistant_id" {
  value = data.openai_assistant.code_reviewer.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `assistant_id` (String) The ID of the assistant to retrieve

### Read-Only

- `created_at` (Number) The timestamp for when the assistant was created
- `description` (String) The description of the assistant
- `file_ids` (List of String) List of file IDs attached to the assistant
- `id` (String) The ID of this resource.
- `instructions` (String) The system instructions of the assistant
- `metadata` (Map of String) Metadata attached to the assistant
- `model` (String) The model used by the assistant
- `name` (String) The name of the assistant
- `object` (String) The object type, which is always 'assistant'
- `reasoning_effort` (String) Constrains the effort spent on reasoning for reasoning models (low, medium, or high)
- `response_format` (String) The format of responses from the assistant
- `temperature` (Number) What sampling temperature to use for this assistant
- `tools` (List of Object) The tools enabled on the assistant (see [below for nested schema](#nestedatt--tools))
- `top_p` (Number) An alternative to sampling with temperature, called nucleus sampling

<a id="nestedatt--tools"></a>
### Nested Schema for `tools`

Read-Only:

- `function` (List of Object) (see [below for nested schema](#nestedobjatt--tools--function))
- `type` (String)

<a id="nestedobjatt--tools--function"></a>
### Nested Schema for `tools.function`

Read-Only:

- `description` (String)
- `name` (String)
- `parameters` (String)
