# openai_fine_tuned_model

This resource allows you to create and manage fine-tuning jobs for OpenAI models. Fine-tuning lets you customize OpenAI's models with your own data to create specialized models that better fit your specific use cases.

## Example Usage

```hcl
# Simple example
resource "openai_fine_tuned_model" "simple" {
  model         = "gpt-3.5-turbo"
  training_file = "file-abc123"
}

# Advanced example with all parameters
resource "openai_fine_tuned_model" "advanced" {
  model           = "gpt-3.5-turbo"
  training_file   = "file-abc123"
  validation_file = "file-xyz789"
  suffix          = "custom-model-name"
  
  hyperparameters {
    n_epochs                = "4"
    batch_size              = 8
    learning_rate_multiplier = 0.8
  }
  
  # Wait for up to 30 minutes during creation
  completion_window = 1800
}
```

## Using the Fine-Tuning Module

For a more streamlined approach, you can use the fine_tuning module:

```hcl
module "fine_tuned_model" {
  source        = "../../modules/fine_tuning"
  
  model         = "gpt-3.5-turbo"
  training_file = "file-abc123"
  
  hyperparameters = {
    n_epochs = "auto"
  }
}

output "model_id" {
  value = module.fine_tuned_model.fine_tuned_model_id
}
```

## Argument Reference

* `model` - (Required) The name of the base model to fine-tune. Currently supported model is "gpt-3.5-turbo". Check OpenAI's documentation for the latest list of supported models.
* `training_file` - (Required) The ID of the file containing training data. This file should be uploaded with purpose="fine-tune".
* `validation_file` - (Optional) The ID of the file containing validation data. This file should be uploaded with purpose="fine-tune".
* `hyperparameters` - (Optional) Configuration block for the hyperparameters to use for the fine-tuning job.
  * `n_epochs` - (Optional) Number of epochs to train for. This can be an integer as a string (e.g., "4") or the string "auto". Defaults to "4".
  * `batch_size` - (Optional) Batch size to use for training.
  * `learning_rate_multiplier` - (Optional) Learning rate multiplier to use for training. Value should be between 0.01 and 10.0.
* `suffix` - (Optional) A string of up to 64 characters that will be added to your fine-tuned model name.
* `completion_window` - (Optional) Time in seconds to wait for job to complete during creation. Default is 0, which means don't wait.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the fine-tuning job.
* `created_at` - The Unix timestamp for when the fine-tuning job was created.
* `finished_at` - The Unix timestamp for when the fine-tuning job completed.
* `fine_tuned_model` - The name of the fine-tuned model, once the job is complete.
* `status` - The current status of the fine-tuning job. Possible values include: "pending", "running", "succeeded", "failed", "cancelled".

## Data Preparation

Your training data should be in JSONL format with each line containing a prompt-completion pair according to the model type:

For chat models (gpt-3.5-turbo):
```json
{"messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "Hello!"}, {"role": "assistant", "content": "Hi there! How can I help you today?"}]}
```

For completion models (davinci-002):
```json
{"prompt": "<prompt text>", "completion": "<ideal completion>"}
```

## Timeouts and Limitations

- The resource creation process can time out if the fine-tuning job takes longer than the `completion_window`. Consider using a longer timeout or setting `completion_window = 0` for very large datasets.
- This resource is immutable. Any changes will result in the destruction and recreation of the resource.
- Fine-tuned models can incur both training and usage costs. See OpenAI's pricing page for details.

## Importing

Fine-tuned models can be imported using their job ID:

```
terraform import openai_fine_tuned_model.example ft-abc123
```

## Adding Checkpoint Permissions

To share a fine-tuned model with other projects in your organization, you'll need to use the OpenAI API directly:

```bash
curl https://api.openai.com/v1/fine_tuning/checkpoints/{model_id}/permissions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{"project_ids": ["proj_abGMw1llN8IrBb6SvvY5A1iH"]}'
```

Replace `{model_id}` with your fine-tuned model ID and provide the project IDs you wish to grant access to.

## Notes

- Fine-tuning jobs can take from minutes to hours to complete depending on the size of your dataset and model.
- Currently, gpt-3.5-turbo is supported for fine-tuning. Always check OpenAI documentation for the latest list of supported models.
- See the [OpenAI fine-tuning documentation](https://platform.openai.com/docs/api-reference/fine-tuning) for more details on the fine-tuning process. 