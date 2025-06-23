# Example: Editing images using OpenAI's image edit API
# This API allows you to edit portions of an image based on a text prompt

# Edit the image based on a prompt
resource "openai_image_edit" "add_sunglasses" {
  # The original image to edit (must be a PNG file, less than 4MB, and square)
  image = "original_photo.png"

  # The mask indicating which areas to edit (white = edit, black = keep)
  mask = "mask.png"

  # Describe what you want in the edited areas
  prompt = "Add stylish sunglasses"

  # Model to use - currently only "dall-e-2" supports image editing
  model = "dall-e-2"

  # Number of edited variations to generate (1-10)
  n = 1

  # Size of the generated images
  # Must be one of: 256x256, 512x512, or 1024x1024
  size = "1024x1024"

  # Response format: "url" or "b64_json"
  response_format = "url"
}

# Example: Removing an object from an image
resource "openai_image_edit" "remove_background" {
  image = "original_photo.png"
  mask  = "mask.png"

  prompt = "Plain white background"
  model  = "dall-e-2"
  size   = "512x512"

  # Generate multiple variations
  n = 3
}

# Example: Changing specific elements
resource "openai_image_edit" "change_clothing" {
  image = "original_photo.png"
  mask  = "mask.png"

  # Detailed prompt for better results
  prompt = "Business suit with blue tie, professional appearance, high quality"
  model  = "dall-e-2"
  size   = "1024x1024"

  # You can specify a user identifier for tracking
  user = "user-123"
}

# Output the edited image URL
output "edited_image_url" {
  value = openai_image_edit.add_sunglasses.data[0].url
}