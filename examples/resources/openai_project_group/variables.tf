# Optional variables for customizing the example

variable "dev_project_name" {
  description = "Name for the development project"
  type        = string
  default     = "Development Environment"
}

variable "prod_project_name" {
  description = "Name for the production project"
  type        = string
  default     = "Production API"
}
