---
page_title: "OpenAI: openai_edit Resource"
subcategory: ""
description: |-
  Modifies text using OpenAI's edit models.
---

# openai_edit Resource

The `openai_edit` resource creates edited versions of input text using OpenAI's edit models. This resource is useful for making specific changes to text, such as fixing spelling errors, changing writing styles, or modifying content in other specified ways.

## Example Usage

```hcl
resource "openai_edit" "spelling_correction" {
  model       = "text-davinci-edit-001"
  input       = "The quik brown fox jumps ovr the lazzy dog."
  instruction = "Fix the spelling mistakes."
}

output "corrected_text" {
  value = openai_edit.spelling_correction.choices[0].text
}

# Style modification example
resource "openai_edit" "style_change" {
  model       = "text-davinci-edit-001"
  input       = "I think this food is pretty good."
  instruction = "Make this sound like Shakespeare."
  temperature = 0.7
  n           = 1
}

output "shakespeare_style" {
  value = openai_edit.style_change.choices[0].text
}
```

## Argument Reference

* `model` - (Required) The ID of the model to use for the edit. Currently supported models:
  * `text-davinci-edit-001`
  * `code-davinci-edit-001`
* `input` - (Optional) The text to use as a starting point for the edit. If not provided, will default to an empty string.
* `instruction` - (Required) The instruction that tells the model how to edit the input.
* `temperature` - (Optional) Controls randomness. Range: 0.0 (deterministic) to 2.0 (more random). Defaults to 1.0.
* `top_p` - (Optional) Controls diversity via nucleus sampling. Range: 0.0 to 1.0. Defaults to 1.0.
* `n` - (Optional) How many edits to generate. Defaults to 1.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A unique identifier for this edit resource.
* `choices` - A list of generated edits:
  * `text` - The edited text.
  * `index` - The index of this edit in the list.
* `usage` - Information about token usage:
  * `prompt_tokens` - The number of tokens in the input.
  * `completion_tokens` - The number of tokens in the edited output.
  * `total_tokens` - The total number of tokens used (input + output).
* `created` - The timestamp when the edit was created.
* `object` - The object type, always "edit".

## Import

Edit resources cannot be imported because they represent one-time API calls rather than persistent resources. 