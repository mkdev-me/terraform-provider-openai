---
page_title: "OpenAI: openai_model Resource"
subcategory: ""
description: |-
  Manages OpenAI model configurations.
---

# openai_model Resource

The `openai_model` resource allows you to configure how OpenAI models are used within your project. Note that this resource does not create new models, but rather manages configuration associated with existing models.

## Example Usage

```hcl
resource "openai_model" "gpt4" {
  model_id = "gpt-4"
  
  # Custom configuration for this model
  config {
    default_model_version = "gpt-4-0613"
    allowed_model_versions = [
      "gpt-4-0613",
      "gpt-4-1106-preview"
    ]
  }

  # Set your own custom custom context window
  context_window_settings {
    max_input_tokens  = 8000
    max_output_tokens = 4000
  }
}
```

## Argument Reference

* `model_id` - (Required) The ID of the OpenAI model to configure.
* `config` - (Optional) Configuration block for the model. Supports:
  * `default_model_version` - (Optional) The default version of the model to use.
  * `allowed_model_versions` - (Optional) A list of allowed model versions.
* `context_window_settings` - (Optional) Context window settings for the model. Supports:
  * `max_input_tokens` - (Optional) The maximum number of input tokens.
  * `max_output_tokens` - (Optional) The maximum number of output tokens.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the model, which is the same as the `model_id`.
* `created_at` - The timestamp when the model configuration was created.

## Import

Model configurations can be imported using the model ID, e.g.,

```bash
terraform import openai_model.example gpt-4
``` 