terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "~> 1.0.0"
    }
  }
  required_version = ">= 1.0.0"
}

provider "openai" {
  organization = "fodoj"
  admin_key    = var.openai_admin_key
}

variable "project_id" {
  description = "The OpenAI project ID"
  type        = string
  default     = "proj_lZBcynA3G9hY3fJa3gc8AS3T"
}

variable "ignore_rate_limit_warning" {
  description = "Set to true to acknowledge that OpenAI rate limits cannot be truly deleted and will be reset to defaults on removal"
  type        = bool
  default     = true # Set to true by default to make deletion testing easier
}

variable "openai_admin_key" {
  description = "The OpenAI admin API key"
  type        = string
}

# When using openai_rate_limit resources, it's important to understand what happens during deletion.
# When you run "terraform destroy", the provider will:
#  1. Reset the rate limit to default values (not truly delete it, as that's not supported by the API)
#  2. The default values are specific to each model (e.g., dall-e-3 has different defaults than gpt-4o-mini)
#  3. The provider now handles this reset process reliably, resolving previous errors with type conversions
#  4. The fixes ensure a consistent and panic-free experience during deletion operations

# This resource sets rate limits for DALL-E 3
resource "openai_rate_limit" "dalle3_rate_limit" {
  project_id                = var.project_id
  model                     = "dall-e-3"
  max_requests_per_minute   = 999
  ignore_rate_limit_warning = true # This must be true to allow deletion
}

# This resource sets rate limits for GPT-4o-mini
resource "openai_rate_limit" "gpt4o_mini_limit" {
  project_id                   = var.project_id
  model                        = "gpt-4o-mini"
  max_requests_per_minute      = 30
  max_tokens_per_minute        = 1000
  max_images_per_minute        = 1
  batch_1_day_max_input_tokens = 100000
  ignore_rate_limit_warning    = true # This must be true to allow deletion
}
