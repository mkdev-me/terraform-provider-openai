variable "model" {
  description = "The name of the model to fine-tune (e.g., gpt-3.5-turbo)"
  type        = string
}

variable "training_file" {
  description = "The ID of an uploaded file that contains training data"
  type        = string
}

variable "validation_file" {
  description = "The ID of an uploaded file that contains validation data"
  type        = string
  default     = null
}

variable "hyperparameters" {
  description = "Hyperparameters for the fine-tuning job"
  type        = map(any)
  default     = null
}

variable "suffix" {
  description = "A string of up to 64 characters that will be added to your fine-tuned model name"
  type        = string
  default     = null
}

variable "completion_window" {
  description = "Time in seconds to wait for job to complete during creation. 0 means don't wait."
  type        = number
  default     = 0
} 