# Terraform Provider for OpenAI

A comprehensive Terraform provider for managing OpenAI resources.

## Features

- **ChatGPT & GPT-4**: Create chat completions with the latest models
- **File Management**: Upload, read, and delete files with import support
- **Fine-Tuning**: Create and manage fine-tuning jobs with custom models
- **Images**: Generate, edit, and create variations of images
- **Embeddings**: Create embeddings for text and retrieve them
- **Audio**: Convert speech to text and text to speech
- **Moderation**: Detect harmful content in text
- **Organization Management**: Manage projects, users, and service accounts
- **Model Responses**: Generate and retrieve responses with input data and comprehensive token usage statistics
- **Rate Limits**: Manage and retrieve rate limits for models in your OpenAI projects

## Installation

Since this provider is not yet officially available in the Terraform Registry, you'll need to build and install it locally:

```bash
# Clone the repository (if you haven't already)
git clone https://github.com/fjcorp/terraform-provider-openai.git
cd terraform-provider-openai

# Build the provider binary
go build -o terraform-provider-openai 

```

Then, create the appropriate plugin directory based on your operating system and architecture:

### For macOS:
```bash
# Apple Silicon (M1/M2/M3)
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/fjcorp/openai/1.0.0/darwin_arm64/
cp terraform-provider-openai ~/.terraform.d/plugins/registry.terraform.io/fjcorp/openai/1.0.0/darwin_arm64/

# Intel-based Mac
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/fjcorp/openai/1.0.0/darwin_amd64/
cp terraform-provider-openai ~/.terraform.d/plugins/registry.terraform.io/fjcorp/openai/1.0.0/darwin_amd64/
```

### For Linux:
```bash
# AMD64 architecture (most common)
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/fjcorp/openai/1.0.0/linux_amd64/
cp terraform-provider-openai ~/.terraform.d/plugins/registry.terraform.io/fjcorp/openai/1.0.0/linux_amd64/

# ARM64 architecture
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/fjcorp/openai/1.0.0/linux_arm64/
cp terraform-provider-openai ~/.terraform.d/plugins/registry.terraform.io/fjcorp/openai/1.0.0/linux_arm64/
```

### For Windows:
```powershell
# Create directory (PowerShell)
New-Item -ItemType Directory -Force -Path "$env:APPDATA\terraform.d\plugins\registry.terraform.io\fjcorp\openai\1.0.0\windows_amd64"

# Copy the executable
Copy-Item "terraform-provider-openai.exe" -Destination "$env:APPDATA\terraform.d\plugins\registry.terraform.io\fjcorp\openai\1.0.0\windows_amd64\"
```

> **Note**: 
> - The `1.0.0` in the path should match the version in your Terraform configuration.
> - If you're unsure about your system's architecture, you can find it using:
>   - macOS/Linux: `uname -m` (x86_64 = amd64, arm64 = arm64)
>   - Windows: System Information or `systeminfo | findstr /C:"System Type"`

After installing, you can use the provider in your Terraform configuration as shown in the Basic Usage section below.

## Basic Usage

```hcl
terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
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

- [Usage Guide](USAGE.md): Detailed usage instructions
- [Implementation Details](IMPLEMENTATION.md): How the provider is implemented
- [Contributing Guide](CONTRIBUTING.md): How to contribute

## License

[MIT](LICENSE)

## Thread Runs

The provider supports the `openai_thread_run` resource, which allows you to create a thread and start a run in a single operation. This simplifies the process of using OpenAI's Assistants API, combining two API calls into one Terraform resource.

### Example Usage

```hcl
resource "openai_thread_run" "example" {
  assistant_id = openai_assistant.example.id
  
  thread {
    messages {
      role    = "user"
      content = "Hello, can you help me with a question?"
    }
  }
  
  instructions = "You are a helpful assistant."
  model        = "gpt-4o"
}
```

You can also use it with an existing thread:

```hcl
resource "openai_thread_run" "on_existing" {
  assistant_id       = openai_assistant.example.id
  existing_thread_id = openai_thread.existing.id
}
```

For more examples and detailed documentation, see:
- [Thread Run Resource Documentation](docs/resources/thread_run.md)
- [Thread Run Data Source Documentation](docs/data-sources/thread_run.md)
- [Thread Run Module Documentation](modules/run/README.md)
- [Thread Run Examples](examples/run/)
