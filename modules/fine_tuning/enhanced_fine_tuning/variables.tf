variable "model" {
  description = "The name of the base model to fine-tune"
  type        = string
}

variable "training_file" {
  description = "The ID of the training file uploaded to OpenAI"
  type        = string
}

variable "validation_file" {
  description = "The ID of the validation file uploaded to OpenAI"
  type        = string
  default     = null
}

variable "hyperparameters" {
  description = "Hyperparameters for the fine-tuning job"
  type        = map(any)
  default     = null
}

variable "suffix" {
  description = "A suffix to append to the fine-tuned model name"
  type        = string
  default     = null
}

variable "enable_monitoring" {
  description = "Whether to enable monitoring of the fine-tuning job"
  type        = bool
  default     = false
}

variable "enable_checkpoint_access" {
  description = "Whether to enable access to checkpoints for the fine-tuning job"
  type        = bool
  default     = false
}

variable "share_with_organizations" {
  description = "List of organization IDs to share the fine-tuned model with"
  type        = list(string)
  default     = []
}

variable "enabled" {
  description = "Whether to create the fine-tuned model"
  type        = bool
  default     = true
} 