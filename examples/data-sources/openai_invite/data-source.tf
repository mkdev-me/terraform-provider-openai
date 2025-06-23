# Fetch a specific invite by ID
data "openai_invite" "developer_invite" {
  invite_id = "invite-abc123"
}

# Output invite status
output "invite_status" {
  value = data.openai_invite.developer_invite.status
}
