# Email Lookup Examples for OpenAI User Data Sources
# --------------------------------------------------
# This file demonstrates how to use email addresses instead of user IDs
# with the OpenAI data source resources.

# 1. Retrieve project user information using email
data "openai_project_user" "by_email" {
  project_id = openai_project.data_source_example.id
  email      = "pablo@mkdev.me" # Replace with a real email from your project

  # Data source will only be evaluated after the user is added to the project
  depends_on = [openai_project_user.data_source_user]
}

# Outputs - Project User Data Source (by email)
output "project_user_by_email_id" {
  description = "The user ID retrieved by email lookup in the project"
  value       = data.openai_project_user.by_email.user_id
}

output "project_user_by_email_role" {
  description = "The user role retrieved by email lookup in the project"
  value       = data.openai_project_user.by_email.role
}

output "project_user_by_email_added_at" {
  description = "When the user was added to the project (retrieved by email lookup)"
  value       = data.openai_project_user.by_email.added_at
}

