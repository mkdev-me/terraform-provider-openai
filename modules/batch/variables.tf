# Input variables for the batch module

variable "input_file_id" {
  description = "The ID of the file containing the inputs for the batch (must be uploaded with purpose='batch')"
  type        = string
  default     = ""
}

variable "project_id" {
  description = "The ID of the OpenAI project to use for this batch. If not specified, the default project will be used."
  type        = string
  default     = ""
}

variable "endpoint" {
  description = "The endpoint to use for all requests in the batch (e.g., '/v1/chat/completions')"
  type        = string
  default     = "/v1/chat/completions"

  validation {
    condition     = contains(["/v1/responses", "/v1/chat/completions", "/v1/embeddings", "/v1/completions"], var.endpoint)
    error_message = "The endpoint must be one of: '/v1/responses', '/v1/chat/completions', '/v1/embeddings', '/v1/completions'."
  }
}

variable "completion_window" {
  description = "The time frame within which the batch should be processed (currently only '24h' is supported)"
  type        = string
  default     = "24h"
}

variable "model" {
  description = "The ID of the model to use for this batch"
  type        = string
  default     = "gpt-3.5-turbo"
}

variable "metadata" {
  description = "Set of key-value pairs that can be attached to the batch object (max 16 pairs)"
  type        = map(string)
  default     = {}
}

variable "list_mode" {
  description = "Whether to retrieve all batches instead of working with a single batch"
  type        = bool
  default     = false
}

variable "batch_id" {
  description = "The ID of an existing batch to retrieve (only used when not in list_mode)"
  type        = string
  default     = ""
}

variable "api_key" {
  description = "The OpenAI API key to use for authentication. If not provided, uses the provider's default."
  type        = string
  default     = ""
  sensitive   = true
} 