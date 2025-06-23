resource "openai_image_generation" "example" {
  prompt = "A serene landscape with mountains reflected in a crystal-clear lake at sunset"

  # Optional parameters
  model = "dall-e-3"  # or "dall-e-2"
  n     = 1           # Number of images to generate (1-10 for dall-e-2, only 1 for dall-e-3)
  size  = "1024x1024" # Options: 256x256, 512x512, 1024x1024 (dall-e-2), 1024x1792, 1792x1024 (dall-e-3)

  # Optional: Get response in base64 instead of URL
  # response_format = "b64_json"

  # Optional: Quality setting for DALL-E 3
  # quality = "hd"  # "standard" or "hd"

  # Optional: Style for DALL-E 3
  # style = "vivid"  # "vivid" or "natural"
}

output "image_url" {
  value = openai_image_generation.example.data[0].url
}
