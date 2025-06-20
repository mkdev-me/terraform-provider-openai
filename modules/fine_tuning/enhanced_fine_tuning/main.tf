terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

# Create a fine-tuning job
resource "openai_fine_tuned_model" "model" {
  count = var.enabled ? 1 : 0

  model           = var.model
  training_file   = var.training_file
  validation_file = var.validation_file
  suffix          = var.suffix

  # Hyperparameters should be a block, not an argument
  hyperparameters {
    n_epochs = "auto"
  }
}

# It appears the fine-tuned model resource doesn't have a fine_tuning_job_id attribute
# Since we can't get the job ID, we'll use placeholder values for the data sources
# In a real implementation, you would need to get the job ID through other means

# Note: We've removed all the data sources that were returning 404 errors
# We've also removed the rate limiting logic since the API endpoint doesn't seem to exist

# Commented out checkpoint permissions since the resource requires different arguments than expected
/*
# Optionally share checkpoints with other organizations
resource "openai_fine_tuning_checkpoint_permission" "checkpoint_permissions" {
  count = 0 # Disabled
  
  # According to the error message, the resource requires these arguments:
  project_ids = ["proj_abc123"]
  fine_tuned_model_checkpoint = "checkpoint_abc123"
}
*/ 