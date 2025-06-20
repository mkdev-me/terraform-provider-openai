# OpenAI Moderation Module

This Terraform module provides access to OpenAI's content moderation API, which helps detect harmful content in text submissions.

~> **NOTE:** OpenAI does not provide an API to retrieve moderation results by ID. Therefore, a data source for moderation is not available. Moderation can only be used as a resource where you provide the input text and get the analysis results.

## Features

- Check if content violates OpenAI's usage policies
- Detailed category scores and flagged categories
- Optional model selection
- Easy integration with Terraform workflows
- Lifecycle management to prevent unnecessary resource replacement

## Usage

### Single Text Moderation
```hcl
module "text_moderation" {
  source = "path/to/modules/moderation"
  
  input = "Text to moderate"
  
  # Optional: specify moderation model
  model = "text-moderation-latest"
}
```

### Direct Resource Usage (Alternative)

You can also use the resource directly without the module:

```hcl
resource "openai_moderation" "example" {
  input = "Text to moderate"
  model = "text-moderation-latest"
}
```

Using the direct resource approach will make the resources appear directly in the Terraform state list without module prefixes.

### Handling Multiple Texts (Recommended Approach)
For multiple texts, create separate module instances or resources:

```hcl
module "text1_moderation" {
  source = "path/to/modules/moderation"
  input  = "First text to moderate"
  model  = "text-moderation-latest"
}

module "text2_moderation" {
  source = "path/to/modules/moderation"
  input  = "Second text to moderate"
  model  = "text-moderation-latest"
}

# Or with direct resources:
resource "openai_moderation" "first_text" {
  input = "First text to moderate"
  model = "text-moderation-latest"
}

resource "openai_moderation" "second_text" {
  input = "Second text to moderate"
  model = "text-moderation-latest"
}
```

This approach ensures proper state handling in Terraform and prevents inconsistencies.

## Model Version Differences

When you specify `text-moderation-latest` in your configuration, the OpenAI API returns a specific version (e.g., `text-moderation-007`) in the response. 

To prevent unnecessary resource replacement, this module uses two complementary approaches:

1. **Provider-level handling**: The provider automatically handles model version differences using a CustomizeDiff function that prevents unnecessary resource replacement.

2. **Resource-level lifecycle block**: This module also includes a `lifecycle` block that explicitly tells Terraform to ignore model changes:

```hcl
lifecycle {
  ignore_changes = [model]
}
```

This dual approach ensures maximum stability and prevents Terraform from trying to recreate the resource when only the model version differs.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| `input` | Input to classify. Should be a single string. For multiple texts, use separate module instances. | `any` | n/a | yes |
| `model` | The content moderation model to use. | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| `id` | The ID of the moderation response |
| `model` | The model used for moderation |
| `results` | The moderation results including categories and category scores |
| `flagged` | Whether the input was flagged by the moderation model |
| `categories` | The content categories that were flagged (if any) |
| `category_scores` | The scores for each content category |

## Import Support

Moderation resources can be imported using the moderation ID:

```bash
# With module:
terraform import module.text_moderation.openai_moderation.this modr-12345abcdef

# With direct resource:
terraform import openai_moderation.example modr-12345abcdef
```

~> **IMPORTANT:** OpenAI does not provide an API endpoint to retrieve moderation results by ID. When importing a moderation resource, the original input text and detailed analysis results cannot be retrieved from the API. Instead, placeholder values will be used:
- The input field will show `[Imported: Original input not available]`
- The moderation categories, scores, and results fields will be empty or contain default values
- The `_api_response` will contain minimal information

This means that after importing, the resource's state will not match its original creation state. If you run `terraform plan` after importing, Terraform will detect differences and plan to update the resource on the next `terraform apply`, which will recreate the moderation with the input text specified in your configuration.

Example of an imported resource state:
```hcl
resource "openai_moderation" "example" {
  _api_response   = jsonencode({
    id      = "modr-12345abcdef"
    model   = "text-moderation-latest"
    results = [{
      categories      = {}
      category_scores = {}
      flagged         = false
    }]
  })
  categories      = {}
  category_scores = {}
  flagged         = false
  id              = "modr-12345abcdef"
  input           = "[Imported: Original input not available]"
  results         = [{
    categories                   = {}
    category_applied_input_types = {}
    category_scores              = {}
    flagged                      = false
  }]
}
```

## Common API Errors and Solutions

### Billing Not Active

**Error:** `OpenAI billing not active: Your account is not active, please check your billing details on our website.`

**Solution:**
1. Visit [OpenAI Billing](https://platform.openai.com/account/billing)
2. Add a valid payment method
3. Ensure you have sufficient credits or payment method

### Rate Limiting

**Error:** `OpenAI rate limit exceeded: Too Many Requests`

**Solution:**
1. Implement retry logic with exponential backoff in your application code
2. Reduce the frequency of requests
3. For production systems, use a queueing system to manage API call rates

### Production Recommendations

For production use of content moderation:

1. **Error Handling**: Implement robust error handling for all API errors
2. **Retry Logic**: Add exponential backoff for rate limit errors
3. **Fallbacks**: Consider having a fallback moderation system
4. **Caching**: Cache moderation results for identical inputs to reduce API calls
5. **Cost Management**: Monitor API usage to prevent unexpected billing

## Important Notes

- **State Consistency**: For multiple input texts, create separate module instances instead of using array inputs to ensure proper Terraform state handling.
- **Immutability**: Moderation resources are immutable - running `terraform apply` multiple times will recreate the resource if there are changes to the input values.
- **Model Version**: The provider automatically uses the latest version of the moderation model or the specific version you specify.
- **Data Source Not Available**: There is no data source for moderation as OpenAI's API does not provide a way to retrieve moderation results by ID.

## Content Categories

The moderation API checks content against the following categories:

- **sexual**: Sexual content
- **hate**: Hateful, harassing, or violent content
- **harassment**: Harassment content
- **self-harm**: Content that promotes, encourages, or depicts acts of self-harm
- **sexual/minors**: Sexual content involving minors
- **hate/threatening**: Hateful content that also includes violence or serious harm 
- **violence**: Violent content
- **violence/graphic**: Graphic violence
- **self-harm/intent**: Content that expresses intention to commit self-harm
- **self-harm/instructions**: Content that encourages or provides instructions for self-harm
- **harassment/threatening**: Harassment content that also includes violence or serious harm

## Limitations

- The moderation API is designed for content policy adherence and should not be used as a replacement for human review.
- Results may vary depending on the specific moderation model used.
- The API supports multiple input types, but performance may vary with very long texts.
- Your OpenAI account must be active with billing properly configured to use this API.
- The API has rate limits and may reject requests if too many are made in a short period of time.
- OpenAI does not provide an API to retrieve moderation results by ID, so a data source is not available.

## Related Resources

- [OpenAI Moderation API Documentation](https://platform.openai.com/docs/api-reference/moderations)
- [Content Policy Guidelines](https://platform.openai.com/docs/guides/moderation)