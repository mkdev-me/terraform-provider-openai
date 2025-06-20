# Module variables for the OpenAI project module

# Mode control
variable "create_project" {
  type        = bool
  description = "Whether to create a new project or use the data source"
  default     = true
}

variable "list_mode" {
  type        = bool
  description = "Whether to retrieve all projects instead of working with a single project"
  default     = false
}

# Project creation variables
variable "name" {
  type        = string
  description = "Name of the project (required when create_project is true)"
  default     = null
}

# Data source variables
variable "project_id" {
  type        = string
  description = "ID of the OpenAI project to retrieve information about (required when create_project is false)"
  default     = null
}

# Authentication variables
variable "openai_admin_key" {
  description = "Admin API key to use for project operations. If not provided, the provider's default API key will be used."
  type        = string
  default     = null
  sensitive   = true
}

variable "organization_id" {
  type        = string
  description = "OpenAI Organization ID (org-xxxx)"
  default     = ""
}

variable "rate_limits" {
  description = "Rate limits for the project"
  type = list(object({
    model                        = string
    max_requests_per_minute      = optional(number)
    max_tokens_per_minute        = optional(number)
    max_images_per_minute        = optional(number)
    batch_1_day_max_input_tokens = optional(number)
  }))
  default = []
}

variable "users" {
  description = "Users to add to the project"
  type = list(object({
    user_id = string
    role    = string
  }))
  default = []
}

variable "is_default" {
  description = "Whether this project should be the default project"
  type        = bool
  default     = false
}
