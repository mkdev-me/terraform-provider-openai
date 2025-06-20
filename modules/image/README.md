# OpenAI Image Module

This Terraform module provides a unified interface for working with OpenAI's DALL-E image generation capabilities, including image generation, editing, and creating variations.

## Features

- Supports all three OpenAI image operations:
  - **Image Generation**: Create images from textual descriptions
  - **Image Editing**: Edit existing images with a text prompt and optional mask
  - **Image Variation**: Create variations of existing images
- Configurable outputs in URL or base64 format
- Support for DALL-E 2 and DALL-E 3 models
- Advanced controls for image quality, style, and size

## Usage

### Image Generation

```hcl
module "cat_image" {
  source    = "../../modules/image"
  
  operation = "generation"
  prompt    = "A photorealistic cat wearing a space suit on the moon"
  model     = "dall-e-3"  # Optional, defaults to "dall-e-3" for generation
  size      = "1024x1024"
  quality   = "hd"        # Optional, only for dall-e-3
  style     = "vivid"     # Optional, only for dall-e-3
}

output "image_url" {
  value = module.cat_image.image_urls[0]
}
```

### Image Editing

```hcl
module "edited_image" {
  source    = "../../modules/image"
  
  operation = "edit"
  image     = "/path/to/image.png"
  mask      = "/path/to/mask.png"  # Optional
  prompt    = "Add a crown on the head"
  model     = "dall-e-2"  # Optional, defaults to "dall-e-2" for editing
  size      = "1024x1024"
}

output "edited_image_url" {
  value = module.edited_image.image_urls[0]
}
```

### Image Variation

```hcl
module "image_variations" {
  source    = "../../modules/image"
  
  operation = "variation"
  image     = "/path/to/image.png"
  n         = 3  # Generate 3 variations
  model     = "dall-e-2"  # Optional, defaults to "dall-e-2" for variations
  size      = "1024x1024"
}

output "variation_urls" {
  value = module.image_variations.image_urls
}
```

## Input Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| operation | The image operation to perform: 'generation', 'edit', or 'variation' | `string` | n/a | yes |
| model | The model to use for image operations (dall-e-2 or dall-e-3) | `string` | Operation-dependent | no |
| prompt | Text description of the desired image(s) | `string` | `""` | yes for generation and edit |
| image | The image to edit or use as the basis for variation | `string` | `""` | yes for edit and variation |
| mask | An additional image whose transparent areas indicate where to edit | `string` | `""` | no |
| n | Number of images to generate (1-10, for dall-e-3 only 1 is supported) | `number` | `1` | no |
| size | Size of the generated images | `string` | `"1024x1024"` | no |
| quality | Quality of the generated images (standard or hd) | `string` | `"standard"` | no |
| style | Style of the generated images (vivid or natural) | `string` | `"vivid"` | no |
| response_format | Format of the response (url or b64_json) | `string` | `"url"` | no |
| user | A unique identifier representing your end-user | `string` | `""` | no |

## Output Values

| Name | Description |
|------|-------------|
| created | The timestamp when the image(s) were created |
| images | Complete information about all generated images |
| image_urls | List of URLs of the generated images (if response_format is 'url') |
| operation | The image operation that was performed |

## Size Restrictions

### DALL-E 2
- Supported sizes: `256x256`, `512x512`, `1024x1024`

### DALL-E 3
- Supported sizes: `1024x1024`, `1792x1024`, `1024x1792`
- Only supports `n=1` (single image generation)
- Offers additional `quality` and `style` parameters

## Notes

- Image files must be in PNG format, less than 4MB, and square.
- For image edits, if a mask is not provided, the image must have transparency which will be used as the mask.
- Image URLs are only valid for 60 minutes after generation.

## Importing Existing Resources

To import existing OpenAI image resources, you must specify all original parameters used during creation, as OpenAI doesn't provide an API to retrieve these parameters.

### Import Format

The general format for importing image resources is:
```
terraform import RESOURCE_TYPE.RESOURCE_NAME "ID,param1=value1,param2=value2,..."
```

Examples:

#### Image Generation
```bash
terraform import openai_image_generation.example \
"img-1743882006,prompt=A serene landscape with mountains and a lake at sunset,\
model=dall-e-3,n=1,quality=standard,response_format=url,\
size=1024x1024,style=vivid"
```

#### Image Edit
```bash
terraform import openai_image_edit.example \
"img-edit-1743882006,prompt=Add a hot air balloon in the sky,image=./sample_image.png,mask=./sample_mask.png,model=dall-e-2,n=1,response_format=url,size=1024x1024"
```

#### Image Variation
```bash
terraform import openai_image_variation.example \
"img-var-1743882006,image=./sample_image.png,model=dall-e-2,n=1,response_format=url,size=1024x1024"
```

### Import Process

1. Define the resource in your Terraform configuration with the same parameters used during creation
2. Add a lifecycle block with `ignore_changes` to prevent recreation
3. Run the import command with all original parameters
4. Verify the import with `terraform state show`

## Troubleshooting

### Common Issues

#### 1. Invalid API URL

Make sure the provider is configured with the correct API URL path:

```hcl
provider "openai" {
  api_url = "https://api.openai.com/v1"
}
```

The `/v1` part is essential for proper endpoint construction.

#### 2. Image File Problems

For editing and variation operations:
- Verify that image files actually exist and are not empty
- Ensure images are valid PNG files
- Images must be square with the same width and height
- Keep images under 4MB in size

#### 3. Download Script Issues

When downloading images, URLs from OpenAI contain special characters that can break bash scripts. Use this template instead:

```hcl
provisioner "local-exec" {
  command = <<-EOT
    mkdir -p ./output
    echo "${join("\n", module.image_operation.image_urls)}" > /tmp/image_urls.txt
    while IFS= read -r url; do
      echo "Downloading image"
      curl -s "$url" -o "./output/image_$RANDOM.png"
    done < /tmp/image_urls.txt
  EOT
}
```

#### 4. Rate Limiting

If you hit rate limits:
- Focus on one operation at a time
- Reduce the number of concurrent operations
- Implement delays between operations

#### 5. Account/Billing Issues

For "account not active" errors:
- Verify your OpenAI account has active billing
- Ensure your API key is correct and active
- Check your organization's access to DALL-E features

## Related Resources

- [OpenAI Image API Documentation](https://platform.openai.com/docs/api-reference/images)
- [DALL-E 3 API Documentation](https://platform.openai.com/docs/guides/images/dall-e-3)
- [DALL-E 2 API Documentation](https://platform.openai.com/docs/guides/images/dall-e-2) 