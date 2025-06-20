# openai_model Data Source

Retrieves information about a specific OpenAI model.

## Example Usage

```hcl
data "openai_model" "gpt4o" {
  model_id = "gpt-4o"
}

output "model_details" {
  value = "Model ${data.openai_model.gpt4o.model_id} is owned by ${data.openai_model.gpt4o.owned_by} and was created at ${data.openai_model.gpt4o.created}"
}
```

## Argument Reference

* `model_id` - (Required) The ID of the model to retrieve information for.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the model (same as `model_id`).
* `created` - The Unix timestamp for when the model was created.
* `owned_by` - The organization that owns the model.
* `object` - The object type, which is always "model".

## API Key Configuration

You can provide API keys in the following ways:

1. **Provider-level API key** - Set via the provider block or `OPENAI_API_KEY` environment variable.

If the environment variable approach isn't working, explicitly set the key in your provider configuration:

```hcl
provider "openai" {
  api_key = var.openai_api_key
}
```

Then define the variable and provide it via a tfvars file or command line argument:

```hcl
variable "openai_api_key" {
  description = "OpenAI API Key"
  type        = string
  sensitive   = true
}
``` 