# OpenAI Moderation Examples

This directory contains examples demonstrating how to use the OpenAI moderation resource to detect harmful content in text.

~> **NOTE:** OpenAI does not provide an API to retrieve moderation results by ID. Therefore, a data source for moderation is not available. Moderation can only be used as a resource where you provide the input text and get the analysis results.

## Prerequisites

- An OpenAI API key with an active billing account
- Terraform installed

## Setup

1. Set your OpenAI API key as an environment variable:
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

## Examples Included

### 1. Single Text Moderation

A basic example showing how to moderate a single piece of text:

```hcl
resource "openai_moderation" "single_text" {
  input = "I want to kill them."
}
```

### 2. Multiple Text Moderation

For multiple texts, create separate resources for each text to ensure proper state handling:

```hcl
resource "openai_moderation" "harmless_text" {
  input = "This is a completely harmless message."
  model = "text-moderation-latest"
}

resource "openai_moderation" "harmful_text" {
  input = "I want to make a bomb and hurt people."
  model = "text-moderation-latest"
}

resource "openai_moderation" "neutral_text" {
  input = "I like watching movies and reading books."
  model = "text-moderation-latest"
}
```

### 3. Module Approach (Alternative)

You can also use the moderation module:

```hcl
module "single_moderation" {
  source = "../../modules/moderation"
  
  input = "I want to kill them."
}

module "batch_moderation_1" {
  source = "../../modules/moderation"
  
  input = "This is a completely harmless message."
  model = "text-moderation-latest"
}
```

However, the direct resource approach is preferred for clearer state management.

## Import Support

Moderation resources can be imported using the moderation ID:

```bash
# Import with direct resource reference
terraform import openai_moderation.harmful_text modr-12345abcdef

# Import with module reference
terraform import module.batch_moderation_1.openai_moderation.this modr-12345abcdef
```

~> **IMPORTANT:** OpenAI does not provide an API endpoint to retrieve moderation results by ID. When importing a moderation resource, the original input text and detailed analysis results cannot be retrieved from the API. Instead, placeholder values will be used:
- The input field will show `[Imported: Original input not available]`
- The moderation categories, scores, and results fields will be empty or contain default values
- The `_api_response` will contain minimal information

This means that after importing, the resource's state will not match its original creation state. If you run `terraform plan` after importing, Terraform will detect differences and plan to update the resource on the next `terraform apply`, which will recreate the moderation with the input text specified in your configuration.

### Example Import Workflow

Below is an example workflow showing what happens when importing a moderation resource:

1. Original state before removal:
   ```
   $ terraform state show openai_moderation.harmful_text
   # openai_moderation.harmful_text:
   resource "openai_moderation" "harmful_text" {
     _api_response   = jsonencode({...}) # Full API response with categories, scores, etc.
     categories      = {
       "violence"  = true
       # ... other categories with true/false values
     }
     category_scores = {
       "violence"  = 0.9797691106796265
       # ... other category scores
     }
     flagged         = true
     id              = "modr-BKRRldUxn3PelHSUhqPKjmpJpInOw"
     input           = "I want to make a bomb and hurt people."
     # ... other attributes
   }
   ```

2. Remove the resource from state:
   ```
   $ terraform state rm openai_moderation.harmful_text
   Removed openai_moderation.harmful_text
   Successfully removed 1 resource instance(s).
   ```

3. Import the resource back using its ID:
   ```
   $ terraform import openai_moderation.harmful_text modr-BKRRldUxn3PelHSUhqPKjmpJpInOw
   openai_moderation.harmful_text: Importing from ID "modr-BKRRldUxn3PelHSUhqPKjmpJpInOw"...
   openai_moderation.harmful_text: Import prepared!
   Prepared openai_moderation for import
   openai_moderation.harmful_text: Refreshing state... [id=modr-BKRRldUxn3PelHSUhqPKjmpJpInOw]
   
   Import successful!
   ```

4. Imported state after import (notice the missing data):
   ```
   $ terraform state show openai_moderation.harmful_text
   # openai_moderation.harmful_text:
   resource "openai_moderation" "harmful_text" {
     _api_response   = jsonencode({
       id      = "modr-BKRRldUxn3PelHSUhqPKjmpJpInOw"
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
     id              = "modr-BKRRldUxn3PelHSUhqPKjmpJpInOw"
     input           = "[Imported: Original input not available]"
     # ... other attributes with empty/default values
   }
   ```

5. When you run `terraform plan` or `terraform apply` after importing, Terraform will detect differences between the configuration (with your specified input text) and the imported state (with placeholder values), and will recreate the resource.

## Outputs

The examples output detailed moderation results, including:

- Whether content was flagged
- Which content categories were detected
- Confidence scores for each category

Example output structure:

```
Outputs:

batch_texts_categories = {
  "text1" = {
    "harassment" = false
    "hate" = false
    # ... other categories
  }
  "text2" = {
    "violence" = true
    # ... other categories
  }
  # ... text3 categories
}

batch_texts_flagged = {
  "text1" = false
  "text2" = true
  "text3" = false
}

single_text_flagged = true
```

## Common API Errors

If you encounter the following errors when running the examples:

### Billing Not Active

**Error:** `OpenAI billing not active: Your account is not active, please check your billing details on our website.`

**Solution:**
1. Visit [OpenAI Billing](https://platform.openai.com/account/billing)
2. Add a valid payment method
3. Ensure you have sufficient credits or payment method

### Rate Limiting

**Error:** `OpenAI rate limit exceeded: Too Many Requests`

**Solution:**
1. In production environments, implement retry logic with exponential backoff
2. Reduce the frequency of requests
3. Consider using a queueing system to manage API call rates

## Best Practice

For consistent results with Terraform state management, always use separate resources (or module instances) for each text you want to moderate rather than attempting to use array inputs. Using direct resource declarations (`resource "openai_moderation"`) makes the resources appear directly in the state list, which simplifies management and visibility.

## Model Version Differences

When you specify `text-moderation-latest` in your configuration, the OpenAI API might return a specific version (e.g., `text-moderation-007`) in the response. 

To prevent unnecessary resource replacement, these examples use a dual approach:

1. **Provider-level handling**: The provider automatically handles model version differences using a CustomizeDiff function that prevents unnecessary resource replacement. The actual model version returned by the API is stored in the state.

2. **Resource-level lifecycle block**: Each resource also includes a `lifecycle` block that explicitly tells Terraform to ignore model changes:

```hcl
resource "openai_moderation" "example" {
  input = "Text to moderate"
  model = "text-moderation-latest"
  
  lifecycle {
    ignore_changes = [model]
  }
}
```

This dual approach ensures maximum stability and prevents Terraform from trying to recreate the resource when only the model version differs - even if the provider-level handling were to change in a future version.