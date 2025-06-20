# openai_image_edit

This resource allows you to edit existing images using OpenAI's DALL-E models. It supports various editing options including text prompts, masking, and provides control over the output format.

## Example Usage

```hcl
# Edit an image by adding a crown to it
resource "openai_image_edit" "crowned_image" {
  image     = "/path/to/person.png"
  mask      = "/path/to/head_mask.png"  # Optional
  prompt    = "Add a golden crown with jewels"
  model     = "dall-e-2"
  size      = "1024x1024"
}

# Access the edited image URL
output "edited_image_url" {
  value = openai_image_edit.crowned_image.data[0].url
}

# Edit an image and create multiple variations
resource "openai_image_edit" "landscape_edit" {
  image           = "/path/to/landscape.png"
  prompt          = "Add a castle on the mountain"
  n               = 3
  response_format = "url"
}

# Access all edited image URLs
output "edited_image_urls" {
  value = [for img in openai_image_edit.landscape_edit.data : img.url]
}
```

## Using the Image Module

For a more flexible approach, you can use the image module which provides a unified interface for all image operations:

```hcl
module "edited_image" {
  source    = "../../modules/image"
  
  operation = "edit"
  image     = "/path/to/person.png"
  mask      = "/path/to/head_mask.png"  # Optional
  prompt    = "Add a golden crown with jewels"
}

output "edited_image_url" {
  value = module.edited_image.image_urls[0]
}
```

## Argument Reference

* `image` - (Required) The image to edit. Must be a valid PNG file, less than 4MB, and square. If mask is not provided, image must have transparency, which will be used as the mask.
* `prompt` - (Required) A text description of the desired edit. The maximum length is 1000 characters.
* `mask` - (Optional) An additional image whose fully transparent areas (e.g., where alpha is zero) indicate where the image should be edited. Must be a valid PNG file, less than 4MB, and have the same dimensions as the image.
* `model` - (Optional) The model to use for image editing. Currently, only "dall-e-2" is supported. Defaults to "dall-e-2".
* `n` - (Optional) The number of edited images to generate. Must be between 1 and 10. Defaults to 1.
* `size` - (Optional) The size of the generated images. Must be one of "256x256", "512x512", or "1024x1024". Defaults to "1024x1024".
* `response_format` - (Optional) The format in which the edited images are returned. Can be "url" or "b64_json". Defaults to "url".
* `user` - (Optional) A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `created` - The Unix timestamp for when the edited images were created.
* `data` - A list of edited images with the following attributes:
  * `url` - The URL of the edited image (if response_format is "url"). URLs are only valid for 60 minutes after image generation.
  * `b64_json` - The base64-encoded JSON of the edited image (if response_format is "b64_json").

## Image Requirements

### Source Image Requirements
- Must be a PNG file
- Must be less than 4MB in size
- Must be square in dimensions
- If no mask is provided, the image must have transparency which will be used as the mask

### Mask Image Requirements
- Must be a PNG file
- Must be less than 4MB in size
- Must have the same dimensions as the source image
- Fully transparent areas (alpha = 0) indicate where the source image should be edited

## Supported Models

Currently, only DALL-E 2 is supported for image editing. DALL-E 3 does not support the edit operation at this time.

## Timeouts and Limitations

- The editing process may take up to 30 seconds.
- Image URLs are only valid for 60 minutes after generation.
- There are rate limits that vary by subscription and usage tier.
- Prompts are subject to safety filtering.
- The maximum prompt length is 1000 characters.

## Notes

- This resource is immutable. Any change to its parameters will result in the recreation of the resource.
- Edited images will not be automatically deleted from OpenAI's servers.
- Costs vary by size. For current pricing, see the [OpenAI pricing page](https://openai.com/pricing).
- For more information on image editing, see the [OpenAI documentation](https://platform.openai.com/docs/guides/images/editing-images). 