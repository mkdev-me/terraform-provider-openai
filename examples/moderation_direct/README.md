# OpenAI Moderation Direct Example

This example demonstrates how to use the OpenAI moderation resources directly (without using modules) to analyze text for potentially harmful content.

## Key Features Demonstrated

1. Direct use of the `openai_moderation` resource
2. Handling of model version differences between configuration and API response
3. Processing both safe and potentially harmful content
4. Accessing moderation results through outputs

## Important: Model Version Handling

This example demonstrates an important pattern when working with OpenAI's moderation API:

### The Issue

When you specify `text-moderation-latest` in your configuration, the OpenAI API returns a specific version (e.g., `text-moderation-007`) in the response. This difference causes Terraform to detect a change in the resource state and attempt to replace the resource on subsequent `terraform apply` operations.

### The Solution

To prevent unnecessary resource replacement, add a `lifecycle` block to your moderation resources:

```hcl
resource "openai_moderation" "example" {
  input = "Text to moderate"
  model = "text-moderation-latest"
  
  lifecycle {
    ignore_changes = [model]
  }
}
```

This tells Terraform to ignore differences in the model value between your configuration and the state, preventing unnecessary resource replacement.

## Usage

1. Initialize the Terraform working directory:

```bash
terraform init
```

2. Apply the configuration:

```bash
terraform apply
```

3. Check the outputs to see moderation results:

```bash
terraform output
```

## Import Information

If you need to import an existing moderation resource, you can use:

```bash
terraform import openai_moderation.example1 modr-123456789
```

When importing, be aware that:
1. The original input text is not available from the API, so a placeholder will be used
2. On the next `terraform apply`, the resource will be recreated to use your actual input text
3. You should add the `lifecycle` block to prevent subsequent recreations due to model version differences

## Clean Up

To remove all created resources:

```bash
terraform destroy
``` 