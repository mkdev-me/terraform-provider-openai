variable "name" {
  description = "The name of the vector store."
  type        = string
}

variable "file_ids" {
  description = "A list of File IDs that the vector store should use."
  type        = list(string)
  default     = []
}

variable "metadata" {
  description = "Set of key-value pairs that can be attached to the vector store."
  type        = map(string)
  default     = {}
}

variable "chunking_strategy" {
  description = "The chunking strategy used to chunk the files."
  type = object({
    type = string
    # For 'fixed' strategy
    size = optional(number)
    # For 'semantic' strategy
    max_tokens = optional(number)
  })
  default = {
    type = "auto"
  }
}

variable "expires_after" {
  description = "The expiration policy for the vector store."
  type = object({
    days  = optional(number)
    never = optional(bool)
  })
  default = null
}

variable "use_file_batches" {
  description = "Whether to use file batches for adding files to the vector store."
  type        = bool
  default     = false
}

variable "file_attributes" {
  description = "Set of key-value pairs that can be attached to the file objects."
  type        = map(string)
  default     = {}
} 