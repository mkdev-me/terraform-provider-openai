# Example of retrieving and using the OpenAI admin API keys data source

# List all admin API keys with default parameters
data "openai_admin_api_keys" "default" {
  # Using the provider's default admin API key for authentication
}

# EXAMPLE: List admin API keys with custom pagination
# Uncomment and replace key_abc123 with an actual key ID to use pagination
# data "openai_admin_api_keys" "paginated" {
#   limit = 5            # Only retrieve up to 5 keys
#   after = "key_abc123" # Replace with an actual key ID from your account
# }

# List admin API keys using a custom API key for authentication
data "openai_admin_api_keys" "custom_auth" {
  api_key = var.openai_admin_key # Custom API key for authentication
  limit   = 10
}

# Output: Total number of API keys from default query
output "admin_key_count" {
  description = "Total number of admin API keys"
  value       = length(data.openai_admin_api_keys.default.api_keys)
}

# Output: Names of all API keys
output "admin_key_names" {
  description = "Names of all admin API keys"
  value       = [for key in data.openai_admin_api_keys.default.api_keys : key.name]
}

# Output: Map of key IDs to their names
output "admin_key_id_to_name_map" {
  description = "Map of key IDs to their names"
  value       = { for key in data.openai_admin_api_keys.default.api_keys : key.id => key.name }
}

# Output: API keys that expire in the future
output "expiring_admin_keys" {
  description = "API keys that have an expiration date"
  value = [
    for key in data.openai_admin_api_keys.default.api_keys : {
      id         = key.id
      name       = key.name
      expires_at = key.expires_at
    } if lookup(key, "expires_at", null) != null
  ]
}

# Output: Whether pagination is needed
output "has_more_keys" {
  description = "Whether there are more keys beyond this page"
  value       = data.openai_admin_api_keys.default.has_more
} 