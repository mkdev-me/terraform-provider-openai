# Example: Creating an Admin API Key for organization management
# Admin API keys are used for administrative operations like managing users, projects, and rate limits
# Note: This requires organization admin privileges

resource "openai_admin_api_key" "org_admin" {
  # Name for the admin API key - used for identification in the dashboard
  name = "terraform-admin-key"

  # Scopes for the admin key
  # Available scopes: 
  # - "users.read", "users.write" - manage organization users
  # - "projects.read", "projects.write" - manage projects
  # - "api_keys.read", "api_keys.write" - manage API keys
  # - "rate_limits.read", "rate_limits.write" - manage rate limits
  scopes = [
    "users.read",
    "users.write",
    "projects.read",
    "projects.write",
    "api_keys.read",
    "rate_limits.read",
    "rate_limits.write"
  ]

  # Optional: Set expiration date for the key (Unix timestamp)
  # expires_at = 1735689599  # 2024-12-31T23:59:59Z
}

# Output the created admin API key ID
output "admin_key_id" {
  value = openai_admin_api_key.org_admin.id
}

