# Example: Managing rate limits for OpenAI projects
# Rate limits control API usage to prevent abuse and ensure fair resource allocation
# Note: This requires admin API key with rate_limits.write permission

# First, let's reference a project (or create one)
resource "openai_project" "production" {
  name        = "production-app"
  description = "Production environment project"
}

# Set rate limits for GPT-4 model in the project
resource "openai_rate_limit" "gpt4_limits" {
  # The project to apply rate limits to
  project_id = openai_project.production.id

  # The model to limit
  model = "gpt-4"

  # Maximum requests per minute (RPM)
  max_requests_per_minute = 100

  # Maximum tokens per minute (TPM)
  max_tokens_per_minute = 40000

  # Maximum requests per day
  max_requests_per_1_day = 2000

  # Note: max_tokens_per_day is not a supported parameter in the schema

  # Maximum images per minute (for image models)
  # max_images_per_minute = 50
}

# Set different limits for GPT-3.5-turbo (higher limits for cheaper model)
resource "openai_rate_limit" "gpt35_limits" {
  project_id = openai_project.production.id
  model      = "gpt-3.5-turbo"

  max_requests_per_minute = 500
  max_tokens_per_minute   = 90000
  max_requests_per_1_day  = 10000
}

# Rate limits for embeddings model
resource "openai_rate_limit" "embeddings_limits" {
  project_id = openai_project.production.id
  model      = "text-embedding-ada-002"

  max_requests_per_minute = 1000
  max_tokens_per_minute   = 1000000
  max_requests_per_1_day  = 50000
}

# Rate limits for DALL-E image generation
resource "openai_rate_limit" "dalle_limits" {
  project_id = openai_project.production.id
  model      = "dall-e-3"

  max_requests_per_minute = 5
  max_images_per_minute   = 5
  max_requests_per_1_day  = 100
}

# Development project with lower limits
resource "openai_project" "development" {
  name = "dev-environment"
}

resource "openai_rate_limit" "dev_gpt4_limits" {
  project_id = openai_project.development.id
  model      = "gpt-4"

  # Lower limits for development
  max_requests_per_minute = 10
  max_tokens_per_minute   = 8000
  max_requests_per_1_day  = 200
}

# Staging environment with moderate limits
resource "openai_project" "staging" {
  name = "staging-environment"
}

resource "openai_rate_limit" "staging_limits" {
  project_id = openai_project.staging.id
  model      = "gpt-4"

  # Moderate limits for staging
  max_requests_per_minute = 50
  max_tokens_per_minute   = 20000
  max_requests_per_1_day  = 1000
}

# Output the production project ID
output "production_project_id" {
  value = openai_project.production.id
}
