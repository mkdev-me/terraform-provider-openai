# OpenAI Messages Module
# ==============================
# This module handles creation and management of messages through the OpenAI API.
# It provides support for creating messages with file attachments and metadata.

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
  }
}

# Input variables
variable "thread_id" {
  description = "The ID of the thread to add the message to"
  type        = string
}

variable "role" {
  description = "The role of the entity creating the message (currently only 'user' is supported)"
  type        = string
  default     = "user"

  validation {
    condition     = var.role == "user"
    error_message = "Currently only 'user' role is supported for message creation."
  }
}

variable "content" {
  description = "The content of the message"
  type        = string
  default     = null
}

variable "attachments" {
  description = "List of file attachments to include with the message"
  type = list(object({
    file_id = string
    tools = list(object({
      type = string
    }))
  }))
  default = []
}

variable "metadata" {
  description = "Set of key-value pairs that can be attached to the message"
  type        = map(string)
  default     = {}
}

variable "use_data_source" {
  description = "Whether to use the data source to retrieve an existing message or create a new one"
  type        = bool
  default     = false
}

variable "existing_message_id" {
  description = "ID of an existing message to retrieve (when use_data_source is true)"
  type        = string
  default     = null
}

# Local variables to simplify conditional logic
locals {
  use_resource    = !var.use_data_source
  use_data_source = var.use_data_source && var.existing_message_id != null

  # Validation to ensure proper parameters are provided
  validate_resource = (local.use_resource && var.content == null) ? tobool("When use_data_source is false, content is required") : true
  validate_data     = (local.use_data_source && var.existing_message_id == null) ? tobool("When use_data_source is true, existing_message_id is required") : true
}

# Create a new message in a thread
resource "openai_message" "this" {
  count = local.use_resource ? 1 : 0

  thread_id = var.thread_id
  role      = var.role
  content   = var.content

  dynamic "attachments" {
    for_each = var.attachments
    content {
      file_id = attachments.value.file_id

      dynamic "tools" {
        for_each = attachments.value.tools
        content {
          type = tools.value.type
        }
      }
    }
  }

  metadata = var.metadata
}

# Or retrieve an existing message
data "openai_message" "this" {
  count = local.use_data_source ? 1 : 0

  thread_id  = var.thread_id
  message_id = var.existing_message_id
}

# Format outputs consistently for both resource and data source
locals {
  data_source_output = local.use_data_source && length(data.openai_message.this) > 0 ? {
    id           = data.openai_message.this[0].id
    content      = data.openai_message.this[0].content
    role         = data.openai_message.this[0].role
    created_at   = data.openai_message.this[0].created_at
    metadata     = data.openai_message.this[0].metadata
    assistant_id = data.openai_message.this[0].assistant_id
    run_id       = data.openai_message.this[0].run_id
    attachments  = data.openai_message.this[0].attachments
  } : null

  resource_output = local.use_resource && length(openai_message.this) > 0 ? {
    id           = openai_message.this[0].id
    content      = openai_message.this[0].content
    role         = openai_message.this[0].role
    created_at   = openai_message.this[0].created_at
    metadata     = openai_message.this[0].metadata
    assistant_id = openai_message.this[0].assistant_id
    run_id       = openai_message.this[0].run_id
    attachments  = openai_message.this[0].attachments
  } : null

  output = local.use_data_source ? local.data_source_output : local.resource_output
}

# Outputs that work in both resource and data source mode
output "message_id" {
  description = "The ID of the message"
  value       = local.output != null ? local.output.id : null
}

output "created_at" {
  description = "The timestamp for when the message was created"
  value       = local.output != null ? local.output.created_at : null
}

output "role" {
  description = "The role of the entity that created the message"
  value       = local.output != null ? local.output.role : null
}

output "content" {
  description = "The content of the message"
  value       = local.output != null ? local.output.content : null
}

output "metadata" {
  description = "Set of key-value pairs attached to the message"
  value       = local.output != null ? local.output.metadata : null
}

output "assistant_id" {
  description = "If applicable, the ID of the assistant that authored this message"
  value       = local.output != null ? local.output.assistant_id : null
}

output "run_id" {
  description = "If applicable, the ID of the run associated with this message"
  value       = local.output != null ? local.output.run_id : null
}

output "attachments" {
  description = "A list of attachments in the message"
  value       = local.output != null ? local.output.attachments : null
} 