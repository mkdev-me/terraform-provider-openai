terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

provider "openai" {
  # API key and organization ID can be set using environment variables:
  # OPENAI_API_KEY and OPENAI_ORGANIZATION_ID
}

# Direct resource declarations instead of modules
resource "openai_moderation" "single_text" {
  input = "I want to kill them."
  # Note: The provider automatically handles model version differences.
  # If the API returns "text-moderation-007" but config specifies "text-moderation-latest",
  # terraform won't try to replace the resource on subsequent applies.

  # For extra protection, we also add a lifecycle block
  lifecycle {
    ignore_changes = [model]
  }
}

resource "openai_moderation" "harmless_text" {
  input = "This is a completely harmless message."
  model = "text-moderation-latest"

  # For extra protection, we also add a lifecycle block
  lifecycle {
    ignore_changes = [model]
  }
}

resource "openai_moderation" "harmful_text" {
  input = "I want to make a bomb and hurt people."
  model = "text-moderation-latest"

  # For extra protection, we also add a lifecycle block
  lifecycle {
    ignore_changes = [model]
  }
}

resource "openai_moderation" "neutral_text" {
  input = "I like watching movies and reading books."
  model = "text-moderation-latest"

  # For extra protection, we also add a lifecycle block
  lifecycle {
    ignore_changes = [model]
  }
}

# Outputs
output "single_text_flagged" {
  description = "Whether the single text was flagged"
  value       = openai_moderation.single_text.flagged
}

output "single_text_categories" {
  description = "Categories flagged for the single text"
  value       = openai_moderation.single_text.categories
}

output "single_text_category_scores" {
  description = "Category scores for the single text"
  value       = openai_moderation.single_text.category_scores
}

output "batch_texts_flagged" {
  description = "Whether each text in the batch was flagged"
  value = {
    "text1" = openai_moderation.harmless_text.flagged
    "text2" = openai_moderation.harmful_text.flagged
    "text3" = openai_moderation.neutral_text.flagged
  }
}

output "batch_texts_categories" {
  description = "Categories flagged for each text in the batch"
  value = {
    "text1" = openai_moderation.harmless_text.categories
    "text2" = openai_moderation.harmful_text.categories
    "text3" = openai_moderation.neutral_text.categories
  }
}
