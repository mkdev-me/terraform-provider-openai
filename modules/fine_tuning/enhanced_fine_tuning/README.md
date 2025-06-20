# Enhanced Fine-Tuning Module

This Terraform module provides enhanced functionality for OpenAI fine-tuning operations, including:

1. Creating fine-tuning jobs
2. Monitoring job status and events
3. Accessing checkpoints
4. Sharing checkpoints with other organizations

## Usage

```hcl
module "enhanced_fine_tuning" {
  source = "../../modules/fine_tuning/enhanced_fine_tuning"

  # Required parameters
  model         = "gpt-3.5-turbo"
  training_file = "file-abc123"
  
  # Optional parameters
  validation_file   = "file-def456"
  suffix            = "my-custom-model"
  hyperparameters   = {
    n_epochs = "3"
  }
  
  # Monitoring and checkpoint access
  enable_monitoring      = true
  enable_checkpoint_access = true
  
  # Organization IDs to share checkpoints with
  share_with_organizations = [
    "org-partner1",
    "org-partner2"
  ]
}

# Output the fine-tuned model ID
output "fine_tuned_model" {
  value = module.enhanced_fine_tuning.fine_tuned_model_id
}

# Output checkpoints if available
output "checkpoints" {
  value = module.enhanced_fine_tuning.checkpoints
}
```

## Features

### Fine-Tuning Job Creation

The module creates a fine-tuning job using the specified model and training data. You can customize the job with validation data, a custom suffix, and hyperparameters.

### Monitoring

When `enable_monitoring` is set to `true` (default), the module will expose fine-tuning job events through the `events` output. This allows you to track the progress of your fine-tuning job.

### Checkpoint Access

When `enable_checkpoint_access` is set to `true`, the module will attempt to access checkpoints for the fine-tuning job once it has completed successfully. Checkpoints will be exposed through the `checkpoints` output.

### Checkpoint Sharing

You can share checkpoints with other organizations by specifying their organization IDs in the `share_with_organizations` list. The module will automatically create checkpoint permissions for each organization, but only after the fine-tuning job has completed successfully.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| model | The name of the model to fine-tune | `string` | n/a | yes |
| training_file | The ID of an uploaded file that contains training data | `string` | n/a | yes |
| validation_file | The ID of an uploaded file that contains validation data | `string` | `null` | no |
| hyperparameters | Hyperparameters for the fine-tuning job | `map(any)` | `null` | no |
| suffix | A string of up to 64 characters that will be added to your fine-tuned model name | `string` | `null` | no |
| enable_monitoring | Whether to enable monitoring of the fine-tuning job events | `bool` | `true` | no |
| enable_checkpoint_access | Whether to enable access to the fine-tuning job checkpoints | `bool` | `false` | no |
| share_with_organizations | List of organization IDs to share checkpoints with | `list(string)` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| fine_tuned_model_id | The ID of the fine-tuned model created |
| fine_tuning_job_id | The ID of the fine-tuning job |
| status | The current status of the fine-tuning job |
| created_at | The timestamp when the fine-tuning job was created |
| finished_at | The timestamp when the fine-tuning job was completed |
| events | Events from the fine-tuning job (if monitoring is enabled) |
| checkpoints | Checkpoints from the fine-tuning job (if checkpoint access is enabled and job is finished) |
| checkpoint_permissions | Permissions created for sharing checkpoints with other organizations | 