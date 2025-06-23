# List all API keys for a specific project
data "openai_project_api_keys" "example" {
  project_id = "proj_abc123def456" # Replace with your project ID
}

# Output API key count
output "api_key_count" {
  value = length(data.openai_project_api_keys.example.api_keys)
}

