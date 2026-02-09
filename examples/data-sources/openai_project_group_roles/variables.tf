# Optional variables for customizing the example

variable "project_id" {
  description = "The ID of the project"
  type        = string
  default     = "proj_abc123"
}

variable "group_id" {
  description = "The ID of the group to retrieve role assignments for"
  type        = string
  default     = "group_01J1F8ABCDXYZ"
}
