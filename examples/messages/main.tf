# OpenAI Messages Example
# ==============================
# This example demonstrates creating and managing messages within OpenAI threads,
# including attaching files to messages and using metadata.

terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.1"
    }
  }
}

provider "openai" {
  # API key is pulled from the OPENAI_API_KEY environment variable
  # Organization ID can be set with OPENAI_ORGANIZATION environment variable (optional)
}

# Create a directory for sample files if it doesn't exist
resource "null_resource" "ensure_data_dir" {
  provisioner "local-exec" {
    command = "mkdir -p ${path.module}/data"
  }
}

# Create a sample text file for use with OpenAI Assistants API
resource "local_file" "sample_document" {
  filename = "${path.module}/data/sample_document.txt"
  content  = <<-EOT
This is a sample document for testing OpenAI's Assistants API with files.

It contains information that an Assistant might analyze:

1. The capital of France is Paris
2. The speed of light is approximately 299,792,458 meters per second
3. Water boils at 100 degrees Celsius at standard atmospheric pressure
4. The Earth orbits the Sun at an average distance of about 93 million miles
5. The human body contains about 37 trillion cells
  EOT

  depends_on = [null_resource.ensure_data_dir]
}

# Create a sample CSV data file
resource "local_file" "sample_data" {
  filename = "${path.module}/data/sample_data.csv"
  content  = <<-EOT
Country,Capital,Population,Continent
France,Paris,67391582,Europe
Japan,Tokyo,125836021,Asia
Brazil,Brasília,214326223,South America
Australia,Canberra,25499884,Oceania
Kenya,Nairobi,53771296,Africa
  EOT

  depends_on = [null_resource.ensure_data_dir]
}

# Create a sample JSON data file
resource "local_file" "sample_config" {
  filename = "${path.module}/data/sample_config.json"
  content  = <<-EOT
{
  "application": "OpenAI Terraform Example",
  "version": "1.0.0",
  "settings": {
    "debug_mode": false,
    "max_tokens": 500,
    "temperature": 0.7,
    "model": "gpt-4"
  },
  "features": [
    "file_upload",
    "message_creation",
    "thread_management"
  ]
}
  EOT

  depends_on = [null_resource.ensure_data_dir]
}

# Example 1: Upload files that can be attached to messages
resource "openai_file" "document_file" {
  file    = local_file.sample_document.filename
  purpose = "assistants"

  depends_on = [local_file.sample_document]
}

resource "openai_file" "data_file" {
  file    = local_file.sample_data.filename
  purpose = "assistants"

  depends_on = [local_file.sample_data]
}

resource "openai_file" "config_file" {
  file    = local_file.sample_config.filename
  purpose = "assistants"

  depends_on = [local_file.sample_config]
}

# Create a thread for our messages
resource "openai_thread" "example_thread" {
  metadata = {
    "source"  = "terraform_example"
    "purpose" = "demonstration"
  }
}

# Example 1: Basic message without files
resource "openai_message" "basic_message" {
  thread_id = openai_thread.example_thread.id
  role      = "user"
  content   = "Hello, can you help me analyze some files?"

  metadata = {
    "type"   = "greeting"
    "source" = "terraform"
  }

  # Ensure the thread is created first
  depends_on = [openai_thread.example_thread]
}

# Example 2: Message with a single file attachment
resource "openai_message" "message_with_file" {
  thread_id = openai_thread.example_thread.id
  role      = "user"
  content   = "Please analyze this document and extract the key facts."

  attachments {
    file_id = openai_file.document_file.id
    tools {
      type = "file_search"
    }
  }

  metadata = {
    "type"     = "document_analysis"
    "priority" = "high"
    "source"   = "terraform"
  }

  # Ensure this message is created after the basic message
  depends_on = [openai_message.basic_message, openai_file.document_file]
}



# Example 3: Using the messages module for a simple message
module "module_message" {
  source = "../../modules/messages"

  thread_id = openai_thread.example_thread.id
  content   = "This message was created using the messages module."

  # Uncomment and adjust to add file attachments using the module
  # attachments = [
  #   {
  #     file_id = openai_file.data_file.id
  #     tools = [
  #       {
  #         type = "file_search"
  #       }
  #     ]
  #   }
  # ]

  # Ensure this message is created after the previous messages
  depends_on = [openai_message.message_with_file]
}


# Outputs
output "thread_id" {
  value       = openai_thread.example_thread.id
  description = "The ID of the created thread"
}

output "uploaded_files" {
  value = {
    document_file = {
      id       = openai_file.document_file.id
      filename = openai_file.document_file.filename
      bytes    = openai_file.document_file.bytes
      purpose  = openai_file.document_file.purpose
    }
    data_file = {
      id       = openai_file.data_file.id
      filename = openai_file.data_file.filename
      bytes    = openai_file.data_file.bytes
      purpose  = openai_file.data_file.purpose
    }
    config_file = {
      id       = openai_file.config_file.id
      filename = openai_file.config_file.filename
      bytes    = openai_file.config_file.bytes
      purpose  = openai_file.config_file.purpose
    }
  }
  description = "Details of the uploaded files"
}

output "basic_message" {
  value = {
    id         = openai_message.basic_message.id
    thread_id  = openai_message.basic_message.thread_id
    role       = openai_message.basic_message.role
    content    = openai_message.basic_message.content
    created_at = openai_message.basic_message.created_at
  }
  description = "Details of the basic message"
}

output "message_with_file" {
  value = {
    id          = openai_message.message_with_file.id
    thread_id   = openai_message.message_with_file.thread_id
    role        = openai_message.message_with_file.role
    content     = openai_message.message_with_file.content
    created_at  = openai_message.message_with_file.created_at
    attachments = openai_message.message_with_file.attachments
  }
  description = "Details of the message with file attachment"
}


output "module_message" {
  value = {
    id          = module.module_message.message_id
    content     = module.module_message.content
    created_at  = module.module_message.created_at
    attachments = module.module_message.attachments
  }
  description = "Details of the message created using the module"
}
