# OpenAI Fine-Tuning Terraform Module

This module simplifies the management of OpenAI fine-tuning jobs, allowing you to easily create and monitor fine-tuned models through Terraform.

## Features

- Create and manage fine-tuning jobs
- Support for multiple fine-tuning methods (supervised, DPO)
- Custom hyperparameter configuration
- Automatic timeout protection for long-running jobs
- Support for checkpoint permissions management
- Integration with the OpenAI Admin API for checkpoint operations

## Example Usage

### Basic Fine-Tuning

```hcl
module "fine_tuning" {
  source = "../../modules/fine_tuning"

  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_data.id
  suffix        = "my-custom-model"
}

output "fine_tuned_model" {
  value = module.fine_tuning.fine_tuned_model
}
```

### Supervised Fine-Tuning with Custom Hyperparameters

```hcl
module "supervised_fine_tuning" {
  source = "../../modules/fine_tuning"

  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_data.id
  
  training_method = "supervised"
  
  hyperparameters = {
    n_epochs                = 3
    batch_size              = 8
    learning_rate_multiplier = 0.1
  }
}
```

### Timeout Protection

```hcl
module "timeout_protected_fine_tuning" {
  source = "../../modules/fine_tuning"

  model                = "gpt-4o-mini-2024-07-18"
  training_file        = openai_file.training_data.id
  suffix               = "timeout-protected"
  cancel_after_timeout = 3600  # Cancel after 1 hour (in seconds)
}
```

### Managing Checkpoint Permissions

To manage checkpoint permissions, use the enhanced_fine_tuning submodule, which supports admin operations:

```hcl
module "enhanced_fine_tuning" {
  source = "../../modules/fine_tuning/enhanced_fine_tuning"

  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_data.id
  
  # Checkpoint permissions (requires admin API key)
  checkpoint_permissions = {
    enabled      = true
    project_ids  = ["proj_abc123", "proj_def456"]
  }
}
```

## Admin API Key Support

For checkpoint permissions operations, an admin API key with the appropriate scopes is required. The module now automatically reads the admin API key from the `OPENAI_ADMIN_KEY` environment variable, eliminating the need to pass it explicitly in your configuration.

```bash
# Set the admin API key as an environment variable
export OPENAI_ADMIN_KEY="your-admin-api-key"

# Run Terraform commands
terraform apply
```

This approach enhances security by avoiding placing sensitive keys in your Terraform code or state files.

## Input Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| model | The base model to fine-tune | string | n/a | yes |
| training_file | The ID of the training data file | string | n/a | yes |
| validation_file | The ID of an optional validation file | string | null | no |
| suffix | A suffix to add to the name of the fine-tuned model | string | null | no |
| training_method | The training method to use ("supervised" or "dpo") | string | "supervised" | no |
| hyperparameters | Custom hyperparameters for the fine-tuning job | map(string) | {} | no |
| cancel_after_timeout | Cancel the job after this many seconds | number | null | no |
| metadata | Additional metadata for the fine-tuning job | map(string) | {} | no |

## Output Values

| Name | Description |
|------|-------------|
| id | The ID of the fine-tuning job |
| status | The current status of the fine-tuning job |
| fine_tuned_model | The ID of the resulting fine-tuned model |
| created_at | The timestamp when the job was created |
| finished_at | The timestamp when the job was completed |
| trained_tokens | The number of tokens processed during training |

## Submodules

### Enhanced Fine-Tuning

The `enhanced_fine_tuning` submodule extends the basic module with additional capabilities:

- Checkpoint permission management
- Integration with the OpenAI Admin API
- Project-level access control for fine-tuned models

See the [Enhanced Fine-Tuning README](./enhanced_fine_tuning/README.md) for more details.

## Notes on Checkpoint Permissions

Checkpoint permissions require specific authentication and permissions:

1. The API key must have the **Owner** role in your organization.
2. The API key needs the `api.fine_tuning.checkpoints.read` scope for reading permissions and `api.fine_tuning.checkpoints.write` for creating/updating permissions.
3. The module will use the `OPENAI_ADMIN_KEY` environment variable for admin operations rather than requiring it to be passed as a parameter.

If you encounter a 401 Unauthorized error with a message about missing scopes:
- Verify your admin API key has the correct scopes
- Ensure the environment variable is properly set
- Check that the user associated with the API key has Owner privileges in the OpenAI organization

## Common Errors and Troubleshooting

| Error | Description | Solution |
|-------|-------------|----------|
| 401 Unauthorized | Insufficient permissions for checkpoint operations | Set OPENAI_ADMIN_KEY to an admin key with appropriate scopes |
| 404 Not Found | Resource (file, checkpoint) doesn't exist | Verify IDs and ensure resources exist in your account |
| Timeout during apply | Fine-tuning job is still running | Use `cancel_after_timeout` or increase Terraform timeout |
| Invalid model | The specified model doesn't support fine-tuning | Use a supported model (e.g., gpt-4o-mini-2024-07-18, gpt-3.5-turbo) |

For more detailed troubleshooting, refer to the [OpenAI Provider Troubleshooting Guide](../../docs/TROUBLESHOOTING.md).

## Related Resources

- [OpenAI Fine-Tuning API Documentation](https://platform.openai.com/docs/api-reference/fine-tuning)
- [Fine-Tuning Best Practices](https://platform.openai.com/docs/guides/fine-tuning) 