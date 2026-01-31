# List all users in the organization
data "openai_organization_users" "all" {}

# List all organization users without filtering
data "openai_organization_users" "everyone" {}

# Output user statistics
output "all_users_count" {
  value = length(data.openai_organization_users.everyone.users)
}
