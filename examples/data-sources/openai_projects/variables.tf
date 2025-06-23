variable "openai_admin_key" {
  description = "OpenAI admin API key. If not provided, uses OPENAI_ADMIN_KEY environment variable"
  type        = string
  sensitive   = true
  default     = null
}

