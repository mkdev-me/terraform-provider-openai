# Data source for retrieving a single assistant by ID
data "openai_assistant" "math_tutor" {
  # This will reference the ID of the created assistant
  assistant_id = openai_assistant.math_tutor.id
}

# Outputs for the single assistant data source
output "data_math_tutor_name" {
  description = "The name of the math tutor retrieved from the data source"
  value       = data.openai_assistant.math_tutor.name
}

output "data_math_tutor_model" {
  description = "The model used by the math tutor retrieved from the data source"
  value       = data.openai_assistant.math_tutor.model
}

output "data_math_tutor_instructions" {
  description = "The instructions for the math tutor retrieved from the data source"
  value       = data.openai_assistant.math_tutor.instructions
} 