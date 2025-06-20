# This example demonstrates the complete workflow of:
# 1. Creating an assistant
# 2. Reading it with a data source
# 3. Structure for importing existing assistants

# 1. Create a custom assistant
resource "openai_assistant" "customer_support" {
  name         = "Customer Support Assistant"
  model        = "gpt-4o"
  instructions = "You are a helpful customer support assistant. Answer customer questions about our product based on the documents you have access to."

  # Enable file search for accessing knowledge base
  tools {
    type = "file_search"
  }

  # Files would typically be uploaded separately and referenced here
  # file_ids = [openai_file.support_docs.id]

  description = "Assistant for customer support queries"
  metadata = {
    "team"     = "customer_success",
    "priority" = "high"
  }
}

# 2. Read the assistant details using a data source
data "openai_assistant" "customer_support_info" {
  # Reference the ID of the created assistant
  assistant_id = openai_assistant.customer_support.id

  # The data source will automatically retrieve all properties
  # This is useful for accessing computed properties
}

# 3. Structure for importing an assistant (commented out for demonstration)
# Use this approach to import an existing assistant:
/*
resource "openai_assistant" "imported_support" {
  # These values are required but will be overwritten on import
  name  = "Imported Assistant"
  model = "gpt-4o"
  
  # After importing, you could update certain properties
  metadata = {
    "managed_by" = "terraform",
    "imported_on" = "2023-07-01"
  }
}
*/
# To import: terraform import openai_assistant.imported_support asst_abc123

# Output values for demonstration
output "created_assistant_id" {
  value = openai_assistant.customer_support.id
}

output "created_at_timestamp" {
  value = openai_assistant.customer_support.created_at
}

output "data_source_name" {
  value = data.openai_assistant.customer_support_info.name
}

output "data_source_model" {
  value = data.openai_assistant.customer_support_info.model
} 