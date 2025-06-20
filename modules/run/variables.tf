# Input variables for the OpenAI Run module

variable "thread_id" {
  description = "The ID of the thread to run the assistant on"
  type        = string
}

variable "assistant_id" {
  description = "The ID of the assistant to use for the run"
  type        = string
}

variable "model" {
  description = "Override the model used by the assistant for this run"
  type        = string
  default     = null
}

variable "instructions" {
  description = "Override the instructions used by the assistant for this run"
  type        = string
  default     = null
}

variable "tools" {
  description = "Override the tools the assistant can use for this run"
  type        = list(map(string))
  default     = null
}

variable "metadata" {
  description = "Metadata to associate with the run"
  type        = map(string)
  default     = {}
}

variable "temperature" {
  description = "The sampling temperature (0-2). Higher values make output more random, lower values more deterministic"
  type        = number
  default     = null

  validation {
    condition     = var.temperature == null ? true : (var.temperature >= 0 && var.temperature <= 2)
    error_message = "Temperature must be between 0 and 2."
  }
}

variable "max_completion_tokens" {
  description = "The maximum number of tokens to generate for completion"
  type        = number
  default     = null
}

variable "top_p" {
  description = "Alternative to temperature for nucleus sampling (0-1)"
  type        = number
  default     = null

  validation {
    condition     = var.top_p == null ? true : (var.top_p >= 0 && var.top_p <= 1)
    error_message = "Top p must be between 0 and 1."
  }
}

variable "stream_for_tool" {
  description = "Whether to stream tool outputs"
  type        = bool
  default     = null
} 