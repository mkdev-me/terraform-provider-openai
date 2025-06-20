# openai_image_generation

This resource allows you to generate images from text descriptions using OpenAI's DALL-E models. It provides comprehensive control over the generation process, including model selection, image quality, style, and output format.

## Example Usage

```hcl
# Generate a single high-quality image with DALL-E 3
resource "openai_image_generation" "cat_astronaut" {
  prompt    = "A photorealistic image of a cat in an astronaut suit on the moon"
  model     = "dall-e-3"
  size      = "1024x1024"
  quality   = "hd"
  style     = "vivid"
}

# Access the generated image URL
output "image_url" {
  value = openai_image_generation.cat_astronaut.data[0].url
}

# Generate multiple images with DALL-E 2
resource "openai_image_generation" "fantasy_scene" {
  prompt    = "A magical forest with glowing mushrooms and fairy lights"
  model     = "dall-e-2"
  n         = 3
  size      = "1024x1024"
}

# Access all generated image URLs
output "image_urls" {
  value = [for img in openai_image_generation.fantasy_scene.data : img.url]
}
```

### Provider Configuration

When using image operations, ensure the provider is configured with the correct API URL:

```hcl
provider "openai" {
  # API key sourced from environment variable: OPENAI_API_KEY
  api_url = "https://api.openai.com/v1"
}
```

The `/v1` part of the API URL is required for proper endpoint construction.

## Using the Image Module

For a more flexible approach, you can use the image module which provides a unified interface for all image operations:

```hcl
module "generated_image" {
  source    = "../../modules/image"
  
  operation = "generation"
  prompt    = "A photorealistic cat wearing a space suit on the moon"
  model     = "dall-e-3"
  quality   = "hd"
  style     = "vivid"
}

output "image_url" {
  value = module.generated_image.image_urls[0]
}
```

## Argument Reference

* `prompt` - (Required) A text description of the desired image(s). The maximum length is 1000 characters for dall-e-2 and 4000 characters for dall-e-3.
* `model` - (Optional) The model to use for image generation. Available values are "dall-e-2" and "dall-e-3". Defaults to "dall-e-3".
* `n` - (Optional) The number of images to generate. Must be between 1 and 10. For dall-e-3, only n=1 is supported. Defaults to 1.
* `quality` - (Optional) The quality of the image that will be generated. Can be "standard" or "hd". HD creates images with finer details and greater consistency. This parameter is only supported for dall-e-3. Defaults to "standard".
* `size` - (Optional) The size of the generated images. For dall-e-2, must be one of "256x256", "512x512", or "1024x1024". For dall-e-3, must be one of "1024x1024", "1792x1024", or "1024x1792". Defaults to "1024x1024".
* `style` - (Optional) The style of the generated images. Can be "vivid" or "natural". Vivid causes the model to generate hyper-real and dramatic images, while natural produces more natural, less hyper-real looking images. This parameter is only supported for dall-e-3. Defaults to "vivid".
* `response_format` - (Optional) The format in which the generated images are returned. Can be "url" or "b64_json". Defaults to "url".
* `user` - (Optional) A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `created` - The Unix timestamp for when the images were created.
* `data` - A list of generated images with the following attributes:
  * `url` - The URL of the generated image (if response_format is "url"). URLs are only valid for 60 minutes after image generation.
  * `b64_json` - The base64-encoded JSON of the generated image (if response_format is "b64_json").
  * `revised_prompt` - The prompt that was used to generate the image, potentially modified from the original prompt. This field is only populated when using DALL-E 3.

## Model-Specific Features

### DALL-E 3
- Only supports generating a single image at a time (`n=1`)
- Supports high definition (HD) quality images
- Offers style control (vivid or natural)
- Provides revised prompts that may enhance the original prompt
- Supports sizes: 1024x1024, 1792x1024, 1024x1792
- Maximum prompt length: 4000 characters

### DALL-E 2
- Supports generating multiple images in a single request (up to 10)
- Supports sizes: 256x256, 512x512, 1024x1024
- Maximum prompt length: 1000 characters

## Timeouts and Limitations

- The generation process may take up to 30 seconds, especially for high-quality images.
- Image URLs are only valid for 60 minutes after generation.
- There are rate limits that vary by model, subscription, and usage tier.
- Prompts are subject to safety filtering.

## Notes

- This resource is immutable. Any change to its parameters will result in the recreation of the resource.
- Generated images will be not be automatically deleted from OpenAI's servers.
- Costs vary by model, quality, and size. For current pricing, see the [OpenAI pricing page](https://openai.com/pricing).
- For more information on DALL-E models, see the [OpenAI documentation](https://platform.openai.com/docs/guides/images). 