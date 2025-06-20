terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

# Content moderation resource
resource "openai_moderation" "this" {
  input = can(tolist(var.input)) ? jsonencode(var.input) : tostring(var.input)
  model = var.model

  # Prevent replacement due to model version differences
  lifecycle {
    ignore_changes = [model]
  }
}

# Outputs
output "id" {
  description = "The ID of the moderation response"
  value       = openai_moderation.this.id
}

output "model" {
  description = "The model used for moderation"
  value       = openai_moderation.this.model
}

output "results" {
  description = "The moderation results including categories and category scores"
  value       = openai_moderation.this.results
}

output "flagged" {
  description = "Whether the input was flagged by the moderation model"
  value       = openai_moderation.this.flagged
}

output "categories" {
  description = "The content categories that were flagged (if any)"
  value       = openai_moderation.this.categories
}

output "category_scores" {
  description = "The scores for each content category"
  value       = openai_moderation.this.category_scores
} 