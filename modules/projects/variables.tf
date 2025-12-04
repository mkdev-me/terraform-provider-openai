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

