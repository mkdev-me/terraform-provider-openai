# Set rate limits for GPT-4o model in a project
resource "openai_rate_limit" "gpt4o_limits" {
  project_id = "proj_abc123"
  model      = "gpt-4o"

  max_requests_per_minute = 500
  max_tokens_per_minute   = 30000
}

# Set rate limits for GPT-4o-mini with additional constraints
resource "openai_rate_limit" "gpt4o_mini_limits" {
  project_id = "proj_abc123"
  model      = "gpt-4o-mini"

  max_requests_per_minute = 1000
  max_tokens_per_minute   = 60000
  max_requests_per_1_day  = 10000
}

# Set rate limits for DALL-E 3 image generation
resource "openai_rate_limit" "dalle3_limits" {
  project_id = "proj_abc123"
  model      = "dall-e-3"

  max_images_per_minute = 5
}

# Set rate limits for batch processing
resource "openai_rate_limit" "batch_limits" {
  project_id = "proj_abc123"
  model      = "gpt-4o"

  batch_1_day_max_input_tokens = 1000000
}
