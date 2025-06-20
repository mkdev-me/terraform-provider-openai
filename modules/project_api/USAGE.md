# Using the OpenAI Project API Key Module

This document provides detailed instructions for importing and managing existing OpenAI Project API keys with Terraform.

## Overview

The `project_api` module allows you to:

1. Import existing OpenAI project API keys into Terraform state
2. Reference key metadata in other resources 
3. Manage the lifecycle of imported keys (e.g., deleting them when needed)

## Important Limitations

- **Project API keys cannot be created programmatically** via the OpenAI API
- You must create keys manually in the OpenAI dashboard first
- The actual API key value is not exposed via the API (for security reasons)

## Importing Existing Keys

### Step 1: Create API Key Manually

1. Log in to the [OpenAI Platform](https://platform.openai.com)
2. Navigate to the API keys section
3. Create a new API key for your project
4. Copy the API key ID (format: `key_XXXX`) - this appears in the dashboard
5. Note your project ID (format: `proj_XXXX`)

### Step 2: Configure the Module

Create a Terraform configuration file (e.g., `main.tf`):

```hcl
module "openai_key" {
  source     = "./modules/project_api"
  project_id = "proj_abc123"  # Your actual project ID
  name       = "terraform-managed-key"  # Name matching the key in dashboard
}

# Optional: Output the key ID for reference
output "api_key_id" {
  value = module.openai_key.api_key_id
}
```

### Step 3: Initialize and Apply

Run the following commands:

```bash
# Initialize Terraform
terraform init

# Create placeholder resources
terraform apply
```

### Step 4: Import the Key

Import the existing API key:

```bash
terraform import module.openai_key.openai_project_api_key.this "proj_abc123:key_xyz789"
```

Replace:
- `proj_abc123` with your actual project ID
- `key_xyz789` with your actual API key ID

### Step 5: Verify the Import

Check that the import was successful:

```bash
terraform state show module.openai_key.openai_project_api_key.this
```

You should see the API key's metadata including ID, name, and creation timestamp.

## Managing Multiple Keys

You can import and manage multiple API keys by creating multiple module instances:

```hcl
module "development_key" {
  source     = "./modules/project_api"
  project_id = "proj_abc123"
  name       = "dev-key"
}

module "production_key" {
  source     = "./modules/project_api"
  project_id = "proj_def456" 
  name       = "prod-key"
}
```

Import each key separately:

```bash
# Import development key
terraform import module.development_key.openai_project_api_key.this "proj_abc123:key_dev789"

# Import production key
terraform import module.production_key.openai_project_api_key.this "proj_def456:key_prod123"
```

## Using Key References in Other Resources

After importing, you can reference the key's metadata in other resources:

```hcl
resource "some_resource" "example" {
  # ...
  api_key_id = module.openai_key.api_key_id
  # ...
}
```

## Removing Keys

To delete an imported API key:

```bash
terraform destroy -target=module.openai_key.openai_project_api_key.this
```

**Warning**: This will delete the actual API key from OpenAI's servers, not just from Terraform state.

## Troubleshooting

### Import Failed

If the import fails, check:
1. The project ID and API key ID are correct
2. You have permissions to manage the project
3. The API key still exists in the OpenAI dashboard

### Changes Detected After Import

If Terraform shows changes immediately after import, you may need to adjust the `name` attribute in your configuration to match the actual key name in the OpenAI dashboard. 