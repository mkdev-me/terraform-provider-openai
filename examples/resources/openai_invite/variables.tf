variable "openai_api_key" {
  description = "OpenAI API key. If not provided, uses OPENAI_API_KEY environment variable"
  type        = string
  sensitive   = true
  default     = null
}

