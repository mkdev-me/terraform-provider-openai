# OpenAI Assistants Module Variables

variable "enable_assistant" {
  description = "Whether to create the assistant"
  type        = bool
  default     = true
}

variable "assistant_name" {
  description = "The name of the assistant"
  type        = string
  default     = null
}

variable "assistant_model" {
  description = "ID of the model to use for the assistant"
  type        = string
}

variable "assistant_instructions" {
  description = "The system instructions that the assistant uses"
  type        = string
  default     = null
}

variable "assistant_description" {
  description = "The description of the assistant"
  type        = string
  default     = null
}

variable "assistant_tools" {
  description = "List of tools enabled on the assistant"
  type = list(object({
    type = string
    function = optional(object({
      name        = string
      description = optional(string)
      parameters  = string
    }))
  }))
  default = []
}

variable "assistant_file_ids" {
  description = "List of file IDs attached to the assistant"
  type        = list(string)
  default     = []
}

variable "assistant_metadata" {
  description = "Metadata for the assistant"
  type        = map(string)
  default     = {}
}

variable "enable_assistants_data_source" {
  description = "Whether to fetch the assistants list"
  type        = bool
  default     = false
}

variable "assistants_limit" {
  description = "Limit on the number of assistants to fetch (1-100)"
  type        = number
  default     = 20
}

variable "assistants_order" {
  description = "Sort order by created_at timestamp (asc or desc)"
  type        = string
  default     = "desc"
}

variable "assistants_after" {
  description = "Cursor for pagination (fetch after this assistant ID)"
  type        = string
  default     = null
}

variable "assistants_before" {
  description = "Cursor for pagination (fetch before this assistant ID)"
  type        = string
  default     = null
}

# Variables for the single assistant data source
variable "enable_single_assistant_data_source" {
  description = "Whether to fetch a single assistant by ID"
  type        = bool
  default     = false
}

variable "single_assistant_id" {
  description = "ID of a specific assistant to fetch"
  type        = string
  default     = null
} 