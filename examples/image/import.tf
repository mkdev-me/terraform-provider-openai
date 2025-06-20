# Configuration for importing existing OpenAI resources
# IMPORTANT: You must specify EXACTLY the same values used in the original creation

variable "image_generation_id" {
  description = "ID of an existing image generation to import"
  type        = string
  default     = "img-1743428308"
}

variable "image_edit_id" {
  description = "ID of an existing image edit to import"
  type        = string
  default     = "img-edit-1743428308"
}

variable "image_variation_id" {
  description = "ID of an existing image variation to import"
  type        = string
  default     = "img-var-1743428307"
}

# Complete resource definitions with all required attributes
# You must set these attributes to match the original values
resource "openai_image_generation" "imported" {
  prompt          = "A serene landscape with mountains and a lake at sunset"
  model           = "dall-e-3"
  n               = 1
  quality         = "standard"
  response_format = "url"
  size            = "1024x1024"
  style           = "vivid"

  # This lifecycle block is crucial to prevent Terraform from trying to recreate the resource
  # We tell Terraform to ignore changes to all ForceNew attributes
  lifecycle {
    ignore_changes = [
      prompt,
      model,
      n,
      quality,
      response_format,
      size,
      style
    ]
  }
}

resource "openai_image_edit" "imported" {
  prompt          = "Add a hot air balloon in the sky"
  image           = "./sample_image.png"
  mask            = "./sample_mask.png"
  model           = "dall-e-2"
  n               = 1
  response_format = "url"
  size            = "1024x1024"

  # This lifecycle block is crucial to prevent Terraform from trying to recreate the resource
  lifecycle {
    ignore_changes = [
      prompt,
      image,
      mask,
      model,
      n,
      response_format,
      size
    ]
  }
}

resource "openai_image_variation" "imported" {
  image           = "./sample_image.png"
  model           = "dall-e-2"
  n               = 1
  response_format = "url"
  size            = "1024x1024"

  # This lifecycle block is crucial to prevent Terraform from trying to recreate the resource
  lifecycle {
    ignore_changes = [
      image,
      model,
      n,
      response_format,
      size
    ]
  }
}

# Run these commands to import:
# terraform init
# terraform import openai_image_generation.imported var.image_generation_id
# terraform import openai_image_edit.imported var.image_edit_id
# terraform import openai_image_variation.imported var.image_variation_id 

# IMPORTANT NOTE: This approach uses lifecycle blocks to prevent Terraform from
# attempting to replace imported resources. The values specified in the configuration
# are necessary for Terraform to import correctly, but they won't cause
# changes to existing resources thanks to ignore_changes. 