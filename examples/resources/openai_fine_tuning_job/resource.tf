# Upload training data file
resource "openai_file" "training_data" {
  file    = "training_data.jsonl"
  purpose = "fine-tune"
}

# Upload validation data file (optional)
resource "openai_file" "validation_data" {
  file    = "validation_data.jsonl"
  purpose = "fine-tune"
}

# Create a fine-tuning job
resource "openai_fine_tuning_job" "custom_model" {
  training_file = openai_file.training_data.id
  model         = "gpt-3.5-turbo-0125" # Base model to fine-tune

  # Optional: Add validation file for better model evaluation
  validation_file = openai_file.validation_data.id

  # Optional: Configure hyperparameters
  hyperparameters {
    n_epochs                 = 3
    batch_size               = 4
    learning_rate_multiplier = 0.1
  }

  # Optional: Set a custom suffix for the fine-tuned model name
  suffix = "customer-support-v2"

  # Optional: Add metadata
  metadata = {
    department = "customer-service"
    use_case   = "support-automation"
    version    = "2.0"
    trained_by = "ml-team"
  }
}

# Create another fine-tuning job with minimal configuration
resource "openai_fine_tuning_job" "simple_model" {
  training_file = openai_file.training_data.id
  model         = "gpt-3.5-turbo-0125"

  # Use default hyperparameters
  suffix = "basic-v1"
}

# Create a fine-tuning job with specific seed for reproducibility
resource "openai_fine_tuning_job" "reproducible_model" {
  training_file = openai_file.training_data.id
  model         = "gpt-3.5-turbo-0125"

  hyperparameters {
    n_epochs                 = 5
    batch_size               = 8
    learning_rate_multiplier = 0.05
  }

  # Set seed for reproducible results
  seed = 42

  suffix = "reproducible-v1"

  metadata = {
    experiment_id = "EXP-2024-001"
    reproducible  = "true"
  }
}

# Output the fine-tuning job details
output "fine_tuning_job_id" {
  value       = openai_fine_tuning_job.custom_model.id
  description = "The ID of the fine-tuning job"
}
