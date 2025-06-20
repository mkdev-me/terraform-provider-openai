# OpenAI Service Account Examples

This directory contains examples for working with OpenAI service accounts using Terraform.

## IMPORTANT: Admin API Key vs Project API Key

These examples require an **organization-level admin API key**, not a regular project API key:

- **Project API Keys**: Limited to a specific project's resources only
- **Admin API Keys**: Organization-wide admin keys with special scopes

Regular project API keys **will not work** for service account operations even if exported as `OPENAI_API_KEY`.

## Permission Requirements

OpenAI service accounts require specific API key permissions depending on what you're trying to do:

1. **Creating service accounts**: Requires an admin API key with the `api.organization.projects.service_accounts.write` scope
2. **Reading service accounts**: Requires an admin API key with the `api.organization.projects.service_accounts.read` scope

If you don't have these scopes on your admin key, you'll see permission errors.

## Examples

This directory contains two main examples:

### 1. Creating a Service Account

The `main.tf` file demonstrates how to create a service account in an OpenAI project. By default, this functionality is disabled to avoid permission errors, but you can enable it by setting the `try_create_service_account` variable to `true`:

```bash
terraform apply -var="try_create_service_account=true" -var="openai_admin_key=sk-admin-..."
```

### 2. Reading Service Account Data

The `data_sources.tf` file demonstrates how to read information about existing service accounts using data sources. By default, this functionality is also disabled, but you can enable it by setting the `try_data_sources` variable to `true`:

```bash
terraform apply -var="try_data_sources=true" -var="openai_admin_key=sk-admin-..."
```

## Using Both Examples Together

You can enable both creation and reading in the same apply:

```bash
terraform apply -var="try_create_service_account=true" -var="try_data_sources=true" -var="openai_admin_key=sk-admin-..."
```

### Helper Script

A convenient `run.sh` script is provided to help you run the examples more easily:

```bash
# Usage
./run.sh YOUR_OPENAI_ADMIN_KEY

# Example
./run.sh sk-admin-your-key-here
```

This script will apply the configuration with both service account creation and data sources enabled.

## Required Variables

| Name | Description | Required |
|------|-------------|:--------:|
| `openai_admin_key` | OpenAI Admin API Key with the required scopes (organization-level, not project-level) | Yes |
| `try_create_service_account` | Whether to try creating a service account (optional, default: false) | No |
| `try_data_sources` | Whether to try using service account data sources (optional, default: false) | No |

## Notes on API Keys

1. The OpenAI provider does not support programmatic creation of project API keys. After creating a service account, you must manually create API keys in the OpenAI dashboard.

2. API keys must have the appropriate scopes for the operations you're trying to perform. If your admin key lacks these scopes, you'll see permission errors.

## Export the Admin API Key

You can export your admin API key, but be sure to use `TF_VAR_openai_admin_key` and not `OPENAI_API_KEY`:

```bash
# CORRECT: Use this for admin operations
export TF_VAR_openai_admin_key="sk-admin-..."

# INCORRECT: This will not work for admin operations
# export OPENAI_API_KEY="sk-project-..."
```

## Verifying Your Admin Key

You can verify that your admin key has the correct permissions by using curl:

```bash
# Check if you can access a service account
curl https://api.openai.com/v1/organization/projects/YOUR_PROJECT_ID/service_accounts/YOUR_SERVICE_ACCOUNT_ID \
  -H "Authorization: Bearer YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json"

# A 200 response indicates your key has the correct permissions
# A 401 or 403 response typically indicates permission issues
```

## Managing Permission Errors

The examples in this directory are designed to be resilient to permission errors:

1. Resources are only created when explicitly enabled
2. Data sources are only queried when explicitly enabled
3. Fallback values are used for all outputs when permissions are missing
4. Dependencies are carefully managed to ensure data sources are only accessed after resources are created

This design allows you to run the examples even without all permissions and see what would happen if you had the right permissions.

## Building the Provider from Source

If you need to use the latest version of the provider or make custom modifications, you can build it from source:

```bash
cd /path/to/terraform-provider-openai
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o terraform-provider-openai cmd/main.go
cp terraform-provider-openai ~/.terraform.d/plugins/registry.terraform.io/mkdev-me/openai/1.0.0/darwin_arm64/
```

Then navigate to the example directory and apply:

```bash
cd examples/service_account
rm -rf .terraform*
terraform init
terraform apply -auto-approve -var="openai_admin_key=YOUR_ADMIN_KEY" -var="try_create_service_account=true" -var="try_data_sources=true"
``` 