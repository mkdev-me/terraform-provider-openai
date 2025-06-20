# OpenAI Terraform Provider Examples

This directory contains comprehensive examples demonstrating how to use the OpenAI Terraform Provider for various use cases.

## Prerequisites

- Terraform >= 1.0
- OpenAI API key with appropriate permissions
- For organization features: Admin API key

## Getting Started

1. **Set up authentication**:
   ```bash
   export OPENAI_API_KEY="sk-proj-..."
   export OPENAI_ADMIN_KEY="sk-..."  # For organization features
   ```

2. **Navigate to an example**:
   ```bash
   cd examples/chat_completion
   ```

3. **Initialize and apply**:
   ```bash
   terraform init
   terraform apply
   ```

## Available Examples

### Core AI Features

| Example | Description | Key Features |
|---------|-------------|--------------|
| [chat_completion](./chat_completion/) | Chat completions with GPT models | Basic chat, function calling, streaming |
| [assistants](./assistants/) | AI assistants with custom instructions | Assistant creation, file knowledge, tools |
| [embeddings](./embeddings/) | Text embeddings for semantic search | Vector generation, similarity search |
| [fine_tuning](./fine_tuning/) | Custom model training | Training job creation, checkpoint management |
| [model_response](./model_response/) | Model responses with detailed metrics | Token usage, cost tracking |

### Content Generation

| Example | Description | Key Features |
|---------|-------------|--------------|
| [image](./image/) | Image generation with DALL-E | Generation, editing, variations |
| [audio](./audio/) | Audio processing | Speech-to-text, text-to-speech |
| [moderation](./moderation/) | Content moderation | Safety checks, policy compliance |

### Data Management

| Example | Description | Key Features |
|---------|-------------|--------------|
| [files](./files/) | File management | Upload, import, organization |
| [upload](./upload/) | File upload workflows | Bulk upload, import existing |
| [vector_store](./vector_store/) | Vector databases | RAG implementation, file indexing |
| [batch](./batch/) | Batch operations | Bulk processing, async jobs |

### Administrative

| Example | Description | Key Features |
|---------|-------------|--------------|
| [projects](./projects/) | Project management | Project creation, configuration |
| [project_user](./project_user/) | User access control | Role assignment, permissions |
| [organization_users](./organization_users/) | Organization management | User listing, filtering |
| [invite](./invite/) | User invitations | Invite workflow, role setup |
| [rate_limit](./rate_limit/) | API rate limiting | Limit configuration, monitoring |

### API Management

| Example | Description | Key Features |
|---------|-------------|--------------|
| [project_api](./project_api/) | Project API keys | Key management, rotation |
| [system_api](./system_api/) | System API configuration | Admin key management |
| [service_account](./service_account/) | Service accounts | Automated access, key management |

### Advanced Examples

| Example | Description | Key Features |
|---------|-------------|--------------|
| [threads](./threads/) | Conversation threads | Thread management, message history |
| [messages](./messages/) | Message management | Thread messages, attachments |
| [run](./run/) | Assistant execution | Run management, tool execution |

## Authentication Configuration

### Environment Variables (Recommended)

```bash
export OPENAI_API_KEY="sk-proj-..."      # Project API key
export OPENAI_ADMIN_KEY="sk-..."         # Admin key for org features
export OPENAI_ORGANIZATION_ID="org-..."  # Optional org ID
```

### Provider Configuration

```hcl
provider "openai" {
  api_key   = var.openai_api_key
  admin_key = var.openai_admin_key
}
```

### Resource-Level Override

```hcl
resource "openai_chat_completion" "example" {
  api_key = var.specific_project_key  # Override provider default
  model   = "gpt-4"
  # ...
}
```

## Common Patterns

### Error Handling

Most examples include error handling patterns:

```hcl
resource "openai_file" "training" {
  filename = "training.jsonl"
  purpose  = "fine-tune"
  content  = file("${path.module}/data/training.jsonl")
  
  lifecycle {
    prevent_destroy = true  # Prevent accidental deletion
  }
}
```

### Output Management

Examples demonstrate useful outputs:

```hcl
output "assistant_id" {
  value       = openai_assistant.example.id
  description = "The ID of the created assistant"
}

output "total_tokens" {
  value       = openai_chat_completion.example.usage.total_tokens
  description = "Total tokens used"
}
```

## Testing Examples

Use the provided test script:

```bash
# Test all examples (plan only)
./testing/test_examples.sh plan

# Test specific example
./testing/test_examples.sh plan chat_completion

# Apply and destroy (creates real resources)
./testing/test_examples.sh apply chat_completion
```

## Contributing

When adding new examples:

1. Create a new directory with a descriptive name
2. Include a comprehensive README.md
3. Provide working Terraform configuration
4. Add sample data files if needed
5. Include outputs demonstrating the results
6. Test thoroughly before submitting

## Support

- [Documentation](../docs/)
- [GitHub Issues](https://github.com/mkdev-me/terraform-provider-openai/issues)
- [API Reference](https://platform.openai.com/docs/api-reference)