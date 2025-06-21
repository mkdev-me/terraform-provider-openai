terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
    }
  }
}

provider "openai" {
  # API key and organization ID can be set using environment variables:
  # OPENAI_API_KEY and OPENAI_ORGANIZATION_ID
}

# Direct resource usage instead of module
resource "openai_moderation" "example1" {
  input = "This is a completely harmless message."
  model = "text-moderation-latest"

  # This lifecycle block is critical to prevent resource replacement on every apply.
  # The OpenAI API returns a specific model version (e.g., text-moderation-007) even when
  # we request text-moderation-latest, causing Terraform to detect a change in state.
  lifecycle {
    ignore_changes = [model]
  }
}

resource "openai_moderation" "example2" {
  input = "I want to make a bomb and hurt people."
  model = "text-moderation-latest"

  # Prevent replacement due to model version differences
  lifecycle {
    ignore_changes = [model]
  }
}

resource "openai_moderation" "example3" {
  input = "I like watching movies and reading books."
  model = "text-moderation-latest"

  # Prevent replacement due to model version differences
  lifecycle {
    ignore_changes = [model]
  }
}

# Outputs
output "direct_example1_flagged" {
  description = "Whether the first example was flagged"
  value       = openai_moderation.example1.flagged
}

output "direct_example1_categories" {
  description = "Categories flagged for the first example"
  value       = openai_moderation.example1.categories
}

output "direct_example1_category_scores" {
  description = "Category scores for the first example"
  value       = openai_moderation.example1.category_scores
}

output "direct_texts_flagged" {
  description = "Whether each example was flagged"
  value = {
    "text1" = openai_moderation.example1.flagged
    "text2" = openai_moderation.example2.flagged
    "text3" = openai_moderation.example3.flagged
  }
}

output "direct_texts_categories" {
  description = "Categories flagged for each example"
  value = {
    "text1" = openai_moderation.example1.categories
    "text2" = openai_moderation.example2.categories
    "text3" = openai_moderation.example3.categories
  }
} 