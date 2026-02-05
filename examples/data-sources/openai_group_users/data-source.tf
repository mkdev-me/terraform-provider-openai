# List all users in a group
data "openai_group_users" "engineering" {
  group_id = "group-abc123"
}

# Output total user count
output "total_users" {
  value = data.openai_group_users.engineering.user_count
}

# Output all user IDs
output "all_user_ids" {
  value = data.openai_group_users.engineering.user_ids
}

# Output detailed user information
output "user_details" {
  value = [
    for user in data.openai_group_users.engineering.users : {
      id       = user.user_id
      name     = user.user_name
      email    = user.email
      role     = user.role
      added_at = user.added_at
    }
  ]
}
