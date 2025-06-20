# Fine-Tuning Examples

This directory contains comprehensive examples for managing OpenAI fine-tuning jobs and related resources using Terraform.

## Overview

Fine-tuning allows you to customize OpenAI models with your own training data. This example demonstrates:

- Uploading training data files
- Creating fine-tuning jobs with various configurations
- Managing checkpoint permissions
- Monitoring job progress and events

## Prerequisites

- OpenAI API key with fine-tuning permissions
- Admin API key (for checkpoint permissions)
- Training data in JSONL format

## Quick Start

```bash
# Set up authentication
export OPENAI_API_KEY="sk-proj-..."
export OPENAI_ADMIN_KEY="sk-..."  # For checkpoint permissions

# Initialize and apply
terraform init
terraform apply
```

## Examples Included

### 1. Basic Fine-Tuning (`1_basic_fine_tuning.tf`)

Simple fine-tuning job with minimal configuration:

```hcl
resource "openai_fine_tuning_job" "basic_example" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = "file-abc123"  # Your uploaded file ID
  suffix        = "custom-v1"
}
```

### 2. Supervised Fine-Tuning (`2_supervised_fine_tuning.tf`)

Fine-tuning with custom hyperparameters:

```hcl
resource "openai_fine_tuning_job" "supervised_example" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = "file-abc123"
  
  method {
    type = "supervised"
    supervised {
      hyperparameters {
        n_epochs                 = 3
        batch_size              = 8
        learning_rate_multiplier = 0.1
      }
    }
  }
}
```

### 3. DPO Fine-Tuning (`3_dpo_fine_tuning.tf`)

Direct Preference Optimization for preference-based training:

```hcl
resource "openai_fine_tuning_job" "dpo_example" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = "file-abc123"
  
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

### 4. W&B Integration (`4_wandb_integration.tf`)

Track training metrics with Weights & Biases:

```hcl
resource "openai_fine_tuning_job" "wandb_example" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = "file-abc123"
  
  integrations {
    type = "wandb"
    wandb {
      project = "fine-tuning-experiments"
      name    = "run-v1"
      tags    = ["production", "v1"]
    }
  }
}
```

### 5. Timeout Protection (`5_timeout_example.tf`)

Automatically cancel long-running jobs:

```hcl
resource "openai_fine_tuning_job" "timeout_example" {
  model                = "gpt-4o-mini-2024-07-18"
  training_file        = "file-abc123"
  cancel_after_timeout = 3600  # 1 hour in seconds
}
```

### 6. Checkpoint Permissions (`7_checkpoint_permissions.tf`)

Share checkpoints between projects:

```hcl
resource "openai_fine_tuning_checkpoint_permission" "share_example" {
  checkpoint_id = "ft:gpt-4o-mini:org-xyz::abc123"
  project_ids   = ["proj_def456", "proj_ghi789"]
  allow_view    = true
  allow_create  = true
}
```

## Data Sources

Monitor and query fine-tuning resources (`data_source_examples.tf`):

```hcl
# Get job details
data "openai_fine_tuning_job" "job_info" {
  fine_tuning_job_id = "ftjob-abc123"
}

# List all jobs
data "openai_fine_tuning_jobs" "all_jobs" {
  limit = 20
}

# Get job events
data "openai_fine_tuning_events" "job_events" {
  fine_tuning_job_id = "ftjob-abc123"
  limit              = 50
}

# Get checkpoints
data "openai_fine_tuning_checkpoints" "job_checkpoints" {
  fine_tuning_job_id = "ftjob-abc123"
}
```

## Complete Workflow Example

```hcl
# 1. Upload training data
resource "openai_file" "training_data" {
  filename = "training.jsonl"
  purpose  = "fine-tune"
  content  = file("${path.module}/data/training.jsonl")
}

# 2. Create fine-tuning job
resource "openai_fine_tuning_job" "model" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_data.id
  suffix        = "custom-model-v1"
}

# 3. Monitor events
data "openai_fine_tuning_events" "progress" {
  fine_tuning_job_id = openai_fine_tuning_job.model.id
}

# 4. Output results
output "model_id" {
  value = openai_fine_tuning_job.model.fine_tuned_model
}

output "status" {
  value = openai_fine_tuning_job.model.status
}
```

## Supported Models

Fine-tuning is available for:

- `gpt-4o-2024-08-06`
- `gpt-4o-mini-2024-07-18` (recommended)
- `gpt-4-0613`
- `gpt-3.5-turbo-0125`
- `gpt-3.5-turbo-1106`
- `gpt-3.5-turbo-0613`

## Training Data Format

Training data must be in JSONL format with conversation examples:

```json
{"messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "Hello"}, {"role": "assistant", "content": "Hi! How can I help you?"}]}
{"messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "What's the weather?"}, {"role": "assistant", "content": "I don't have access to real-time weather data."}]}
```

## Important Notes

### Checkpoint Permissions

- Requires admin API key with `api.fine_tuning.checkpoints.write` scope
- User must have Owner role in the organization
- Cannot grant permissions to the project that owns the checkpoint
- Use `-target` flag when destroying permissions

### Cost Considerations

- Fine-tuning incurs costs based on tokens processed
- Monitor job progress to avoid unexpected charges
- Use timeout protection for experimentation

### Troubleshooting

Common issues and solutions:

1. **404 Not Found**: Verify API key has fine-tuning access
2. **Invalid file**: Ensure JSONL format is correct
3. **Permission denied**: Check admin key for checkpoint operations
4. **W&B integration**: Enable in OpenAI organization settings first

## Additional Resources

- [Checkpoint Permissions Guide](./README-CHECKPOINT-PERMISSIONS.md)
- [OpenAI Fine-Tuning Guide](https://platform.openai.com/docs/guides/fine-tuning)
- [JSONL Format Requirements](https://platform.openai.com/docs/guides/fine-tuning/preparing-your-dataset)