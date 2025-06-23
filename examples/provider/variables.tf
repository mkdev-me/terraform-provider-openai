variable "openai_api_key" {
  description = "OpenAI API key"
  type        = string
  sensitive   = true
}

variable "openai_admin_key" {
  description = "OpenAI admin API key for organization management"
  type        = string
  sensitive   = true
}

variable "organization_id" {
  description = "OpenAI organization ID"
  type        = string
}

