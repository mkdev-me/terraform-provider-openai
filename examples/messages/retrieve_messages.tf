# OpenAI Message Retrieval Examples
# ==============================
# This example demonstrates how to retrieve messages from OpenAI threads
# using data sources and the messages module in data source mode.
#
# NOTE: This example assumes you already have messages created.
# You should first apply main.tf to create the messages, then
# uncomment and apply the code below in a second phase.

# =========================================================
# IMPORTANT: First phase - Create messages with main.tf
# After running terraform apply on main.tf, copy the message IDs 
# from the output and replace the placeholder values below.
# =========================================================

# Uncomment and update with actual IDs after first phase
/*
# Retrieve a specific message by ID using the data source directly
data "openai_message" "specific_message" {
  # Replace these with actual IDs from the first phase
  thread_id  = "thread_123456789"  # Replace with actual thread ID
  message_id = "msg_123456789"     # Replace with actual message ID
}

# List all messages in a thread
data "openai_messages" "thread_messages" {
  # Replace with actual thread ID from first phase
  thread_id = "thread_123456789"
  limit     = 5
  order     = "desc"  # Most recent messages first
}

# Using the messages module in data source mode to retrieve an existing message
module "retrieve_message" {
  source = "../../modules/messages"
  
  use_data_source     = true
  # Replace these with actual IDs from the first phase
  thread_id           = "thread_123456789"  # Replace with actual thread ID
  existing_message_id = "msg_123456789"     # Replace with actual message ID
}

# Outputs for retrieved message using direct data source
output "retrieved_message" {
  value = {
    id           = data.openai_message.specific_message.id
    thread_id    = data.openai_message.specific_message.thread_id
    role         = data.openai_message.specific_message.role
    content      = data.openai_message.specific_message.content
    created_at   = data.openai_message.specific_message.created_at
    attachments  = data.openai_message.specific_message.attachments
  }
  description = "Details of a specific message retrieved using the data source"
}

# Outputs for message list
output "message_list" {
  value = {
    count    = length(data.openai_messages.thread_messages.messages)
    first_id = data.openai_messages.thread_messages.first_id
    last_id  = data.openai_messages.thread_messages.last_id
    has_more = data.openai_messages.thread_messages.has_more
    messages = [for msg in data.openai_messages.thread_messages.messages : {
      id      = msg.id
      role    = msg.role
      content = msg.content
    }]
  }
  description = "Details of messages retrieved from the thread"
}

# Outputs for retrieved message using module
output "module_retrieved_message" {
  value = {
    id         = module.retrieve_message.message_id
    content    = module.retrieve_message.content
    attachments = module.retrieve_message.attachments
    created_at = module.retrieve_message.created_at
    metadata   = module.retrieve_message.metadata
  }
  description = "Details of the message retrieved using the module in data source mode"
}
*/ 