# Email Lookup Examples for OpenAI Organization User Data Source
# --------------------------------------------------
# This file demonstrates how to use email addresses instead of user IDs
# when retrieving user information from the OpenAI organization.

# Retrieve organization user information using email
data "openai_organization_user" "by_email" {
  email = "pablo+newtest7@mkdev.me" # Replace with a real email from your organization

  # Optional: Use a specific API key with sufficient permissions
  # api_key = var.openai_admin_key
}

# Outputs from the organization user data source
output "org_user_by_email_id" {
  description = "The user ID retrieved by email lookup in the organization"
  value       = data.openai_organization_user.by_email.user_id
}

output "org_user_by_email_name" {
  description = "The user name retrieved by email lookup in the organization"
  value       = data.openai_organization_user.by_email.name
}

output "org_user_by_email_role" {
  description = "The user role retrieved by email lookup in the organization"
  value       = data.openai_organization_user.by_email.role
}

output "org_user_by_email_added_at" {
  description = "When the user was added to the organization (retrieved by email lookup)"
  value       = data.openai_organization_user.by_email.added_at
}

# Additional example - practical use case: finding a user by email and using their ID in a resource
# This approach can be useful when you know user emails but not their IDs

# First, retrieve the user ID using the email lookup
data "openai_organization_user" "find_by_email" {
  email = "pablo+newtest8@mkdev.me" # Replace with a real email
}

# Then, use the retrieved user ID to create a new resource
# (For demonstration - uncomment if needed)
# resource "openai_organization_user_role" "update_user_role" {
#   user_id = data.openai_organization_user.find_by_email.user_id
#   role    = "member"
# } 