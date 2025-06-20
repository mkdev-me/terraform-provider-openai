# OpenAI Assistants Module Outputs

# Assistant resource outputs
output "assistant_id" {
  description = "The ID of the created assistant"
  value       = var.enable_assistant ? openai_assistant.this[0].id : null
}

output "assistant_name" {
  description = "The name of the created assistant"
  value       = var.enable_assistant ? openai_assistant.this[0].name : null
}

output "assistant_created_at" {
  description = "The timestamp when the assistant was created"
  value       = var.enable_assistant ? openai_assistant.this[0].created_at : null
}

output "assistant_model" {
  description = "The model used by the assistant"
  value       = var.enable_assistant ? openai_assistant.this[0].model : null
}

output "assistant_description" {
  description = "The description of the assistant"
  value       = var.enable_assistant ? openai_assistant.this[0].description : null
}

output "assistant_instructions" {
  description = "The system instructions of the assistant"
  value       = var.enable_assistant ? openai_assistant.this[0].instructions : null
}

output "assistant_tools" {
  description = "The tools enabled on the assistant"
  value       = var.enable_assistant ? openai_assistant.this[0].tools : null
}

output "assistant_file_ids" {
  description = "The file IDs attached to the assistant"
  value       = var.enable_assistant ? openai_assistant.this[0].file_ids : null
}

output "assistant_metadata" {
  description = "The metadata attached to the assistant"
  value       = var.enable_assistant ? openai_assistant.this[0].metadata : null
}

# Assistants data source outputs
output "all_assistants" {
  description = "List of all assistants retrieved by the data source"
  value       = var.enable_assistants_data_source ? data.openai_assistants.all[0].assistants : null
}

output "all_assistants_count" {
  description = "Count of all assistants retrieved by the data source"
  value       = var.enable_assistants_data_source ? length(data.openai_assistants.all[0].assistants) : null
}

output "first_assistant_id" {
  description = "The ID of the first assistant in the list"
  value       = var.enable_assistants_data_source ? data.openai_assistants.all[0].first_id : null
}

output "last_assistant_id" {
  description = "The ID of the last assistant in the list"
  value       = var.enable_assistants_data_source ? data.openai_assistants.all[0].last_id : null
}

output "has_more_assistants" {
  description = "Whether there are more assistants available beyond the current list"
  value       = var.enable_assistants_data_source ? data.openai_assistants.all[0].has_more : null
}

# Single assistant data source outputs
output "single_assistant" {
  description = "Details of a specific assistant retrieved by ID"
  value       = var.enable_single_assistant_data_source && var.single_assistant_id != null ? data.openai_assistant.single[0] : null
}

output "single_assistant_id" {
  description = "ID of the specific assistant retrieved"
  value       = var.enable_single_assistant_data_source && var.single_assistant_id != null ? data.openai_assistant.single[0].id : null
}

output "single_assistant_name" {
  description = "Name of the specific assistant retrieved"
  value       = var.enable_single_assistant_data_source && var.single_assistant_id != null ? data.openai_assistant.single[0].name : null
}

output "single_assistant_model" {
  description = "Model of the specific assistant retrieved"
  value       = var.enable_single_assistant_data_source && var.single_assistant_id != null ? data.openai_assistant.single[0].model : null
}

output "single_assistant_instructions" {
  description = "Instructions of the specific assistant retrieved"
  value       = var.enable_single_assistant_data_source && var.single_assistant_id != null ? data.openai_assistant.single[0].instructions : null
} 