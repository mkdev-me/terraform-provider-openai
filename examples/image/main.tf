# OpenAI Image Resources Examples
# This example demonstrates how to use OpenAI's image-related resources:
# - Image Generation: Creates images from text prompts
# - Image Edit: Edits images based on a mask and instructions
# - Image Variation: Creates variations of an existing image

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
  }
}

# Configure the OpenAI Provider
provider "openai" {
  # API key is taken from OPENAI_API_KEY environment variable
  # For testing: uncomment and add your keys
  # api_key   = "sk-proj-..."
  # admin_key = "sk-admin-..."
}

# Local variables
locals {
  output_dir = "${path.module}/output"
}

# 1. Image Generation Example
resource "openai_image_generation" "example" {
  prompt  = "A serene landscape with mountains and a lake at sunset"
  model   = "dall-e-3"
  size    = "1024x1024"
  quality = "standard"
  style   = "vivid"
  n       = 1

  # Optional: Adjust response format if you want base64 encoded images instead of URLs
  response_format = "url"

  # Optional: Add a unique identifier for your end user
  # user = "user-123"
}

# 2. Image Edit Example
resource "openai_image_edit" "example" {
  # Path to the source image (must be a PNG file, less than 4MB, and square)
  image = "${path.module}/sample_image.png"

  # Path to the mask image (transparent areas indicate where to edit)
  mask = "${path.module}/sample_mask.png"

  # Text prompt describing the desired edit
  prompt = "Add a hot air balloon in the sky"

  # Currently only dall-e-2 is supported for image edits
  model = "dall-e-2"

  # Number of edited images to generate
  n = 1

  # Size of the edited image
  size = "1024x1024"

  # Format of the response
  response_format = "url"
}

# 3. Image Variation Example
resource "openai_image_variation" "example" {
  # Path to the source image (must be a PNG file, less than 4MB, and square)
  image = "${path.module}/sample_image.png"

  # Currently only dall-e-2 is supported for image variations
  model = "dall-e-2"

  # Number of variations to generate
  n = 1

  # Size of the variations
  size = "1024x1024"

  # Format of the response
  response_format = "url"
}

# Outputs for Image Generation
output "generated_image_urls" {
  description = "The URLs of the generated images"
  value       = try(openai_image_generation.example.data[*].url, [])
}

output "generated_image_revised_prompts" {
  description = "The revised prompts used for generation (if any)"
  value       = try(openai_image_generation.example.data[*].revised_prompt, [])
}

# Outputs for Image Edit
output "edited_image_urls" {
  description = "The URLs of the edited images"
  value       = try(openai_image_edit.example.data[*].url, [])
}

# Outputs for Image Variation
output "variation_image_urls" {
  description = "The URLs of the image variations"
  value       = try(openai_image_variation.example.data[*].url, [])
} 