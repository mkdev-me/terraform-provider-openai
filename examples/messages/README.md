# OpenAI Messages Example

This example demonstrates how to create and retrieve messages in OpenAI threads using Terraform. It showcases the creation of messages with and without file attachments, using both direct resource declarations and the messages module.

## Two-Phase Application Process

Due to limitations in Terraform's handling of computed values in `for_each` expressions, working with OpenAI messages requires a two-phase approach:

1. **Phase 1: Create Messages** - Create threads, upload files, and create messages
2. **Phase 2: Retrieve Messages** - Use data sources to retrieve and interact with the created messages

## Prerequisites

- An OpenAI API key with access to the Assistants API
- Terraform installed on your system

## Usage

### Message Creation

The example demonstrates how to create messages with file attachments in several ways:

1. Direct reference in the resource:
   ```hcl
   resource "openai_message" "message_with_file" {
     thread_id = openai_thread.example_thread.id
     role      = "user"
     content   = "Please analyze this document."
     attachments {
       file_id = openai_file.document_file.id
       tools   = ["retrieval"]
     }
   }
   ```

2. Using the messages module:
   ```hcl
   module "module_message_with_files" {
     source = "../../modules/messages"
     thread_id = openai_thread.example_thread.id
     content   = "This message with files was created using the messages module."
     attachments = [
       {
         file_id = openai_file.document_file.id
         tools   = ["retrieval"]
       },
       {
         file_id = openai_file.data_file.id
         tools   = ["retrieval"]
       }
     ]
   }
   ```

### Message Retrieval

1. Retrieving a specific message by ID:
   ```hcl
   data "openai_message" "specific_message" {
     thread_id  = openai_thread.example_thread.id
     message_id = openai_message.basic_message.id
   }
   ```

2. Using the module to retrieve a message:
   ```hcl
   module "retrieve_message" {
     source = "../../modules/messages"
     
     use_data_source     = true
     thread_id           = openai_thread.example_thread.id
     existing_message_id = openai_message.basic_message.id
   }
   ```

3. Listing messages in a thread:
   ```hcl
   data "openai_messages" "thread_messages" {
     thread_id = openai_thread.example_thread.id
     limit     = 10
     order     = "desc"  # Most recent first
   }
   ```

### Working with Metadata

Messages can include metadata for better organization:

```hcl
resource "openai_message" "with_metadata" {
  thread_id = openai_thread.example_thread.id
  role      = "user"
  content   = "This message includes metadata."
  
  metadata = {
    source      = "terraform"
    importance  = "high"
    category    = "example"
  }
}
```

## Running the Example

```bash
# Set your OpenAI API Key
export OPENAI_API_KEY="your-api-key-here"

# Initialize Terraform
terraform init

# Apply the Configuration
terraform apply

# Clean Up When Done
terraform destroy
```

## Outputs

The example provides outputs for both created resources and retrieved data:

- `thread_id`: ID of the created thread
- `uploaded_files`: Details of the uploaded files
- `basic_message`: Details of the basic message
- `message_with_file`: Details of the message with file attachment
- `module_message`: Details of the message created using the module
- `retrieved_message`: Details of the specific message retrieved using the data source
- `module_retrieved_message`: Details of the message retrieved using the module 