variable "use_data_source" {
  description = "Whether to use a data source for existing rate limits"
  type        = bool
  default     = false
}

variable "list_mode" {
  description = "Whether to retrieve all rate limits instead of working with a single rate limit"
  type        = bool
  default     = false
}

variable "project_id" {
  description = "The ID of the project (format: proj_abc123)"
  type        = string
}

variable "model" {
  description = "The model to get rate limits for (e.g., 'gpt-4', 'gpt-3.5-turbo')"
  type        = string
}

variable "max_requests_per_minute" {
  description = "Maximum number of API requests allowed per minute"
  type        = number
  default     = null
}

variable "max_tokens_per_minute" {
  description = "Maximum number of tokens that can be processed per minute"
  type        = number
  default     = null
}

variable "max_images_per_minute" {
  description = "Maximum number of images that can be generated per minute (for image models like DALL-E)"
  type        = number
  default     = null
}

variable "batch_1_day_max_input_tokens" {
  description = "Maximum number of input tokens allowed in batch operations per day"
  type        = number
  default     = null
}

variable "max_audio_megabytes_per_1_minute" {
  description = "Maximum number of audio megabytes that can be processed per minute (for audio models like Whisper)"
  type        = number
  default     = null
}

variable "max_requests_per_1_day" {
  description = "Maximum number of API requests allowed per day"
  type        = number
  default     = null
}

variable "openai_admin_key" {
  description = "Admin API key to use for rate limiting operations. If not provided, the provider's default API key will be used."
  type        = string
  default     = null
  sensitive   = true
} 