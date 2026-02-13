# Create a group
resource "openai_group" "engineering" {
  name = "Engineering"
}

# Add a user to a group
resource "openai_group_user" "engineer" {
  group_id = openai_group.engineering.id
  user_id  = "user-abc123"
}

# Reference an existing group and add multiple users
resource "openai_group_user" "support_agent_1" {
  group_id = "group-xyz789"
  user_id  = "user-def456"
}

resource "openai_group_user" "support_agent_2" {
  group_id = "group-xyz789"
  user_id  = "user-ghi789"
}

# Output user details
output "engineer_email" {
  value = openai_group_user.engineer.email
}

output "engineer_name" {
  value = openai_group_user.engineer.user_name
}
