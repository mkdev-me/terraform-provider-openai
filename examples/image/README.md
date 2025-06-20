# OpenAI Image Resources Example

This example demonstrates how to use the OpenAI Terraform Provider to interact with image-related resources:

1. **Image Generation** - Create new images from text descriptions
2. **Image Edit** - Modify existing images based on a mask and text instructions
3. **Image Variation** - Generate variations of an existing image

## Prerequisites

- An OpenAI API key
- Terraform installed
- The OpenAI Terraform Provider installed

## File Structure

- `main.tf` - The main Terraform configuration file with examples
- `sample_image.png` - A sample image for editing and variation examples
- `sample_mask.png` - A sample mask image for editing examples
- `output/` - Directory where image details are stored

## Setup

Before running this example, you need to:

1. Ensure you have a valid OpenAI API key with access to DALL-E models
2. Place a valid `sample_image.png` file in this directory for the editing and variation examples
3. Place a valid `sample_mask.png` file in this directory for the editing example

### Sample Image Requirements

- Must be a PNG file
- Must be less than 4MB in size
- Must be square in dimensions
- For editing without a mask, the image must have transparency

### Sample Mask Requirements

- Must be a PNG file
- Must be the same dimensions as the sample image
- Fully transparent areas (alpha = 0) indicate where the image should be edited

## Usage

1. Set your OpenAI API key:
   ```
   export OPENAI_API_KEY=your_api_key_here
   ```

2. Initialize Terraform:
   ```
   terraform init
   ```

3. Apply the Terraform configuration:
   ```
   terraform apply
   ```

4. After successful execution, view the generated image URLs:
   ```
   terraform output generated_image_urls
   terraform output edited_image_urls
   terraform output variation_image_urls
   ```

## Resource Examples

The example includes three main resources:

### 1. Image Generation (`openai_image_generation`)

```hcl
resource "openai_image_generation" "example" {
  prompt  = "A serene landscape with mountains and a lake at sunset"
  model   = "dall-e-3"
  size    = "1024x1024"
  quality = "standard"
  style   = "vivid"
  n       = 1
  response_format = "url"
}
```

This resource creates an image based on the provided text prompt.

### 2. Image Edit (`openai_image_edit`)

```hcl
resource "openai_image_edit" "example" {
  image = "${path.module}/sample_image.png"
  mask  = "${path.module}/sample_mask.png"
  prompt = "Add a hot air balloon in the sky"
  model = "dall-e-2"
  n = 1
  size = "1024x1024"
  response_format = "url"
}
```

This resource edits an existing image according to the provided mask and prompt.

### 3. Image Variation (`openai_image_variation`)

```hcl
resource "openai_image_variation" "example" {
  image = "${path.module}/sample_image.png"
  model = "dall-e-2"
  n = 1
  size = "1024x1024"
  response_format = "url"
}
```

This resource creates variations of an existing image.

## Costs

Note that using the DALL-E API incurs costs. As of this writing:
- DALL-E 3:
  - Standard quality: $0.040 / image (1024×1024)
  - HD quality: $0.080 / image (1024×1024)
- DALL-E 2: $0.020 / image (1024×1024)

Prices may vary by image size and over time. Check the [OpenAI pricing page](https://openai.com/pricing) for the most up-to-date information.

## Clean Up

To destroy the resources created by Terraform:
```
terraform destroy
```

Note that this will not delete any images that OpenAI has stored.

## DALL-E Model Compatibility

Note the following compatibility constraints:

- **Image Edit**: Currently only supports the `dall-e-2` model
- **Image Variation**: Currently only supports the `dall-e-2` model
- **Image Generation**: Supports both `dall-e-2` and `dall-e-3` models

## Troubleshooting

If you encounter issues with the image operations, check the following:

1. Ensure your API key has access to DALL-E models
2. Verify that your sample images meet the requirements (PNG format, square dimensions, under 4MB)
3. For editing operations, check that your mask properly indicates the areas to be edited
4. If you encounter rate limits, reduce the number of concurrent operations 

## Importing Existing Resources

This example also demonstrates different approaches to managing existing OpenAI resources with Terraform:

1. **Standard Import Approach**: The `import.tf` file shows the traditional import approach, which has limitations with OpenAI resources due to the ForceNew attributes.

2. **terraform_data Approach**: The `external_data.tf` file demonstrates a better approach for OpenAI resources. Instead of trying to manage the resources directly (which would trigger replacement), it uses `terraform_data` to store metadata about existing resources in your Terraform state.

### Image Resource Import Commands

OpenAI does not provide an API to retrieve the original parameters of image generations. Therefore, to import existing image resources, you must specify all the original parameters in the import command.

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

The import process:

1. Requires specifying the full ID in the format: `[resource-id],[param1=value1],[param2=value2],...`
2. All parameters used during creation must be specified exactly as they were used originally
3. The resource configuration in Terraform must match these parameters
4. A lifecycle block with `ignore_changes` should be used to prevent recreation when using the standard import approach

For detailed documentation on these approaches, see:
- `IMPORT_README.md` - Overview of approaches to handling OpenAI resources
- `TERRAFORM_DATA_APPROACH.md` - Detailed guide to using terraform_data

The terraform_data approach is recommended for OpenAI resources because:
- OpenAI doesn't provide APIs to retrieve original parameters
- Most attributes are ForceNew, causing replacement on import
- Many resources are one-time operations that can't be modified

Example:
```hcl
resource "terraform_data" "image_generation" {
  input = {
    id      = "img-1743428308"
    prompt  = "A serene landscape with mountains and a lake at sunset"
    model   = "dall-e-3"
    # Other attributes...
  }
}
``` 