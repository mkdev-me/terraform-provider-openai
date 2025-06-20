# OpenAI Fine-Tuning Example
# This file demonstrates how to use Terraform to manage OpenAI fine-tuning resources

# ----------- PROVIDER CONFIGURATION -----------
terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

# Main provider using the regular API key from environment
provider "openai" {
  # API key and organization ID can be set using environment variables:
  # OPENAI_API_KEY and OPENAI_ORGANIZATION_ID
  # 
  # For checkpoint permissions, you need an API key with admin privileges and
  # the api.fine_tuning.checkpoints.write scope. You need the Owner role in your organization.
}

# ----------- FILE RESOURCES -----------
# Upload the fine-tuning data file to OpenAI
resource "openai_file" "training_file" {
  file    = "./data/fine_tune_data_v4.jsonl"
  purpose = "fine-tune"
}

# ----------- FINE-TUNING JOBS -----------
# Basic example of a fine-tuning job
resource "openai_fine_tuning_job" "basic_example" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_file.id
  suffix        = "my-custom-model-v1"

  # Ensure file is created before the fine-tuning job
  depends_on = [openai_file.training_file]

  # Prevent modifications to imported resources
  lifecycle {
    ignore_changes = all
  }
}

# Example with supervised method and custom hyperparameters
resource "openai_fine_tuning_job" "supervised_example" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_file.id

  # Using the new method format for hyperparameters
  method {
    type = "supervised"
    supervised {
      hyperparameters {
        n_epochs                 = 3
        batch_size               = 8
        learning_rate_multiplier = 0.1
      }
    }
  }

  # Ensure file is created before the fine-tuning job
  depends_on = [openai_file.training_file]

  # Prevent modifications to imported resources
  lifecycle {
    ignore_changes = all
  }
}

# Example with DPO method for fine-tuning
#resource "openai_fine_tuning_job" "dpo_example" {
#  model          = "gpt-4o-2024-08-06"
#  training_file  = openai_file.training_file.id
#  
#  # Using the DPO method
#  method {
#    type = "dpo"
#    dpo {
#      hyperparameters {
#        beta = 0.1
#      }
#    }
#  }
#  
#  # Ensure file is created before the fine-tuning job
#  depends_on = [openai_file.training_file]
#}

# Example with timeout to auto-cancel long-running jobs
resource "openai_fine_tuning_job" "timeout_example" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_file.id
  suffix        = "timeout-protected-v1"

  # Cancel after 1 hour if still running
  cancel_after_timeout = 3600 # 1 hour in seconds

  # Ensure file is created before the fine-tuning job
  depends_on = [openai_file.training_file]

  # Prevent modifications to imported resources
  lifecycle {
    ignore_changes = all
  }
}

# Example with Weights & Biases integration - DISABLED
# Note: To use this example, uncomment and provide your W&B API key
/*
resource "openai_fine_tuning_job" "wandb_example" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_file.id
  
  # W&B integration
  integrations {
    type = "wandb"
    wandb {
      project = "my-openai-project"
      entity = "my-wandb-entity" # Optional
      api_key = "wand_xyz123"    # Your Weights & Biases API key
    }
  }
  
  # Ensure file is created before the fine-tuning job
  depends_on = [openai_file.training_file]
}
*/

# ----------- OUTPUTS -----------
# File outputs
output "training_file_id" {
  value       = openai_file.training_file.id
  description = "The ID of the uploaded training file"
}

# Basic example job outputs
output "fine_tuning_job_id" {
  value = openai_fine_tuning_job.basic_example.id
}

output "fine_tuning_job_status" {
  value = openai_fine_tuning_job.basic_example.status
}

output "fine_tuned_model" {
  value = openai_fine_tuning_job.basic_example.fine_tuned_model
}

# Supervised example job outputs
output "supervised_job_id" {
  value = openai_fine_tuning_job.supervised_example.id
}

output "supervised_job_status" {
  value = openai_fine_tuning_job.supervised_example.status
}

output "supervised_fine_tuned_model" {
  value = openai_fine_tuning_job.supervised_example.fine_tuned_model
}

# DPO example job outputs
#output "dpo_job_id" {
#  value = openai_fine_tuning_job.dpo_example.id
#}
#
#output "dpo_job_status" {
#  value = openai_fine_tuning_job.dpo_example.status
#}

#output "dpo_fine_tuned_model" {
#  value = openai_fine_tuning_job.dpo_example.fine_tuned_model
#}

# Timeout example job outputs
output "timeout_job_id" {
  value = openai_fine_tuning_job.timeout_example.id
}

output "timeout_job_status" {
  value = openai_fine_tuning_job.timeout_example.status
}

output "timeout_job_created_at" {
  value = openai_fine_tuning_job.timeout_example.created_at
}

output "timeout_job_finished_at" {
  value = openai_fine_tuning_job.timeout_example.finished_at
}

# Add these lines at the bottom of main.tf to expose the checkpoint ID to test with
output "checkpoint_for_perms" {
  value       = openai_fine_tuning_job.supervised_example.fine_tuned_model
  description = "Checkpoint ID for permission testing from the supervised example job"
}

# Create the checkpoint permission using the fixed provider resource
# This requires an admin API key with appropriate permissions
resource "openai_fine_tuning_checkpoint_permission" "checkpoint_permission" {
  # Only create this resource if use_admin_key is true and both checkpoint_id and project_id are provided
  count = var.use_admin_key && var.checkpoint_id != "" && var.project_id != "" ? 1 : 0

  checkpoint_id = var.checkpoint_id
  project_ids   = [var.project_id]

  # Import hint: 
  # terraform import -var="admin_api_key=$OPENAI_ADMIN_KEY" -var="checkpoint_id=your-checkpoint-id" -var="project_id=your-project-id" -var="use_admin_key=true" openai_fine_tuning_checkpoint_permission.checkpoint_permission[0] cp_permission_id

  # Prevent modifications to imported resources
  lifecycle {
    ignore_changes = all
  }
}

# Note: The checkpoint_permission_id and checkpoint_permission_details outputs
# have been moved to checkpoints.tf to avoid duplication