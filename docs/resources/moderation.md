---
page_title: "OpenAI: openai_moderation Resource"
subcategory: ""
description: |-
  Performs content moderation using OpenAI's moderation endpoint.
---

# openai_moderation Resource

The `openai_moderation` resource checks text for content that may violate OpenAI's content policy. It analyzes text for categories like hate, threats, self-harm, sexual content, and violence. This resource is useful for implementing content filtering in applications.

~> **NOTE:** OpenAI does not provide an API to retrieve moderation results by ID. Therefore, a data source for moderation is not available. Moderation can only be used as a resource where you provide the input text and get the analysis results.

## Example Usage

```hcl
# Direct resource usage (recommended)
resource "openai_moderation" "example" {
  input = "I want to hurt someone."
  model = "text-moderation-latest"
  
  # Prevent replacement due to model version differences
  lifecycle {
    ignore_changes = [model]
  }
}

# Check if the content was flagged
output "is_flagged" {
  value = openai_moderation.example.flagged
}

# Get category scores
output "violence_score" {
  value = openai_moderation.example.categories_scores["violence"]
}

# Check multiple inputs
resource "openai_moderation" "multiple" {
  input = [
    "I want to hurt someone.",
    "The sky is blue and the grass is green.",
  ]
  model = "text-moderation-stable"
  
  # Prevent replacement due to model version differences
  lifecycle {
    ignore_changes = [model]
  }
}

output "flagged_inputs" {
  value = openai_moderation.multiple.results[*].flagged
}
```

## Argument Reference

* `input` - (Required) The text to moderate. Can be a string or a list of strings.
* `model` - (Optional) The moderation model to use. Options include:
  * `text-moderation-latest` - The most capable moderation model (default).
  * `text-moderation-stable` - A stable moderation model that may not have the latest capabilities.
* `api_key` - (Optional) Custom API key to use for this resource. If not provided, the provider's default API key will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - A unique identifier for this moderation resource.
* `flagged` - `true` if the input contains flagged content (for single input).
* `categories` - Map of moderation categories that were detected (for single input):
  * `hate` - Content that expresses, incites, or promotes hate based on race, gender, ethnicity, religion, nationality, sexual orientation, etc.
  * `hate/threatening` - Hateful content that also includes violence or serious harm towards the targeted group.
  * `harassment` - Content that expresses, incites, or promotes harassing language.
  * `harassment/threatening` - Harassment content that also includes violence or serious harm towards an individual.
  * `self-harm` - Content that promotes, encourages, or depicts acts of self-harm.
  * `self-harm/intent` - Content where the speaker expresses intent to commit acts of self-harm.
  * `self-harm/instructions` - Content that encourages performing acts of self-harm or provides instructions.
  * `sexual` - Content meant to arouse sexual excitement, such as descriptions of sexual activity.
  * `sexual/minors` - Sexual content that includes children.
  * `violence` - Content that depicts death, violence, or serious physical injury.
  * `violence/graphic` - Graphic violence content.
* `categories_scores` - Map of confidence scores (0.0 to 1.0) for each moderation category (for single input).
* `results` - List of moderation results for multiple inputs (when `input` is a list):
  * `flagged` - Whether the input was flagged.
  * `categories` - Map of detected moderation categories.
  * `categories_scores` - Map of confidence scores for each category.

## Known Issues and Solutions

### Model Version Differences

When you specify `text-moderation-latest` in your configuration, the OpenAI API might return a specific version (e.g., `text-moderation-007`) in the response. 

To prevent unnecessary resource replacement, you have two options:

1. **Provider-level handling (recommended)**: The provider now automatically handles this difference using a CustomizeDiff function that prevents unnecessary resource replacement. The `model` field is also marked as both `Optional` and `Computed` to allow the provider to store the actual model version returned by the API.

2. **Resource-level lifecycle block**: For additional protection, you can add a `lifecycle` block to your resource that ignores changes to the model attribute:

```hcl
resource "openai_moderation" "example" {
  input = "Text to moderate"
  model = "text-moderation-latest"
  
  lifecycle {
    ignore_changes = [model]
  }
}
```

For maximum stability, we recommend using both approaches together: letting the provider handle the differences automatically and adding the lifecycle block for extra protection.

## Import

Moderation resources can be imported using the moderation ID.

```bash
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