variable "list_mode" {
  description = "Whether to operate in list mode or single user mode"
  type        = bool
  default     = false
}

variable "user_id" {
  description = "The ID of the user to retrieve (required in single user mode)"
  type        = string
  default     = null
}

variable "after" {
  description = "A cursor for pagination (list mode only)"
  type        = string
  default     = null
}

variable "limit" {
  description = "The number of users to return (list mode only)"
  type        = number
  default     = 20
}

variable "emails" {
  description = "List of email addresses to filter by (list mode only)"
  type        = list(string)
  default     = []
}

variable "api_key" {
  description = "Custom API key to use for this module"
  type        = string
  default     = null
  sensitive   = true
} 