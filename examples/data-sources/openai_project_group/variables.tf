# Optional variables for customizing the example

variable "project_id" {
  description = "The ID of the project to look up groups from"
  type        = string
  default     = "proj-abc123"
}

variable "group_id" {
  description = "The ID of the group to look up"
  type        = string
  default     = "group-xyz789"
}
