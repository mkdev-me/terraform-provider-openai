provider "openai" {
  # OpenAI provider configuration
  # Note: OpenAI API key is usually set via the OPENAI_API_KEY environment variable
}

# Example: Development Project API Key
module "dev_project_api_key" {
  source = "../../project_api"

  # Project ID from the OpenAI dashboard
  project_id = "proj_abc123"

  # Name should match the key name in the OpenAI dashboard
  name = "dev-key"
}

# Example: Production Project API Key
module "prod_project_api_key" {
  source = "../../project_api"

  project_id = "proj_xyz789"
  name       = "prod-key"
}

# Outputs for reference
output "dev_api_key_id" {
  description = "The ID of the imported development project API key"
  value       = module.dev_project_api_key.api_key_id
}

output "prod_api_key_id" {
  description = "The ID of the imported production project API key"
  value       = module.prod_project_api_key.api_key_id
}

# Import commands for the examples:
#
# For the development key:
# terraform import module.dev_project_api_key.openai_project_api_key.this "proj_abc123:key_dev456"
#
# For the production key:
# terraform import module.prod_project_api_key.openai_project_api_key.this "proj_xyz789:key_prod123" 