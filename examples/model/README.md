# OpenAI Model Examples

This directory contains examples demonstrating how to use the OpenAI model data sources in Terraform.

## Usage

To run these examples, you need to set up your OpenAI API key. You can do this by setting the `OPENAI_API_KEY` environment variable:

```bash
export OPENAI_API_KEY="your-api-key"
```

If you want to use a project API key for certain resources, you can pass it as a variable:

```bash
terraform apply -var="project_api_key=sk-project-..."
```

Alternatively, you can create a `terraform.tfvars` file with your sensitive variables:

```
project_api_key = "sk-project-..."
```

Then, you can run the examples:

```bash
terraform init
terraform plan
terraform apply
```

## Examples

The examples demonstrate:

1. Retrieving information about a specific model (`gpt-4o` in this case) using the provider's default API key
2. Retrieving information about the same model but using a project API key
3. Retrieving information about all available models
4. Using output values to display model information

## Using Project API Keys vs. Organization Admin Keys

This provider supports both organization admin keys and project-specific API keys:

- **Organization Admin Key**: Set in the provider block or via `OPENAI_API_KEY` environment variable. Has access to all resources across the organization.
- **Project API Key**: Set via the `api_key` argument on specific resources/data sources. Has limited access to only resources within its project.

For model-related resources, we recommend using project API keys when possible, as they provide better security through isolation.

## Troubleshooting API Key Issues

If you encounter errors like:

```
Error: Error retrieving model: API error: Incorrect API key provided: ''. You can find your API key at https://platform.openai.com/account/api-keys.
```

Try these solutions:

### 1. Explicitly Set the Provider API Key

Update the provider block in your configuration:

```hcl
provider "openai" {
  api_key = var.openai_api_key
}

variable "openai_api_key" {
  description = "OpenAI API Key"
  type        = string
  sensitive   = true
}
```

Then run terraform with the key specified:

```bash
terraform apply -var="openai_api_key=sk-your-key-here" -var="project_api_key=sk-proj-your-key-here"
```

### 2. Create a terraform.tfvars File

Create a file named `terraform.tfvars` in this directory with:

```
openai_api_key = "sk-your-key-here"
project_api_key = "sk-proj-your-key-here"
```

This is often more reliable than environment variables when working across different environments.

### 3. Check for Environment Differences

If these examples work on one machine but not another, check:
- The OS and shell environment 
- Terminal session inheritance of environment variables
- Authorization level of the used API keys

## Resources Created

These examples only use data sources and do not create any resources in your OpenAI account.

## Outputs

The examples provide the following outputs:

- `model_info`: Detailed information about the `gpt-4o` model
- `model_count`: Total number of available models 
- `available_models`: List of IDs for all available models 