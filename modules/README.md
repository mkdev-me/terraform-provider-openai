# OpenAI Terraform Provider Modules

This directory contains reusable Terraform modules for common OpenAI resource patterns. Each module provides a simplified interface for managing specific OpenAI resources with best practices built-in.

## Quick Start

```hcl
module "chat" {
  source = "github.com/fjcorp/terraform-provider-openai//modules/chat_completion"
  
  model = "gpt-4"
  messages = [
    {
      role    = "system"
      content = "You are a helpful assistant."
    },
    {
      role    = "user"
      content = "Hello!"
    }
  ]
}

output "response" {
  value = module.chat.content
}
```

## Available Modules

### Core AI Modules

| Module | Description | Key Features |
|--------|-------------|--------------|
| [chat_completion](./chat_completion/) | Generate conversational responses | Message history, function calling, streaming |
| [assistants](./assistants/) | Create AI assistants | Custom instructions, tools, file knowledge |
| [embeddings](./embeddings/) | Generate text embeddings | Semantic search, clustering support |
| [fine_tuning](./fine_tuning/) | Train custom models | Hyperparameter control, checkpoint management |
| [model_response](./model_response/) | Generate model responses | Token tracking, usage statistics |

### Content Generation

| Module | Description | Key Features |
|--------|-------------|--------------|
| [audio](./audio/) | Audio processing | Speech-to-text, text-to-speech, translation |
| [image](./image/) | Image operations | Generation, editing, variations with DALL-E |
| [moderation](./moderation/) | Content moderation | Policy compliance, safety checks |

### Data Management

| Module | Description | Key Features |
|--------|-------------|--------------|
| [files](./files/) | File management | Upload, import, lifecycle management |
| [upload](./upload/) | Simplified uploads | Import support, change detection |
| [vector_store](./vector_store/) | Vector databases | RAG implementation, file indexing |
| [batch](./batch/) | Batch processing | Bulk operations, async job management |

### Administrative

| Module | Description | Key Features |
|--------|-------------|--------------|
| [projects](./projects/) | Project management | Project creation, configuration |
| [project_user](./project_user/) | User access control | Role assignment, permissions |
| [invite](./invite/) | User invitations | Organization invites, workflow automation |
| [organization_users](./organization_users/) | Org user management | User listing, filtering |

### API Management

| Module | Description | Key Features |
|--------|-------------|--------------|
| [project_api](./project_api/) | Project API keys | Import existing keys, lifecycle management |
| [system_api](./system_api/) | Admin API keys | Create admin keys, scope management |
| [service_account](./service_account/) | Service accounts | Automated access, key rotation |
| [rate_limit](./rate_limit/) | Rate limiting | Request/token limits, cost control |

### Advanced Features

| Module | Description | Key Features |
|--------|-------------|--------------|
| [threads](./threads/) | Conversation threads | Message history, context management |
| [messages](./messages/) | Thread messages | Attachments, metadata |
| [run](./run/) | Assistant execution | Tool execution, status tracking |
| [model](./model/) | Model information | Get model details, capabilities |

## Module Patterns

### Authentication

All modules inherit authentication from the provider:

```hcl
provider "openai" {
  api_key   = var.openai_api_key    # Or OPENAI_API_KEY env var
  admin_key = var.openai_admin_key  # For admin operations
}
```

### Common Usage Pattern

```hcl
# 1. Call the module
module "my_resource" {
  source = "./modules/module_name"
  
  # Required inputs
  input_1 = "value1"
  
  # Optional inputs with defaults
  input_2 = "custom_value"
}

# 2. Use the outputs
output "result" {
  value = module.my_resource.output_name
}
```

### Error Handling

Modules include proper error handling:

```hcl
module "safe_upload" {
  source = "./modules/files"
  
  file_path = "data.jsonl"
  purpose   = "fine-tune"
  
  # Prevents accidental deletion
  lifecycle {
    prevent_destroy = true
  }
}
```

## Featured Workflows

### Fine-Tuning Workflow

```hcl
# 1. Upload training data
module "training_file" {
  source    = "./modules/files"
  file_path = "training.jsonl"
  purpose   = "fine-tune"
}

# 2. Create fine-tuning job
module "fine_tune" {
  source        = "./modules/fine_tuning"
  model         = "gpt-3.5-turbo"
  training_file = module.training_file.file_id
}

# 3. Use the model
module "inference" {
  source = "./modules/chat_completion"
  model  = module.fine_tune.fine_tuned_model_id
  messages = [...]
}
```

### Import Existing Resources

```hcl
# Import an existing file
module "existing_file" {
  source    = "./modules/upload"
  purpose   = "assistants"
  file_path = "placeholder.txt"  # Ignored for imports
}

# Run import command
# terraform import module.existing_file.openai_file.file file-abc123
```

### Organization User Management

```hcl
# 1. Invite user
module "invite" {
  source = "./modules/invite"
  email  = "user@example.com"
  role   = "reader"
}

# 2. Find user after acceptance
data "openai_organization_user" "user" {
  email = "user@example.com"
}

# 3. Add to project
module "project_access" {
  source     = "./modules/project_user"
  project_id = "proj_123"
  user_id    = data.openai_organization_user.user.id
  role       = "member"
}
```

## Best Practices

1. **Version Pinning**: Pin module versions in production
   ```hcl
   source = "github.com/fjcorp/terraform-provider-openai//modules/chat_completion?ref=v1.0.0"
   ```

2. **State Management**: Use remote state for team collaboration
   ```hcl
   terraform {
     backend "s3" {
       bucket = "terraform-state"
       key    = "openai/terraform.tfstate"
     }
   }
   ```

3. **Secret Management**: Never commit API keys
   ```hcl
   variable "openai_api_key" {
     description = "OpenAI API Key"
     type        = string
     sensitive   = true
   }
   ```

4. **Cost Control**: Set appropriate rate limits
   ```hcl
   module "rate_limit" {
     source     = "./modules/rate_limit"
     project_id = "proj_123"
     model      = "gpt-4"
     rpm_limit  = 100
     tpm_limit  = 10000
   }
   ```

## Module Development

When creating new modules:

1. **Structure**: Follow existing module patterns
2. **Documentation**: Include comprehensive README
3. **Examples**: Provide working examples
4. **Variables**: Use clear, descriptive names
5. **Outputs**: Export all useful attributes
6. **Validation**: Add input validation where appropriate

Example module structure:
```
module_name/
├── README.md
├── main.tf
├── variables.tf
├── outputs.tf
└── versions.tf
```

## Known Limitations

### Audio Resources
The OpenAI API doesn't support listing audio resources. Individual resources can be retrieved by ID, but list operations return empty results.

### Project API Keys
Project API keys cannot be created via API - they must be created manually in the OpenAI dashboard and then imported.

### Invitation Workflow
Project assignments in invitations don't work through the API. Use the three-step workflow: invite → wait for acceptance → add to project.

## Support

- [Examples](../examples/): Complete working examples
- [Documentation](../docs/): Provider documentation
- [Issues](https://github.com/fjcorp/terraform-provider-openai/issues): Report issues

## Contributing

Contributions welcome! Please ensure:
- Modules follow established patterns
- Documentation is comprehensive
- Examples are provided
- Tests pass successfully