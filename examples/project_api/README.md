# OpenAI Project API Keys Example

This example demonstrates how to retrieve information about existing OpenAI project API keys using Terraform.

## Important Note

**OpenAI does NOT support programmatically creating project API keys via their API.**

* You must create project API keys manually in the OpenAI dashboard first
* This example allows you to retrieve metadata about existing keys through Terraform

## What This Example Does

This example provides two main ways to work with OpenAI Project API keys:

1. **Retrieve information about an existing API key**
   * Look up a specific API key by ID
   * View metadata including name, creation date, and last used date

2. **Retrieve all API keys for a project**
   * List all API keys for a given project
   * Get counts and details for inventory/audit purposes

## Usage Instructions

### Prerequisites

* An OpenAI account with access to project API keys
* An OpenAI Admin API key with appropriate permissions (stored as environment variable)
* Terraform installed on your system

### Running the Example

1. **Set your OpenAI Admin API key as an environment variable:**
   ```bash
   export OPENAI_ADMIN_KEY=your_admin_api_key_here
   ```

2. **Initialize Terraform:**
   ```bash
   terraform init
   ```

3. **Apply the configuration:**
   ```bash
   terraform apply -var="openai_admin_key=$OPENAI_ADMIN_KEY"
   ```
   
   * If you want to look up a specific API key, add the API key ID:
   ```bash
   terraform apply -var="openai_admin_key=$OPENAI_ADMIN_KEY" -var="existing_api_key_id=key_abc123"
   ```

## Understanding the Code

* The example creates a demo project and then demonstrates two data source patterns:
  * `data.openai_project_api_key` - For retrieving a specific key by ID
  * `data.openai_project_api_keys` - For retrieving all keys for a project

* The outputs display different information depending on what you're looking up:
  * Information about a specific key (when providing `existing_api_key_id`)
  * Information about all keys (showing counts and details)

## Best Practices

* Never store API keys directly in Terraform code
* Use environment variables or a secure secrets manager to handle sensitive keys
* Remember that OpenAI API keys are powerful and should be handled with appropriate security measures
* Audit your API keys regularly to ensure proper usage and security

## Troubleshooting

If you encounter issues:

1. Ensure your admin API key has the correct permissions
2. Verify the project exists in your OpenAI account
3. If looking up a specific key, confirm the API key ID is correct
4. Check for rate limiting or API access issues

## Example Files

* `main.tf`: Contains provider configuration, project creation, and API key retrieval
* `import_project_key.sh`: Script to help import and set up project API keys

## Usage Methods

This example provides two approaches:

1. Using Terraform to retrieve details about existing API keys
2. Using the provided import script to automate importing existing keys

## Prerequisites

* OpenAI account with admin API key
* Terraform 0.14.x or higher
* An existing project in OpenAI with at least one API key (or use the one created by this example)

## Setup

1. Create a `terraform.tfvars` file with your OpenAI credentials:

```hcl
openai_admin_key = "sk-your-openai-admin-key"
existing_api_key_id = "key_abc123xyz"  # ID of an existing API key to look up
```

## Usage

Initialize your Terraform configuration:

```bash
terraform init
```

Apply the configuration:

```bash
terraform apply
```

## Using the Import Script

For importing existing API keys, you can use the provided import script:

```bash
./import_project_key.sh <project_id> <key_name>
```

The script features:
- Automatically looks up the API key ID if you provide only the name
- Validates inputs and displays helpful error messages
- Updates the Terraform file with the correct values
- Requires OPENAI_ADMIN_KEY environment variable to be set

## Example Module Usage

### Retrieving a Specific API Key

```hcl
# Create or reference an existing project
resource "openai_project" "example" {
  name = "Project API Key Example"
}

# Retrieve information about a specific API key
module "existing_key" {
  source = "../../modules/project_api"

  project_id = openai_project.example.id
  api_key_id = var.existing_api_key_id

  openai_admin_key = var.openai_admin_key
}

output "api_key_name" {
  value = module.existing_key.name
}

output "api_key_created_at" {
  value = module.existing_key.created_at
}
```

### Retrieving All API Keys for a Project

```hcl
module "all_keys" {
  source = "../../modules/project_api"

  project_id   = openai_project.example.id
  retrieve_all = true

  openai_admin_key = var.openai_admin_key
}

output "project_api_keys" {
  value = [for key in module.all_keys.api_keys : key.name]
}
```

## Note on API Key Security

* API key values are never exposed through these data sources
* Only metadata like name, ID, creation time, and last used time are available
* This ensures security while still allowing you to track and reference your keys

## Manual Creation of Project API Keys

To create a project API key:

1. Log in to the [OpenAI Platform](https://platform.openai.com/)
2. Navigate to your project
3. Go to the API Keys section
4. Create a new API key
5. Copy the key value (only shown once) and store it securely
6. Note the key ID for use in Terraform

## Related Resources

* For creating and managing admin API keys programmatically, see the [Admin API Key example](../admin_api_key)
