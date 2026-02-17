# Look up a user in a group by user ID
data "openai_group_user" "by_id" {
  group_id = "group-abc123"
  user_id  = "user-xyz789"
}

# Look up a user in a group by email
data "openai_group_user" "by_email" {
  group_id = "group-abc123"
  email    = "engineer@example.com"
}

# Output user details
output "user_name" {
  value = data.openai_group_user.by_email.user_name
}

output "is_service_account" {
  value = data.openai_group_user.by_email.is_service_account
}
