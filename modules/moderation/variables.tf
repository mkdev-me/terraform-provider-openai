variable "input" {
  description = "Input (or inputs) to classify. Can be a single string, an array of strings, or an array of multi-modal input objects."
  type        = any
}

variable "model" {
  description = "The content moderation model to use (e.g., omni-moderation-latest)"
  type        = string
  default     = null
} 