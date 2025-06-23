# Example: Creating variations of existing images
# The Image Variation API creates new images that are variations of an input image

# First, prepare the source image file (ensure it exists)

# Create variations of the image
resource "openai_image_variation" "logo_variations" {
  # The source image to create variations from (must be a PNG file, less than 4MB, and square)
  image = "logo.png"

  # Model to use - currently only "dall-e-2" supports variations
  model = "dall-e-2"

  # Number of variations to generate (1-10)
  n = 4

  # Size of the generated variations
  # Must be one of: 256x256, 512x512, or 1024x1024
  # Note: All variations will be square, regardless of input dimensions
  size = "512x512"

  # Response format: "url" or "b64_json"
  response_format = "url"
}

# Example: Creating a single high-res variation
resource "openai_image_variation" "hero_image_variant" {
  image = "logo.png"
  model = "dall-e-2"
  n     = 1
  size  = "1024x1024"

  # Optional: User identifier for tracking
  user = "marketing-team"
}

# Example: Multiple variations for A/B testing
resource "openai_image_variation" "product_variants" {
  image           = "product_photo.jpg"
  model           = "dall-e-2"
  n               = 10 # Maximum variations
  size            = "512x512"
  response_format = "url"
}

# Example: Getting base64 encoded variations
resource "openai_image_variation" "base64_variants" {
  image           = "logo.png"
  model           = "dall-e-2"
  n               = 2
  size            = "256x256"
  response_format = "b64_json" # Returns base64 encoded images
}

# Output the first variation URL
output "first_variation_url" {
  value = openai_image_variation.logo_variations.data[0].url
}
