output "id" {
  description = "Unique identifier for the response"
  value       = openai_model_response.this.id
}

output "created" {
  description = "Unix timestamp when the response was created"
  value       = openai_model_response.this.created
}

output "object" {
  description = "Object type (usually 'model_response')"
  value       = openai_model_response.this.object
}

output "output" {
  description = "The generated output containing text and token count"
  value       = openai_model_response.this.output
}

output "output_text" {
  description = "The generated response text"
  value       = openai_model_response.this.output.text
}

output "token_count" {
  description = "Number of tokens in the output"
  value       = openai_model_response.this.output.token_count
}

output "usage" {
  description = "Token usage statistics for the request"
  value       = openai_model_response.this.usage
}

output "finish_reason" {
  description = "Reason why the response finished (e.g., stop, length, content)"
  value       = openai_model_response.this.finish_reason
} 