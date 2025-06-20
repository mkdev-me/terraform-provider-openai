# Fine-Tuning Examples

This directory contains examples demonstrating how to use the OpenAI Terraform Provider to manage fine-tuning jobs and related resources.

## Resources Demonstrated

- `openai_file`: Uploading training data files
- `openai_fine_tuning_job`: Creating and managing fine-tuning jobs with different methods
  - Basic fine-tuning
  - Supervised fine-tuning with custom hyperparameters
  - Timeout-protected fine-tuning
- `openai_fine_tuning_checkpoint_permission`: Managing permissions for fine-tuning checkpoints

## Data Sources Demonstrated
- `openai_fine_tuning_job`: Get details about a specific fine-tuning job
- `openai_fine_tuning_jobs`: List multiple fine-tuning jobs
- `openai_fine_tuning_checkpoints`: Get checkpoints for a specific job
- `openai_fine_tuning_events`: Get events for a specific job
- `openai_fine_tuning_checkpoint_permissions`: Get permissions for a checkpoint

## Key Features

### File Management
- Upload training data files for fine-tuning

### Fine-Tuning Methods
- Basic fine-tuning with default parameters
- Supervised fine-tuning with custom hyperparameters
  - Control n_epochs, batch_size, learning_rate_multiplier
- Fine-tuning with timeout protection
  - Automatically cancel long-running jobs

### Checkpoint Management
- Import existing checkpoint permissions
- Set permissions for fine-tuning checkpoints

## Prerequisites

Before running these examples, ensure you have:

1. An OpenAI API key with appropriate permissions
2. For checkpoint permissions, an admin API key with the following requirements:
   - API key with admin privileges
   - The `api.fine_tuning.checkpoints.write` scope
   - Owner role in your organization

## Environment Variables

```bash
# Regular API key for most operations
export OPENAI_API_KEY="your-api-key"

# Admin API key for checkpoint permissions
export OPENAI_ADMIN_KEY="your-admin-api-key"
```

## Running the Examples

```bash
# Initialize Terraform
terraform init

# Apply the configuration
terraform apply -var="admin_api_key=$OPENAI_ADMIN_KEY" -var="project_id=your-project-id" -var="checkpoint_id=your-checkpoint-id" -var="use_admin_key=true"
```

## Files in This Directory

- `main.tf`: Main Terraform configuration with fine-tuning resources
- `checkpoints.tf`: Configuration for checkpoint permissions
- `data_source_examples.tf`: Examples of all five fine-tuning data sources
- `data/`: Directory containing training data files
- `README-CHECKPOINT-PERMISSIONS.md`: Additional documentation for checkpoint permissions
- `README-PERMISSIONS.md`: General permissions documentation

## Importing Existing Checkpoint Permissions

If you've already created checkpoint permissions outside of Terraform, you can import them:

```bash
terraform import -var="admin_api_key=$OPENAI_ADMIN_KEY" -var="checkpoint_id=your-checkpoint-id" openai_fine_tuning_checkpoint_permission.checkpoint_permission your-permission-id
```

## Notes

- Fine-tuning jobs can take a considerable amount of time to complete
- Checkpoint permissions require special admin privileges
- The `use_admin_key` variable controls whether admin operations are attempted
- Fine-tuning costs vary based on model and data size

## Example Files

1. **Basic Fine-Tuning (`1_basic_fine_tuning.tf`)**
   - Simple example of creating a fine-tuning job with minimal configuration
   - Demonstrates the basic required fields: model and training_file

2. **Supervised Fine-Tuning (`2_supervised_fine_tuning.tf`)**
   - Shows how to use the supervised training method with custom hyperparameters
   - Includes settings for n_epochs, batch_size, and learning_rate_multiplier

3. **DPO Fine-Tuning (`3_dpo_fine_tuning.tf`)**
   - Demonstrates using Direct Preference Optimization (DPO) method
   - Includes beta parameter configuration

4. **Weights & Biases Integration (`4_wandb_integration.tf`)**
   - Shows how to integrate with Weights & Biases for tracking
   - Requires W&B integration to be enabled in your OpenAI organization settings

5. **Timeout Protection (`5_timeout_example.tf`)**
   - Demonstrates how to set a timeout to automatically cancel long-running jobs
   - Useful for preventing excessive costs or quota consumption

6. **Job Events and Data Sources**

Fine-tuning jobs emit events that can be monitored using data sources:

```hcl
data "openai_fine_tuning_events" "job_events" {
  fine_tuning_job_id = "ftjob-abc123"  # Replace with actual job ID
  limit = 50  # Number of events to retrieve
}

output "event_count" {
  value = length(data.openai_fine_tuning_events.job_events.events)
}

output "latest_event_message" {
  value = length(data.openai_fine_tuning_events.job_events.events) > 0 ?
    data.openai_fine_tuning_events.job_events.events[0].message : ""
}
```

Events provide:
- Training progress updates
- Error messages if issues occur
- Metrics about the training process
- Status changes (e.g., when the job completes)

You can filter events by level (info, warning, error) and sort them by creation time.

For comprehensive examples of all five fine-tuning data sources, see the `data_source_examples.tf` file in this directory, which demonstrates:
- `openai_fine_tuning_job`: Get details about a specific job
- `openai_fine_tuning_jobs`: List multiple jobs
- `openai_fine_tuning_checkpoints`: Get checkpoints for a job
- `openai_fine_tuning_events`: Get events for a job
- `openai_fine_tuning_checkpoint_permissions`: Get permissions for a checkpoint

7. **Checkpoint Permissions (`7_checkpoint_permissions.tf`)**
   - Demonstrates managing permissions for fine-tuning checkpoints
   - Shows how to grant access to one or more projects
   - Uses project_ids as an array rather than a single project_id

## Usage

To use these examples:

1. Replace placeholder values (like file IDs) with actual values from your OpenAI account.
   - For training files, make sure they are uploaded with purpose="fine-tune"
   - Ex: `file-NwvULSFoHTVHqQHbfcjvyZ` (replace with your actual file ID)

2. Configure provider authentication using environment variables:
   ```
   export OPENAI_API_KEY="your-api-key"
   export OPENAI_ORGANIZATION="your-org-id" # Optional
   ```

3. Initialize Terraform:
   ```
   terraform init
   ```

4. Run Terraform:
   ```
   terraform apply
   ```

## Detailed Examples Explanation

### 1. Basic Fine-Tuning

The basic fine-tuning example demonstrates creating a fine-tuning job with minimal configuration:

```hcl
resource "openai_fine_tuning_job" "basic_example" {
  model         = "gpt-4o-2024-08-06"
  training_file = "file-NwvULSFoHTVHqQHbfcjvyZ"  
  suffix        = "my-custom-model-v1"
}
```

Key components:
- `model`: The base model to fine-tune (must be one of the supported models)
- `training_file`: The ID of an uploaded JSONL file containing training data
- `suffix`: A custom identifier that will be appended to the resulting model name

This example also includes outputs to monitor the fine-tuning job:
```hcl
output "fine_tuning_job_id" {
  value = openai_fine_tuning_job.basic_example.id
}

output "fine_tuning_job_status" {
  value = openai_fine_tuning_job.basic_example.status
}

output "fine_tuned_model" {
  value = openai_fine_tuning_job.basic_example.fine_tuned_model
}
```

### 2. Supervised Fine-Tuning

The supervised fine-tuning example shows how to customize the training process with specific hyperparameters:

```hcl
resource "openai_fine_tuning_job" "supervised_example" {
  model         = "gpt-4o-2024-08-06"
  training_file = "file-NwvULSFoHTVHqQHbfcjvyZ"
  
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

Key hyperparameters:
- `n_epochs`: The number of training epochs (passes through the data)
- `batch_size`: The number of examples processed in each training batch
- `learning_rate_multiplier`: Controls how quickly the model adapts to the training data

Adjusting these parameters can improve model performance or reduce training time, but requires understanding of machine learning concepts.

### 3. DPO Fine-Tuning

Direct Preference Optimization (DPO) is an alternative training method that uses preference data:

```hcl
resource "openai_fine_tuning_job" "dpo_example" {
  model         = "gpt-4o-2024-08-06"
  training_file = "file-NwvULSFoHTVHqQHbfcjvyZ"
  
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

DPO uses:
- A training file with preference pairs (chosen and rejected outputs)
- The `beta` parameter controls the trade-off between following preferences and staying close to the base model behavior
- Higher beta values push the model to more strongly prefer the chosen outputs

### 4. W&B Integration

Weights & Biases integration allows you to track training metrics and experiments:

```hcl
resource "openai_fine_tuning_job" "wandb_example" {
  model         = "gpt-4o-2024-08-06"
  training_file = "file-NwvULSFoHTVHqQHbfcjvyZ"
  
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

W&B integration requires:
- The W&B integration enabled in your OpenAI organization settings
- A W&B project where training metrics will be sent
- Optional name and tags to organize experiments in the W&B dashboard

### 5. Timeout Protection

Long-running fine-tuning jobs can be automatically cancelled after a timeout:

```hcl
resource "openai_fine_tuning_job" "timeout_example" {
  model               = "gpt-4o-2024-08-06"
  training_file       = "file-NwvULSFoHTVHqQHbfcjvyZ"
  
  cancel_after_timeout = 3600  # Cancel after 1 hour
  suffix = "timeout-protected-v1"
}
```

This is useful for:
- Preventing unexpected costs if a job takes too long
- Setting up fallback mechanisms for failed training runs
- Setting deadlines for experimentation

The `cancel_after_timeout` value is specified in seconds. In this example, 3600 seconds = 1 hour.

### 6. Job Events and Data Sources

Fine-tuning jobs emit events that can be monitored using data sources:

```hcl
data "openai_fine_tuning_events" "job_events" {
  fine_tuning_job_id = "ftjob-abc123"  # Replace with actual job ID
  limit = 50  # Number of events to retrieve
}

output "event_count" {
  value = length(data.openai_fine_tuning_events.job_events.events)
}

output "latest_event_message" {
  value = length(data.openai_fine_tuning_events.job_events.events) > 0 ?
    data.openai_fine_tuning_events.job_events.events[0].message : ""
}
```

Events provide:
- Training progress updates
- Error messages if issues occur
- Metrics about the training process
- Status changes (e.g., when the job completes)

You can filter events by level (info, warning, error) and sort them by creation time.

For comprehensive examples of all five fine-tuning data sources, see the `data_source_examples.tf` file in this directory, which demonstrates:
- `openai_fine_tuning_job`: Get details about a specific job
- `openai_fine_tuning_jobs`: List multiple jobs
- `openai_fine_tuning_checkpoints`: Get checkpoints for a job
- `openai_fine_tuning_events`: Get events for a job
- `openai_fine_tuning_checkpoint_permissions`: Get permissions for a checkpoint

### 7. Checkpoint Permissions

Fine-tuning checkpoints can be shared across projects using permissions:

```hcl
resource "openai_fine_tuning_checkpoint_permission" "share_example" {
  checkpoint_id = "checkpoint-abc123"  # Replace with actual checkpoint ID
  
  # Important: project_ids must be an array
  project_ids = ["proj_def456", "proj_ghi789"]
  
  # Permission settings
  allow_view   = true  # Allow viewing the checkpoint
  allow_create = true  # Allow creating models from the checkpoint
}
```

Key points about checkpoint permissions:
- Checkpoints are only available after a fine-tuning job completes
- The `project_ids` parameter must be an array, even for a single project
- You can set view and create permissions separately
- Permissions can be managed and revoked as needed

Example of accessing a checkpoint from another project:
```hcl
data "openai_fine_tuning_checkpoints" "job_checkpoints" {
  fine_tuning_job_id = "ftjob-abc123"  # Replace with actual job ID
}

resource "openai_fine_tuning_checkpoint_permission" "share_with_project" {
  checkpoint_id = data.openai_fine_tuning_checkpoints.job_checkpoints.checkpoints[0].id
  project_ids   = ["proj_xyz789"]  # The project to share with
}
```

## Creating a Complete Fine-Tuning Workflow

A typical fine-tuning workflow combines multiple resources:

1. **Upload training data file**:
```hcl
resource "openai_file" "training_data" {
  file    = "path/to/training_data.jsonl"
  purpose = "fine-tune"
}
```

2. **Create fine-tuning job**:
```hcl
resource "openai_fine_tuning_job" "custom_model" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_data.id
  suffix        = "customer-support-v1"
}
```

3. **Monitor job events**:
```hcl
data "openai_fine_tuning_events" "job_events" {
  fine_tuning_job_id = openai_fine_tuning_job.custom_model.id
}
```

4. **Share checkpoints with other projects**:
```hcl
data "openai_fine_tuning_checkpoints" "job_checkpoints" {
  fine_tuning_job_id = openai_fine_tuning_job.custom_model.id
}

resource "openai_fine_tuning_checkpoint_permission" "share_checkpoint" {
  count = length(data.openai_fine_tuning_checkpoints.job_checkpoints.checkpoints) > 0 ? 1 : 0
  
  checkpoint_id = data.openai_fine_tuning_checkpoints.job_checkpoints.checkpoints[0].id
  project_ids   = ["proj_abc123"]
}
```

This workflow allows you to automate the entire fine-tuning process with Terraform, from data upload to model deployment.

## Supported Models

Fine-tuning is currently available for the following models:

- gpt-4o-2024-08-06
- gpt-4o-mini-2024-07-18
- gpt-4-0613
- gpt-3.5-turbo-0125
- gpt-3.5-turbo-1106
- gpt-3.5-turbo-0613

We recommend using gpt-4o-mini as the best balance of performance, cost, and ease of use for most users.

## Important Notes

- Fine-tuning can incur costs on your OpenAI account
- Jobs may take a long time to complete, especially for larger datasets
- Checkpoint permissions allow sharing fine-tuned models across projects
- Always monitor fine-tuning events to catch any issues early

## Troubleshooting

### 404 Not Found Errors
If you encounter 404 errors when working with fine-tuning resources, check:

1. Your OpenAI API key has access to fine-tuning features
2. You're using a supported model (see list above)
3. The training file exists and was uploaded with purpose="fine-tune"
4. The file format is properly formatted JSONL according to OpenAI's requirements

### Using Weights & Biases Integration
To use the W&B integration example:

1. Enable W&B integration in your OpenAI organization settings:
   - Visit https://platform.openai.com/account/organization
   - Under "Integrations", add your W&B API key

2. Uncomment the code in `4_wandb_integration.tf`

### Checkpoint Permissions
When managing checkpoint permissions:

1. Use `project_ids` (plural) and pass it as an array: `project_ids = ["proj_abc123"]`
2. Note that checkpoint permissions are only available after a fine-tuning job completes
3. You must have appropriate permissions to access the checkpoints

## Checkpoint Permissions Management

Managing checkpoint permissions requires special privileges and specific procedures. The `openai_fine_tuning_checkpoint_permission` resource allows you to share fine-tuning checkpoints between projects.

### Key Requirements:

1. **Admin API Key Required**: All checkpoint permission operations require an admin API key with the `api.fine_tuning.checkpoints.write` scope
2. **Owner Role Required**: Your user must have the Owner role in the organization
3. **Different Target Project**: You cannot grant permission to the project that already owns the checkpoint

### Creating Checkpoint Permissions

```bash
terraform apply \
  -var="admin_api_key=$OPENAI_ADMIN_KEY" \
  -var="checkpoint_id=ft:gpt-4o-mini-2024-07-18:org-xyz::ABcDeFgH" \
  -var="project_id=proj_AbCdEfGhIj123456" \
  -var="use_admin_key=true"
```

### Deleting Checkpoint Permissions

**Important**: When removing checkpoint permissions, you MUST use the `-target` flag:

```bash
terraform destroy \
  -target='openai_fine_tuning_checkpoint_permission.checkpoint_permission[0]' \
  -var="admin_api_key=$OPENAI_ADMIN_KEY" \
  -var="checkpoint_id=ft:gpt-4o-mini-2024-07-18:org-xyz::ABcDeFgH" \
  -var="project_id=proj_AbCdEfGhIj123456" \
  -var="use_admin_key=true"
```

For complete details on checkpoint permissions, see [README-CHECKPOINT-PERMISSIONS.md](./README-CHECKPOINT-PERMISSIONS.md)

## State Management and Importing Resources

When working with OpenAI resources that may already exist outside of Terraform, you can import them into Terraform state:

### Importing Fine-Tuning Jobs

```bash
terraform import 'openai_fine_tuning_job.basic_example' ftjob-AbCdEfGhIjKlMnOp123456
```

### Importing Checkpoint Permissions

```bash
terraform import \
  -var="admin_api_key=$OPENAI_ADMIN_KEY" \
  -var="checkpoint_id=ft:gpt-4o-mini-2024-07-18:org-xyz::AbCdEfGh" \
  -var="project_id=proj_AbCdEfGhIjKlMnOp" \
  -var="use_admin_key=true" \
  'openai_fine_tuning_checkpoint_permission.checkpoint_permission[0]' cp_AbCdEfGhIjKlMnOp123456
```

For imported resources, always add a `lifecycle { ignore_changes = all }` block to prevent Terraform from attempting to modify them.

For detailed information on state management, see [README-CHECKPOINT-PERMISSIONS.md](./README-CHECKPOINT-PERMISSIONS.md#terraform-state-management). 