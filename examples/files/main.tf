# OpenAI File Management Examples
# ==============================
# This example demonstrates uploading and managing files with OpenAI, 
# showing various purposes and use cases for files.

terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = "1.0.0"
    }
  }
}

provider "openai" {
  # API key is pulled from the OPENAI_API_KEY environment variable
  # Organization ID can be set with OPENAI_ORGANIZATION_ID (optional)
}

# Create a local directory for sample files
resource "local_file" "fine_tune_sample" {
  filename = "${path.module}/data/fine_tune_sample.jsonl"
  content  = <<-EOT
{"prompt": "What is the capital of France?", "completion": "Paris"}
{"prompt": "What is the capital of Italy?", "completion": "Rome"}
{"prompt": "What is the capital of Japan?", "completion": "Tokyo"}
  EOT
}

resource "local_file" "batch_sample" {
  filename = "${path.module}/data/batch_sample.jsonl"
  content  = <<-EOT
{"model": "text-embedding-ada-002", "input": "The food was delicious and the service was excellent."}
{"model": "text-embedding-ada-002", "input": "I had a terrible experience at the restaurant."}
  EOT
}

resource "local_file" "assistants_sample" {
  filename = "${path.module}/data/assistants_sample.txt"
  content  = <<-EOT
This is a sample text file for use with OpenAI Assistants API.

It contains information about world capitals:
- France: Paris
- Italy: Rome
- Japan: Tokyo
- Spain: Madrid
- Germany: Berlin
  EOT
}

# Example 1: Upload a file for fine-tuning
resource "openai_file" "fine_tune_file" {
  file    = local_file.fine_tune_sample.filename
  purpose = "fine-tune"

  # Ensure the file is created before uploading
  depends_on = [local_file.fine_tune_sample]
}

# Example 2: Upload a file for batch processing
resource "openai_file" "batch_file" {
  file    = local_file.batch_sample.filename
  purpose = "batch"

  # Ensure the file is created before uploading
  depends_on = [local_file.batch_sample]
}

# Example 3: Using the files module for a fine-tuning file
module "training_file" {
  source = "../../modules/files"

  file_path = local_file.fine_tune_sample.filename
  purpose   = "fine-tune"

  # Ensure the file is created before uploading
  depends_on = [local_file.fine_tune_sample]
}

# Example 4: Create a file for assistants
resource "openai_file" "assistants_file" {
  file    = local_file.assistants_sample.filename
  purpose = "assistants"

  # Ensure the file is created before uploading
  depends_on = [local_file.assistants_sample]
}

# Example 5: Using the module with data source mode to retrieve an existing file
# Note: This requires an existing file ID - replace with your actual file ID
# The file_id can be from a previously created file or from another example
module "existing_file" {
  source = "../../modules/files"

  use_data_source = true
  file_id         = openai_file.assistants_file.id # Using the assistants file we just created

  # This will wait until the file is created by the previous resource
  depends_on = [openai_file.assistants_file]
}

# Outputs
output "fine_tune_file_id" {
  value = openai_file.fine_tune_file.id
}

output "batch_file_id" {
  value = openai_file.batch_file.id
}

output "module_file_id" {
  value = module.training_file.file_id
}

output "assistants_file_details" {
  value = {
    id         = openai_file.assistants_file.id
    filename   = openai_file.assistants_file.filename
    purpose    = openai_file.assistants_file.purpose
    bytes      = openai_file.assistants_file.bytes
    created_at = openai_file.assistants_file.created_at
  }
}

output "retrieved_file_details" {
  value = {
    id         = module.existing_file.file_id
    filename   = module.existing_file.filename
    purpose    = module.existing_file.purpose
    bytes      = module.existing_file.bytes
    created_at = module.existing_file.created_at
  }
  description = "Details of the file retrieved using data source mode"
}
