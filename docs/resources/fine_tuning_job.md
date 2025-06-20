# openai_fine_tuning_job

This resource allows you to create and manage fine-tuning jobs for OpenAI models. Fine-tuning enables you to customize OpenAI's models with your own data, creating specialized models that better fit your specific use cases.

## Example Usage

### Basic Example
```hcl
resource "openai_fine_tuning_job" "basic" {
  model         = "gpt-4o-mini"
  training_file = "file-abc123"
  suffix        = "my-custom-model-v1"
}
```

### Supervised Fine-Tuning with Custom Hyperparameters
```hcl
resource "openai_fine_tuning_job" "supervised" {
  model         = "gpt-4o-mini"
  training_file = "file-abc123"
  
  method {
    type = "supervised"
    supervised {
      hyperparameters {
        n_epochs = 3
        batch_size = 8
        learning_rate_multiplier = 0.1
      }
    }
  }
}
```

### DPO Fine-Tuning
```hcl
resource "openai_fine_tuning_job" "dpo" {
  model          = "gpt-4o-mini"
  training_file  = "file-abc123"
  validation_file = "file-xyz789"
  
  method {
    type = "dpo"
    dpo {
      hyperparameters {
        beta = 0.1
      }
    }
  }
}
```

### With Integrations (Weights & Biases)
```hcl
resource "openai_fine_tuning_job" "with_wandb" {
  model         = "gpt-4o-mini"
  training_file = "file-abc123"
  
  integrations {
    type = "wandb"
    wandb {
      project = "my-wandb-project"
      name    = "ft-run-display-name"
      tags    = ["first-experiment", "v2"]
    }
  }
}
```

### With Automatic Cancellation After Timeout
```hcl
resource "openai_fine_tuning_job" "with_timeout" {
  model               = "gpt-4o-mini"
  training_file       = "file-abc123"
  cancel_after_timeout = 3600  # Cancel after 1 hour
}
```

## Argument Reference

* `model` - (Required) The name of the base model to fine-tune (e.g., "gpt-4o-mini").
* `training_file` - (Required) The ID of the file containing training data, which must have been uploaded with purpose="fine-tune".
* `validation_file` - (Optional) The ID of the file containing validation data.
* `method` - (Optional) The method used for fine-tuning.
  * `type` - (Required) The type of fine-tuning method (e.g., "supervised" or "dpo").
  * `supervised` - (Optional) Configuration for supervised fine-tuning.
    * `hyperparameters` - (Optional) Hyperparameters for supervised fine-tuning.
      * `n_epochs` - (Optional) Number of epochs to train for.
      * `batch_size` - (Optional) Number of examples in each batch.
      * `learning_rate_multiplier` - (Optional) Learning rate multiplier.
  * `dpo` - (Optional) Configuration for DPO fine-tuning.
    * `hyperparameters` - (Optional) Hyperparameters for DPO fine-tuning.
      * `beta` - (Optional) Beta parameter for DPO.
* `hyperparameters` - (Deprecated) Use `method` instead. Legacy hyperparameters structure.
  * `n_epochs` - (Optional) Number of epochs to train for.
  * `batch_size` - (Optional) Number of examples in each batch.
  * `learning_rate_multiplier` - (Optional) Learning rate multiplier.
* `integrations` - (Optional) List of integrations to enable for the fine-tuning job.
  * `type` - (Required) The type of integration (e.g., "wandb").
  * `wandb` - (Optional) Configuration for Weights & Biases integration.
    * `project` - (Required) The W&B project name.
    * `name` - (Optional) The W&B run name.
    * `tags` - (Optional) Tags for the W&B run.
* `metadata` - (Optional) Key-value pairs attached to the fine-tuning job (up to 16 pairs).
* `seed` - (Optional) Seed for reproducibility.
* `suffix` - (Optional) A string of up to 64 characters that will be added to your fine-tuned model name.
* `cancel_after_timeout` - (Optional) Automatically cancel the job after this many seconds.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the fine-tuning job.
* `status` - The status of the fine-tuning job.
* `fine_tuned_model` - The name of the fine-tuned model.
* `organization_id` - The organization ID the fine-tuning job belongs to.
* `result_files` - Result files from the fine-tuning job.
* `validation_loss` - The validation loss for the fine-tuning job.
* `trained_tokens` - The number of tokens trained during the fine-tuning job.
* `created_at` - Timestamp when the fine-tuning job was created.
* `finished_at` - Timestamp when the fine-tuning job was completed.
* `last_updated` - Timestamp of when this resource was last updated.

## Import

Fine-tuning jobs can be imported using their ID:

```
terraform import openai_fine_tuning_job.example ft-abc123
```

## Data Preparation

Your training data should be in JSONL format with each line containing a prompt-completion pair according to the model type:

For chat models:
```json
{"messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "Hello!"}, {"role": "assistant", "content": "Hi there! How can I help you today?"}]}
```

For DPO fine-tuning, each example should include a preferred response and a rejected response:
```json
{"messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "Hello!"}], "chosen": "Hi there! How can I help you today?", "rejected": "Yo, what's up?"}
```

## Notes and Limitations

- Fine-tuning jobs can take from minutes to hours depending on the model and dataset size.
- For more details, see the [OpenAI fine-tuning documentation](https://platform.openai.com/docs/api-reference/fine-tuning). 