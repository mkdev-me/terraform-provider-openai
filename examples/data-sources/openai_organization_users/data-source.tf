# List all users in the organization
data "openai_organization_users" "all" {
  # Optional: filter by role
  role = "owner"
}

# List all organization users without filtering
data "openai_organization_users" "everyone" {}

# Output user statistics
output "all_users_count" {
  value = length(data.openai_organization_users.everyone.users)
}
