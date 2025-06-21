# OpenAI Image Module
# ==============================
# This module manages various image operations with OpenAI's DALL-E models,
# including image generation, editing, and creating variations.

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.2.1"
    }
    local = {
      source  = "hashicorp/local"
      version = "2.4.0"
    }
  }
}

# Common variables
variable "operation" {
  description = "The image operation to perform: 'generation', 'edit', or 'variation'"
  type        = string

  validation {
    condition     = contains(["generation", "edit", "variation"], var.operation)
    error_message = "The operation must be one of: 'generation', 'edit', or 'variation'."
  }
}

variable "model" {
  description = "The model to use for image operations. Note that dall-e-3 is only available for generation operations."
  type        = string
  default     = ""
}

# Variables for all operations
variable "n" {
  description = "The number of images to generate. Must be between 1 and 10. For dall-e-3, only n=1 is supported."
  type        = number
  default     = 1

  validation {
    condition     = var.n >= 1 && var.n <= 10
    error_message = "The number of images must be between 1 and 10."
  }
}

variable "size" {
  description = "The size of the generated images. Available sizes vary by model."
  type        = string
  default     = "1024x1024"
}

variable "response_format" {
  description = "The format in which the generated images are returned: 'url' or 'b64_json'."
  type        = string
  default     = "url"

  validation {
    condition     = contains(["url", "b64_json"], var.response_format)
    error_message = "The response_format must be either 'url' or 'b64_json'."
  }
}

variable "user" {
  description = "A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse."
  type        = string
  default     = ""
}

# Variables for image generation
variable "prompt" {
  description = "A text description of the desired image(s). Required for generation and edit operations."
  type        = string
  default     = ""
}

variable "quality" {
  description = "The quality of the image that will be generated. Available for dall-e-3 only."
  type        = string
  default     = "standard"

  validation {
    condition     = contains(["standard", "hd"], var.quality)
    error_message = "The quality must be either 'standard' or 'hd'."
  }
}

variable "style" {
  description = "The style of the generated images. Available for dall-e-3 only."
  type        = string
  default     = "vivid"

  validation {
    condition     = contains(["vivid", "natural"], var.style)
    error_message = "The style must be either 'vivid' or 'natural'."
  }
}

# Variables for image edit and variation
variable "image" {
  description = "The image to edit or use as the basis for variation. Must be a valid PNG file, less than 4MB, and square."
  type        = string
  default     = ""
}

# Variable for image edit only
variable "mask" {
  description = "An additional image whose fully transparent areas indicate where the image should be edited."
  type        = string
  default     = ""
}

# Determine default model based on operation
locals {
  default_model = var.operation == "generation" ? "dall-e-3" : "dall-e-2"
  actual_model  = var.model != "" ? var.model : local.default_model
}

# Image Generation
resource "openai_image_generation" "this" {
  count = var.operation == "generation" ? 1 : 0

  model           = local.actual_model
  prompt          = var.prompt
  n               = var.n
  size            = var.size
  response_format = var.response_format
  quality         = var.quality
  style           = var.style
  user            = var.user
}

# Image Edit
resource "openai_image_edit" "this" {
  count = var.operation == "edit" ? 1 : 0

  model           = local.actual_model
  image           = var.image
  mask            = var.mask
  prompt          = var.prompt
  n               = var.n
  size            = var.size
  response_format = var.response_format
  user            = var.user
}

# Image Variation
resource "openai_image_variation" "this" {
  count = var.operation == "variation" ? 1 : 0

  model           = local.actual_model
  image           = var.image
  n               = var.n
  size            = var.size
  response_format = var.response_format
  user            = var.user
}

# Outputs
output "created" {
  description = "The timestamp when the image(s) were created"
  value = (
    var.operation == "generation" ? (length(openai_image_generation.this) > 0 ? openai_image_generation.this[0].created : null) :
    var.operation == "edit" ? (length(openai_image_edit.this) > 0 ? openai_image_edit.this[0].created : null) :
    var.operation == "variation" ? (length(openai_image_variation.this) > 0 ? openai_image_variation.this[0].created : null) :
    null
  )
}

output "images" {
  description = "Information about the generated, edited, or varied images"
  value = (
    var.operation == "generation" ? (length(openai_image_generation.this) > 0 ? openai_image_generation.this[0].data : []) :
    var.operation == "edit" ? (length(openai_image_edit.this) > 0 ? openai_image_edit.this[0].data : []) :
    var.operation == "variation" ? (length(openai_image_variation.this) > 0 ? openai_image_variation.this[0].data : []) :
    []
  )
}

output "image_urls" {
  description = "List of URLs of the generated images (if response_format is 'url')"
  value = (
    var.response_format == "url" ? (
      var.operation == "generation" ? (length(openai_image_generation.this) > 0 ? [for img in openai_image_generation.this[0].data : img.url if img.url != ""] : []) :
      var.operation == "edit" ? (length(openai_image_edit.this) > 0 ? [for img in openai_image_edit.this[0].data : img.url if img.url != ""] : []) :
      var.operation == "variation" ? (length(openai_image_variation.this) > 0 ? [for img in openai_image_variation.this[0].data : img.url if img.url != ""] : []) :
      []
    ) : []
  )
}

output "operation" {
  description = "The image operation that was performed"
  value       = var.operation
} 