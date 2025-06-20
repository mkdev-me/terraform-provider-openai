# OpenAI Threads Terraform Module

This module provides a simplified interface for working with OpenAI Threads using the OpenAI Terraform Provider.

## Features

- Create threads with or without initial messages
- Configure thread metadata
- Retrieve existing threads by ID
- Convenient outputs for all thread properties

## Usage

```hcl
module "empty_thread" {
  source = "../../modules/threads"
  
  # This creates an empty thread with no messages or metadata
}

module "thread_with_messages" {
  source = "../../modules/threads"
  
  # Add initial messages
  thread_messages = [
    {
      role     = "user"
      content  = "Hello, can you help me with something?"
      metadata = {
        "priority" = "high"
      }
    },
    {
      role    = "user"
      content = "I need information about quantum computing."
    }
  ]
  
  # Add thread metadata
  thread_metadata = {
    "category" = "science",
    "source"   = "terraform"
  }
}

# Output the thread ID
output "thread_id" {
  value = module.thread_with_messages.thread_id
}
```

## Retrieving an Existing Thread

```hcl
module "existing_thread" {
  source = "../../modules/threads"
  
  # Disable thread creation
  enable_thread = false
  
  # Enable thread data source
  enable_thread_data_source = true
  thread_id = "thread_abc123xyz"  # Replace with your thread ID
}

# Output details about the existing thread
output "thread_metadata" {
  value = module.existing_thread.single_thread_metadata
}
```

## Importing an Existing Thread

You can also import threads through Terraform:

1. First, create a placeholder resource:
```hcl
resource "openai_thread" "imported" {
  # No configuration needed for import
}
```

2. Import the thread:
```bash
terraform import openai_thread.imported thread_abc123
```

3. Then use the module with the imported resource:
```hcl
module "use_imported" {
  source = "../../modules/threads"
  
  # Disable thread creation
  enable_thread = false
  
  # Enable thread data source
  enable_thread_data_source = true
  thread_id = openai_thread.imported.id
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 0.13.0 |
| openai | >= 1.0.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| enable_thread | Whether to create the thread | `bool` | `true` | no |
| thread_messages | List of initial messages to include in the thread | `list(object)` | `[]` | no |
| thread_metadata | Metadata for the thread | `map(string)` | `{}` | no |
| enable_thread_data_source | Whether to fetch a thread by ID | `bool` | `false` | no |
| thread_id | ID of a specific thread to fetch | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| thread_id | The ID of the created thread |
| thread_created_at | The timestamp when the thread was created |
| thread_metadata | The metadata attached to the thread |
| single_thread | Details of a specific thread retrieved by ID |
| single_thread_id | ID of the specific thread retrieved |
| single_thread_created_at | Timestamp when the specific thread was created |
| single_thread_metadata | Metadata of the specific thread retrieved | 