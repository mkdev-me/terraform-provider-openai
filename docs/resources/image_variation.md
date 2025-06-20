# openai_image_variation

This resource allows you to create variations of existing images using OpenAI's DALL-E models. It provides a way to generate multiple alternatives based on a source image.

## Example Usage

```hcl
# Create variations of an existing image
resource "openai_image_variation" "cat_variations" {
  image     = "/path/to/cat.png"
  n         = 4
  model     = "dall-e-2"
  size      = "1024x1024"
}

# Access the variation image URLs
output "variation_image_urls" {
  value = [for img in openai_image_variation.cat_variations.data : img.url]
}

# Create a single variation with base64 output
resource "openai_image_variation" "logo_variation" {
  image           = "/path/to/logo.png"
  response_format = "b64_json"
}

# Access the base64-encoded image data
output "variation_base64" {
  value = openai_image_variation.logo_variation.data[0].b64_json
}
```

## Using the Image Module

For a more flexible approach, you can use the image module which provides a unified interface for all image operations:

```hcl
module "image_variations" {
  source    = "../../modules/image"
  
  operation = "variation"
  image     = "/path/to/cat.png"
  n         = 3
}

output "variation_urls" {
  value = module.image_variations.image_urls
}
```

## Argument Reference

* `image` - (Required) The image to use as the basis for the variation(s). Must be a valid PNG file, less than 4MB, and square.
* `model` - (Optional) The model to use for creating image variations. Currently, only "dall-e-2" is supported. Defaults to "dall-e-2".
* `n` - (Optional) The number of variation images to generate. Must be between 1 and 10. Defaults to 1.
* `size` - (Optional) The size of the generated images. Must be one of "256x256", "512x512", or "1024x1024". Defaults to "1024x1024".
* `response_format` - (Optional) The format in which the generated images are returned. Can be "url" or "b64_json". Defaults to "url".
* `user` - (Optional) A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `created` - The Unix timestamp for when the variation images were created.
* `data` - A list of variation images with the following attributes:
  * `url` - The URL of the variation image (if response_format is "url"). URLs are only valid for 60 minutes after image generation.
  * `b64_json` - The base64-encoded JSON of the variation image (if response_format is "b64_json").

## Source Image Requirements

- Must be a PNG file
- Must be less than 4MB in size
- Must be square in dimensions
- Should not contain any text, faces, or copyrighted content for best results
- Should have a clear subject and simple background for more predictable variations

## Supported Models

Currently, only DALL-E 2 is supported for image variations. DALL-E 3 does not support the variation operation at this time.

## Timeouts and Limitations

- The variation generation process may take up to 30 seconds.
- Image URLs are only valid for 60 minutes after generation.
- There are rate limits that vary by subscription and usage tier.
- Images are subject to safety filtering.
- The maximum number of variations per request is 10.

## How Variations Work

Image variations produce alternatives that maintain key aspects of the source image while introducing creative differences. The variations:

- Generally preserve the composition and layout of the original
- May modify colors, textures, and smaller details
- Maintain the general subject matter and style
- Do not allow specific textual prompting (unlike image generation or editing)

## Notes

- This resource is immutable. Any change to its parameters will result in the recreation of the resource.
- Variation images will not be automatically deleted from OpenAI's servers.
- Costs vary by size. For current pricing, see the [OpenAI pricing page](https://openai.com/pricing).
- For more information on image variations, see the [OpenAI documentation](https://platform.openai.com/docs/guides/images/variations). 