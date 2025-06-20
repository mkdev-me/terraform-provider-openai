# OpenAI Assistants Module Data Sources

# Data source to retrieve all assistants
data "openai_assistants" "all" {
  count  = var.enable_assistants_data_source ? 1 : 0
  limit  = var.assistants_limit
  order  = var.assistants_order
  after  = var.assistants_after
  before = var.assistants_before
}

# Data source to retrieve a single assistant by ID
data "openai_assistant" "single" {
  count        = var.enable_single_assistant_data_source && var.single_assistant_id != null ? 1 : 0
  assistant_id = var.single_assistant_id
} 