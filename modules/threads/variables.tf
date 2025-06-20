# OpenAI Threads Module Variables

variable "enable_thread" {
  description = "Whether to create the thread"
  type        = bool
  default     = true
}

variable "thread_messages" {
  description = "List of initial messages to include in the thread"
  type = list(object({
    role     = string
    content  = string
    file_ids = optional(list(string))
    metadata = optional(map(string))
  }))
  default = []
}

variable "thread_metadata" {
  description = "Metadata for the thread"
  type        = map(string)
  default     = {}
}

variable "enable_thread_data_source" {
  description = "Whether to fetch a thread by ID"
  type        = bool
  default     = false
}

variable "thread_id" {
  description = "ID of a specific thread to fetch (required if enable_thread_data_source is true)"
  type        = string
  default     = null
} 