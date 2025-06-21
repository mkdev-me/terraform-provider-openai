terraform {
  required_providers {
    openai = {
      source = "mkdev-me/openai"
    }
  }
}

# Chat completion resource
resource "openai_chat_completion" "this" {
  model             = var.model
  temperature       = var.temperature
  max_tokens        = var.max_tokens
  top_p             = var.top_p
  frequency_penalty = var.frequency_penalty
  presence_penalty  = var.presence_penalty
  n                 = var.n
  stream            = var.stream
  store             = var.store
  imported          = var.imported
  stop              = var.stop

  dynamic "messages" {
    for_each = var.messages
    content {
      role    = messages.value["role"]
      content = messages.value["content"]
      name    = lookup(messages.value, "name", null)

      dynamic "function_call" {
        for_each = lookup(messages.value, "function_call", null) != null ? [messages.value.function_call] : []
        content {
          name      = function_call.value.name
          arguments = function_call.value.arguments
        }
      }
    }
  }

  dynamic "functions" {
    for_each = var.functions != null ? var.functions : []
    content {
      name        = functions.value.name
      description = lookup(functions.value, "description", null)
      parameters  = functions.value.parameters
    }
  }

  function_call = var.function_call
  logit_bias    = var.logit_bias
  user          = var.user
  metadata      = var.metadata

  # Prevent recreation of imported resources
  lifecycle {
    ignore_changes = [
      messages,
      temperature,
      model,
      max_tokens,
      top_p,
      frequency_penalty,
      presence_penalty,
      functions,
      function_call,
      store
    ]
  }
}

# Output the completion result
output "completion" {
  description = "The full chat completion response"
  value       = openai_chat_completion.this
}

output "content" {
  description = "The content of the first choice in the chat completion response"
  value       = try(openai_chat_completion.this.choices[0].message[0].content, "")
}

output "function_call" {
  description = "The function call in the chat completion response, if any"
  value       = try(openai_chat_completion.this.choices[0].message[0].function_call, [])
}

output "id" {
  description = "The ID of the chat completion"
  value       = openai_chat_completion.this.id
}

output "created" {
  description = "The Unix timestamp of when the completion was created"
  value       = openai_chat_completion.this.created
}

output "model" {
  description = "The model used for the completion"
  value       = openai_chat_completion.this.model
}

output "choices" {
  description = "The generated completions"
  value       = openai_chat_completion.this.choices
}

output "usage" {
  description = "Token usage statistics"
  value       = openai_chat_completion.this.usage
} 