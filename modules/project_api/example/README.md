# OpenAI Project API Key Import Example

This example demonstrates how to import existing OpenAI project API keys using the `project_api` module.

## Prerequisites

1. An OpenAI account with API access
2. At least one project with an API key created manually through the OpenAI dashboard
3. Terraform installed locally

## Setup

1. Update the `main.tf` file with your actual project IDs and key names:

```hcl
module "dev_project_api_key" {
  source     = "../../project_api"
  project_id = "proj_abc123"  # Replace with your actual project ID
  name       = "dev-key"      # Replace with your actual key name
}
```

2. Initialize Terraform:

```bash
terraform init
```

3. Apply the configuration to create placeholders:

```bash
terraform apply
```

## Import Process

1. Find your API key ID in the OpenAI dashboard (format: `key_XXXX`)

2. Import the key using the following command:

```bash
terraform import module.dev_project_api_key.openai_project_api_key.this "proj_abc123:key_dev456"
```

Replace:
- `proj_abc123` with your actual project ID
- `key_dev456` with your actual API key ID

3. Verify the import was successful:

```bash
terraform state show module.dev_project_api_key.openai_project_api_key.this
```

## Managing Multiple Keys

The example includes configuration for both development and production keys. To import multiple keys, run separate import commands for each key:

```bash
# Import development key
terraform import module.dev_project_api_key.openai_project_api_key.this "proj_abc123:key_dev456"

# Import production key
terraform import module.prod_project_api_key.openai_project_api_key.this "proj_xyz789:key_prod123"
```

## Notes

- The imported API key can be referenced in other Terraform resources via outputs
- The actual API key value is not available via the API and won't be in your state
- If you delete this resource in Terraform, it will attempt to delete the actual API key 