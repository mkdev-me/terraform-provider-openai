# OpenAI Admin API Key Management Example

This example demonstrates how to create and retrieve Admin API keys with the OpenAI Terraform provider.

## Prerequisites

- An OpenAI API key with administrative permissions (`api.management.read`, `api.management.write`)
- Terraform installed on your system
- Basic knowledge of Terraform

## Features

This example showcases:

1. Creating an admin API key with specific permissions and an expiration date
2. Retrieving a specific admin API key by ID using a data source
3. Listing all admin API keys in the organization
4. Saving the API key to a secure local file

## Usage

1. Clone the repository:
   ```
   git clone https://github.com/fjcorp/terraform-provider-openai.git
   cd terraform-provider-openai/examples/system_api
   ```

2. Create a `terraform.tfvars` file with your admin API key:
   ```hcl
   openai_admin_key = "sk-adm-your-admin-api-key"
   ```

   Alternatively, you can set it as an environment variable:
   ```
   export TF_VAR_openai_admin_key="sk-adm-your-admin-api-key"
   ```

3. Initialize Terraform:
   ```
   terraform init
   ```

4. Apply the configuration:
   ```
   terraform apply
   ```

5. Review the created and retrieved keys:
   ```
   terraform output
   ```

   To view sensitive values:
   ```
   terraform output -json created_admin_key_value
   ```

## Key Features

1. **Resource-Based API Key Creation**
   - Direct resource usage with `openai_admin_api_key`
   - Customizable expiration date
   - Permission scopes to restrict access

2. **Data-Based API Key Retrieval**
   - Retrieve specific keys by ID with `openai_admin_api_key` data source
   - List all keys with `openai_admin_api_keys` data source
   - Pagination support for organizations with many keys

## Security Note

The admin API key values are marked as sensitive in Terraform outputs. However, they might still appear in the Terraform state file. Make sure to secure your state files appropriately.

## Variables

| Name | Description | Default | Required |
|------|-------------|---------|----------|
| `openai_admin_key` | OpenAI Admin API key with administrative permissions | - | Yes |

## Outputs

| Name | Description |
|------|-------------|
| `created_admin_key` | Details about the created admin key (ID, name, creation time) |
| `created_admin_key_value` | Secret value of the created admin key (sensitive) |
| `retrieved_admin_key` | Details about the retrieved admin key using data source |
| `all_admin_keys_count` | Number of admin API keys in the organization |
| `all_admin_key_names` | List of names of all admin API keys | 