# Fine-Tuning Data Source Examples
# This file demonstrates how to use all five fine-tuning data sources in the Terraform OpenAI provider
# IMPORTANT: By default, all examples are commented out to prevent errors during terraform apply
# To use any example, uncomment it and replace placeholder IDs with your actual IDs

# -------------------------------------------------------------------------
# HOW TO USE THESE EXAMPLES:
# 1. Uncomment the data source you want to use
# 2. Replace placeholder IDs with your actual IDs (such as those shown in the outputs)
# 3. Run terraform apply with appropriate permissions
# -------------------------------------------------------------------------

# Run with TF_LOG=DEBUG to see detailed debug statements:
# TF_LOG=DEBUG terraform apply -var="admin_api_key=$OPENAI_ADMIN_KEY" -target=data.openai_fine_tuning_checkpoint_permissions.example

# Example IDs from your environment (based on terraform output):
# - Job ID: ftjob-KYl8AIYu1vSDzh8bNWZ4O4Yx or ftjob-j3Ikg9y0PHP5LzSrltydhf0P
# - Model ID: ft:gpt-4o-mini-2024-07-18:fodoj-gmbh:my-custom-model-v1:BGvF3cDG
# - Checkpoint ID: ftckpt_fJFnoqNfnm4xpLh7khTuT6Qf (from first_checkpoint_id output)

# 1. Retrieve a specific fine-tuning job by ID
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE
*/
data "openai_fine_tuning_job" "example" {
  fine_tuning_job_id = "ftjob-rGLPqpdu0SiDseNWUHulw9wi" # Replace with your actual job ID
}

output "fine_tuning_job_example" {
  value = {
    id               = data.openai_fine_tuning_job.example.id
    model            = data.openai_fine_tuning_job.example.model
    status           = data.openai_fine_tuning_job.example.status
    fine_tuned_model = data.openai_fine_tuning_job.example.fine_tuned_model
    created_at       = data.openai_fine_tuning_job.example.created_at
    finished_at      = data.openai_fine_tuning_job.example.finished_at
    training_file    = data.openai_fine_tuning_job.example.training_file
  }
}
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE

# 2. List multiple fine-tuning jobs with optional filtering
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE
*/
data "openai_fine_tuning_jobs" "all" {
  # Optional parameters
  limit = 10
  # after = "ftjob-xyz789"  # For pagination
  # metadata = {            # Filter by specific metadata
  #   "project" = "customer-support"
  # }
}

output "fine_tuning_jobs_example" {
  value = [for job in data.openai_fine_tuning_jobs.all.jobs : {
    id               = job.id
    status           = job.status
    fine_tuned_model = job.fine_tuned_model
    created_at       = job.created_at
    model            = job.model
    hyperparameters  = job.hyperparameters
  }]
}

output "has_more_jobs" {
  value = data.openai_fine_tuning_jobs.all.has_more
}
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE

# 3. Get checkpoints for a specific fine-tuning job
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE
*/
#data "openai_fine_tuning_checkpoints" "example" {
#  fine_tuning_job_id = "ftjob-j3Ikg9y0PHP5LzSrltydhf0P" # Using the supervised example job ID
#
#  # Optional parameters
#  limit = 20
#  # after = "checkpoint-xyz789"  # For pagination
#}
#
#output "fine_tuning_checkpoints_example" {
#  value = data.openai_fine_tuning_checkpoints.example.checkpoints
#}
#
#output "first_checkpoint_id" {
#  value = length(data.openai_fine_tuning_checkpoints.example.checkpoints) > 0 ? data.openai_fine_tuning_checkpoints.example.checkpoints[0].id : null
#}
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE

# 4. Get events for a specific fine-tuning job
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE
*/
#data "openai_fine_tuning_events" "example" {
#  fine_tuning_job_id = "ftjob-KYl8AIYu1vSDzh8bNWZ4O4Yx" # Using the basic example job ID
#
#  # Optional parameters
#  limit = 50
#  # after = "event-xyz789"  # For pagination
#}
#
#output "fine_tuning_events_example" {
#  value = data.openai_fine_tuning_events.example.events
#}
#
#output "latest_event_message" {
#  value = length(data.openai_fine_tuning_events.example.events) > 0 ? data.openai_fine_tuning_events.example.events[0].message : ""
#}
#
## Filter events by level (info, warning, error)
#output "warning_events" {
#  value = [
#    for event in data.openai_fine_tuning_events.example.events :
#    event.message if event.level == "warning"
#  ]
#}
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE

# 5. Get checkpoint permissions (requires admin privileges and appropriate scopes)
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE
*/
#data "openai_fine_tuning_checkpoint_permissions" "example" {
#  # IMPORTANT: Use the fine-tuned model ID format, not the checkpoint ID format
#  checkpoint_id = "ft:gpt-4o-mini-2024-07-18:fodoj-gmbh::BGvDTdTK" # Example from the supervised job
#
#  # Optional parameters
#  limit = 20
#  # after = "permission-xyz789"  # For pagination
#
#  # The admin API key is automatically read from OPENAI_ADMIN_KEY environment variable
#  # No need to set admin_api_key explicitly
#}
#
#output "checkpoint_permissions_example" {
#  value = data.openai_fine_tuning_checkpoint_permissions.example.permissions
#}
#
#output "admin_key_length" {
#  value     = "Admin key length from var: ${length(var.admin_api_key)}"
#  sensitive = true
#}
/* COMMENTING OUT THIS LINE TO ENABLE THE EXAMPLE

# Comment out the examples above when applying, and use this realistic workflow instead:
# To enable this example, remove the surrounding comment markers and replace placeholder IDs

/*
# REALISTIC WORKFLOW EXAMPLE
# This demonstrates how all data sources can work together

# 1. First, get information about all your fine-tuning jobs
data "openai_fine_tuning_jobs" "recent" {
  limit = 5
}

# 2. Get detailed information about your most recent job
data "openai_fine_tuning_job" "latest" {
  fine_tuning_job_id = data.openai_fine_tuning_jobs.recent.jobs[0].id
}

# 3. Get checkpoints for the job if it's complete
data "openai_fine_tuning_checkpoints" "latest_checkpoints" {
  count = data.openai_fine_tuning_job.latest.status == "succeeded" ? 1 : 0
  fine_tuning_job_id = data.openai_fine_tuning_job.latest.id
}

# 4. Get events to monitor the job's progress
data "openai_fine_tuning_events" "latest_events" {
  fine_tuning_job_id = data.openai_fine_tuning_job.latest.id
  limit = 20
}

# 5. Check permissions for a checkpoint if available (requires admin privileges)
data "openai_fine_tuning_checkpoint_permissions" "latest_permissions" {
  count = data.openai_fine_tuning_job.latest.status == "succeeded" ? 1 : 0
  # Use the fine_tuned_model output from the job as the checkpoint_id
  checkpoint_id = data.openai_fine_tuning_job.latest.fine_tuned_model
}

# Comprehensive output
output "fine_tuning_workflow_summary" {
  value = {
    job_id = data.openai_fine_tuning_job.latest.id
    status = data.openai_fine_tuning_job.latest.status
    model = data.openai_fine_tuning_job.latest.model
    fine_tuned_model = data.openai_fine_tuning_job.latest.fine_tuned_model
    created_at = data.openai_fine_tuning_job.latest.created_at
    checkpoint_count = data.openai_fine_tuning_job.latest.status == "succeeded" ? length(data.openai_fine_tuning_checkpoints.latest_checkpoints[0].checkpoints) : 0
    event_count = length(data.openai_fine_tuning_events.latest_events.events)
    latest_event = length(data.openai_fine_tuning_events.latest_events.events) > 0 ? data.openai_fine_tuning_events.latest_events.events[0].message : null
  }
}
*/

# TROUBLESHOOTING TIPS:
# 1. 404 Error: Make sure to use valid job/checkpoint IDs that exist in your account
# 2. 401 Unauthorized: Checkpoint permissions require admin privileges and the api.fine_tuning.checkpoints.read scope
# 3. Type conversion issues (RESOLVED): 
#    - The provider now automatically converts numeric values for hyperparameters to strings
#    - All data sources should be working correctly now
# 4. 500 Internal Server Error: For checkpoint permissions, make sure to use the model ID format (ft:...) and not the checkpoint ID format (ftckpt_...)
