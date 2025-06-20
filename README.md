# Terraform Provider for OpenAI

[![Release](https://img.shields.io/github/v/release/fjcorp/terraform-provider-openai)](https://github.com/fjcorp/terraform-provider-openai/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/fjcorp/terraform-provider-openai)](https://goreportcard.com/report/github.com/fjcorp/terraform-provider-openai)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL--2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A comprehensive Terraform provider for managing OpenAI resources, enabling infrastructure-as-code management of OpenAI's API services including chat completions, assistants, fine-tuning, embeddings, and more.

## Features

### Core AI Capabilities
- **Chat Completions**: Generate conversational responses with GPT-4, GPT-3.5-Turbo, and other models
- **Assistants API**: Create and manage AI assistants with custom instructions and file knowledge
- **Fine-Tuning**: Create custom models by fine-tuning base models on your data
- **Embeddings**: Generate vector representations of text for semantic search and analysis
- **Image Generation**: Create, edit, and generate variations of images using DALL-E
- **Audio Processing**: Convert speech to text (transcription) and text to speech
- **Moderation**: Detect potentially harmful content in text

### Resource Management
- **File Management**: Upload, organize, and manage files for training and assistants
- **Vector Stores**: Create and manage vector databases for retrieval-augmented generation
- **Batch Processing**: Process multiple requests efficiently with batch operations
- **Thread Management**: Create and manage conversation threads for assistants

### Administrative Features
- **Organization Management**: Manage projects, users, and service accounts
- **API Key Management**: Handle project and admin API keys securely
- **Rate Limit Control**: Configure and monitor rate limits for your projects
- **Usage Tracking**: Monitor token usage and costs across resources

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for development)
- OpenAI API Key

## Installation

### Using Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "~> 0.1"
    }
  }
}

provider "openai" {
  api_key = var.openai_api_key
}
```

### Manual Installation

For development or if you need the latest unreleased version:

```bash
# Clone the repository
git clone https://github.com/fjcorp/terraform-provider-openai.git
cd terraform-provider-openai

# Build and install in one command
make install
```

## Development

For provider developers, this repository includes a comprehensive Makefile with targets for building, testing, and managing the provider.

### Quick Start for Developers

```bash
# Clone and setup
git clone https://github.com/fjcorp/terraform-provider-openai.git
cd terraform-provider-openai

# Install dependencies and build
make install

# Run tests
make test
```

### Available Make Targets

#### Building and Installing
```bash
# Build the provider binary
make build

# Build and install the provider locally for testing
make install

# Build for multiple platforms (creates ./bin/ directory with cross-compiled binaries)
make release

# Clean up built binaries and temporary files
make clean
```

#### Code Quality
```bash
# Format Go code
make fmt

# Run linting checks
make lint

# Run unit tests with coverage
make test

# Run acceptance tests (requires valid OpenAI API keys)
make testacc
```

#### Testing with Examples

To test the provider with specific examples, navigate to the example directories:

```bash
# Test a specific example
cd examples/chat_completion
terraform init
terraform plan
terraform apply

# Or test another example
cd ../assistants
terraform init
terraform plan
```

#### Development Workflow

```bash
# 1. Make your changes to the provider code
# 2. Format and test your code
make fmt lint test

# 3. Run acceptance tests (optional, requires API keys)
make testacc

# 4. Clean up when done
make clean
```

#### Environment Variables for Testing

Set these environment variables for testing:

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_ADMIN_KEY="your-admin-key"  # For admin operations
export OPENAI_ORGANIZATION_ID="your-org-id"  # Optional
```

## Testing

### Environment Setup

Set the required environment variables:

```bash
# Option 1: Export directly
export OPENAI_API_KEY="sk-proj-..."
export OPENAI_ADMIN_KEY="sk-admin-..."  # Required for organization resources

# Option 2: Use a .env file
cat > .env << EOF
OPENAI_API_KEY=sk-proj-...
OPENAI_ADMIN_KEY=sk-admin-...
EOF
source .env
```

### Testing Examples

Use the test script in `testing/`:

```bash
# Quick verification test
./testing/test_examples.sh quick

# Test all examples (plan only)
./testing/test_examples.sh plan

# Test a specific example (plan only)
./testing/test_examples.sh plan image

# Test with real resources (apply + destroy)
./testing/test_examples.sh apply chat_completion

# Test all examples with apply (WARNING: creates resources!)
./testing/test_examples.sh apply

# Clean up all terraform files
./testing/test_examples.sh cleanup
```

### Manual Testing

You can also test examples directly:

```bash
cd examples/chat_completion
terraform init
terraform plan
terraform apply
terraform destroy
```

## Basic Usage

```hcl
terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "~> 0.1"
    }
  }
}

provider "openai" {
  # API key and organization ID are read from environment variables:
  # OPENAI_API_KEY and OPENAI_ORGANIZATION_ID
}

# Create a chat completion
resource "openai_chat_completion" "example" {
  model = "gpt-4"
  
  messages {
    role    = "system"
    content = "You are a helpful assistant."
  }
  
  messages {
    role    = "user"
    content = "Hello, how are you?"
  }
}

# Output the assistant's response
output "response" {
  value = openai_chat_completion.example.choices[0].message.content
}
```

## Resources and Data Sources

The provider includes numerous resources and data sources for all OpenAI API features. 
See the [resources](RESOURCES.md) documentation for a complete list of resources and the [data sources](DATA_SOURCES.md) documentation for a complete list of data sources.

### Key Data Sources

- `openai_model_response`: Retrieve a specific model response by ID
- `openai_model_response_input_items`: Get the input items for a model response
- `openai_model_responses`: List multiple model responses (requires browser session authentication)
- `openai_file`: Retrieve metadata for a specific file
- `openai_files`: List all files in your organization
- `openai_fine_tuning_job`: Get details about a fine-tuning job
- `openai_rate_limit`: Retrieve rate limit information for a specific model
- `openai_rate_limits`: List all rate limits for a project
- `openai_organization_user`: Retrieve information about a specific user in your organization
- `openai_organization_users`: List all users in your organization with filtering options
- See [docs/data-sources](docs/data-sources) for complete documentation

### Key Resources

- `openai_chat_completion`: Generate chat completions
- `openai_file`: Upload files to OpenAI
- `openai_fine_tuning_job`: Create and manage fine-tuning jobs
- `openai_image_generation`: Generate images with DALL-E
- `openai_rate_limit`: Set rate limits for specific models in a project
- See [docs/resources](docs/resources) for complete documentation


## Authentication and API Key Requirements

The OpenAI Terraform Provider requires different types of API keys depending on the resources you're managing. Here's a breakdown of which resources require which types of keys:

### Admin API Key vs. Project API Key

| Key Type | Description | Environment Variable | Provider Configuration |
|----------|-------------|----------------------|------------------------|
| **Admin API Key** | Organization-level key with administrative permissions | `OPENAI_ADMIN_KEY` | `admin_key = var.openai_admin_key` |
| **Project API Key** | Limited to specific project operations | `OPENAI_API_KEY` | `api_key = var.openai_api_key` |

### Resources Requiring Admin API Key

These resources and data sources require organization admin permissions:

| Resource/Data Source | Description | 
|----------------------|-------------|
| `openai_project` | Create and manage OpenAI projects |
| `openai_projects` | List all projects in an organization |
| `openai_organization_user` | Retrieve organization user information |
| `openai_organization_users` | List all organization users |
| `openai_project_user` | Manage user access to projects |
| `openai_project_users` | List users in a project |
| `openai_invite` | Create and manage organization invites |
| `openai_invites` | List all organization invites |
| `openai_rate_limit` | Manage rate limits for models in projects |

### Resources That Work with Project API Key

These resources and data sources can be used with project-scoped API keys:

| Resource/Data Source | Description |
|----------------------|-------------|
| `openai_chat_completion` | Generate chat completions |
| `openai_assistant` | Create and manage assistants |
| `openai_thread` | Create and manage conversation threads |
| `openai_file` | Upload and manage files |
| `openai_image_generation` | Generate images with DALL-E |
| `openai_embedding` | Create text embeddings |
| `openai_model_response` | Generate text with models |
| `openai_fine_tuning_job` | Create fine-tuning jobs |

### Using the Correct Keys

For proper operation:

1. **Provider Configuration**: Set both keys in the provider block when working with mixed resources:
   ```hcl
   provider "openai" {
     api_key   = var.openai_project_api_key
     admin_key = var.openai_admin_key
   }
   ```

2. **Resource-Specific Keys**: Override the key for specific resources:
   ```hcl
   resource "openai_rate_limit" "example" {
     project_id = "proj_abc123"
     model      = "gpt-4"
     # Rate limits always require an admin key with appropriate permissions
     api_key    = var.openai_admin_key
   }
   ```

3. **Data Source Configuration**: Similar to resources, you can specify which key to use:
   ```hcl
   data "openai_projects" "all" {
     api_key = var.openai_admin_key
   }
   ```

4. **Priority Rules**: When both keys are provided, the provider uses:
   - Admin key for administrative operations (projects, users, etc.)
   - Project key for model usage operations (completions, embeddings, etc.)
   - Resource-specific `api_key` parameter overrides the provider defaults

By using the correct key for each operation, you ensure proper permissions while maintaining security best practices.

## Modules

Reusable modules are available for common patterns:

- [upload](modules/upload/): Upload and import files
- [files](modules/files/): Comprehensive file management
- [fine_tuning](modules/fine_tuning/): Create fine-tuning jobs
- [chat_completion](modules/chat_completion/): Generate chat completions
- [model_response](modules/model_response/): Generate text with comprehensive token usage statistics
- [audio](modules/audio/): Process audio files
- [and many more](modules/)

## Examples

The [examples](examples/) directory contains working examples for all features.

## Initial Setup Recommendations

For initial setup, it's recommended to:

1. Create a `.env` file to store your keys:
   ```
   OPENAI_API_KEY=sk-proj-xxxx
   OPENAI_ADMIN_KEY=sk-xxxx
   OPENAI_ORGANIZATION_ID=org-xxxx
   ```

2. Source the file before running Terraform:
   ```bash
   source .env
   terraform apply
   ```

3. Or use a terraform.tfvars file and explicitly configure the provider:
   ```hcl
   # In provider block
   provider "openai" {
     api_key         = var.openai_api_key
     admin_key       = var.openai_admin_key  # For admin operations
     organization_id = var.openai_org_id     # Optional
   }
   ```

## API Key Troubleshooting

If you encounter API key issues:
- Verify keys are correctly set with `echo $OPENAI_API_KEY`
- Explicitly set keys in provider configuration instead of using environment variables
- For resource-specific keys, use the `api_key` parameter on supported resources
- See [Troubleshooting Guide](docs/TROUBLESHOOTING.md#api-key-configuration-and-troubleshooting) for more details

## Documentation

### Getting Started
- [Installation](#installation): Setup instructions
- [Basic Usage](#basic-usage): Quick start guide
- [Examples](examples/): Complete working examples for all features

### Reference Documentation
- [Terraform Registry Documentation](https://registry.terraform.io/providers/fjcorp/openai/latest/docs): Official provider documentation
- [Resources](docs/resources/): Comprehensive resource documentation
- [Data Sources](docs/data-sources/): Data source documentation
- [Modules](modules/): Reusable Terraform modules

### Guides and Support
- [Authentication Guide](#authentication-and-api-key-requirements): API key setup and requirements
- [Testing Guide](#testing): How to test the provider
- [Development Guide](#development): For contributors
- [Troubleshooting](docs/TROUBLESHOOTING.md): Common issues and solutions

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to get started.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, feature requests, or questions:
- [GitHub Issues](https://github.com/fjcorp/terraform-provider-openai/issues)
- [Documentation](docs/)
- [Examples](examples/)

