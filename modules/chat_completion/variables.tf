variable "model" {
  description = "ID of the model to use (e.g., gpt-4o, gpt-4, gpt-3.5-turbo)"
  type        = string
}

variable "messages" {
  description = "A list of messages comprising the conversation context"
  type = list(object({
    role    = string
    content = string
    name    = optional(string)
    function_call = optional(object({
      name      = string
      arguments = string
    }))
  }))

  validation {
    condition     = length(var.messages) > 0
    error_message = "At least one message must be provided."
  }

  validation {
    condition = alltrue([
      for msg in var.messages : contains(["system", "user", "assistant", "function", "developer"], msg.role)
    ])
    error_message = "Message roles must be one of: system, user, assistant, function, developer."
  }
}

variable "functions" {
  description = "A list of functions the model may generate JSON inputs for"
  type = list(object({
    name        = string
    description = optional(string)
    parameters  = string
  }))
  default = null
}

variable "function_call" {
  description = "Controls how the model responds to function calls. 'none' means the model doesn't call a function, 'auto' means the model can pick between calling a function or generating a message"
  type        = string
  default     = null
}

variable "temperature" {
  description = "What sampling temperature to use, between 0 and 2. Higher values like 0.8 make output more random, while lower values like 0.2 make it more focused and deterministic"
  type        = number
  default     = 1.0

  validation {
    condition     = var.temperature >= 0 && var.temperature <= 2
    error_message = "Temperature must be between 0 and 2."
  }
}

variable "max_tokens" {
  description = "The maximum number of tokens to generate in the chat completion"
  type        = number
  default     = null
}

variable "top_p" {
  description = "An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass"
  type        = number
  default     = 1.0

  validation {
    condition     = var.top_p > 0 && var.top_p <= 1
    error_message = "Top_p must be between 0 and 1."
  }
}

variable "n" {
  description = "How many chat completion choices to generate for each input message"
  type        = number
  default     = 1
}

variable "stop" {
  description = "Up to 4 sequences where the API will stop generating further tokens"
  type        = list(string)
  default     = null
}

variable "presence_penalty" {
  description = "Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far"
  type        = number
  default     = 0

  validation {
    condition     = var.presence_penalty >= -2 && var.presence_penalty <= 2
    error_message = "Presence penalty must be between -2 and 2."
  }
}

variable "frequency_penalty" {
  description = "Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far"
  type        = number
  default     = 0

  validation {
    condition     = var.frequency_penalty >= -2 && var.frequency_penalty <= 2
    error_message = "Frequency penalty must be between -2 and 2."
  }
}

variable "user" {
  description = "A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse."
  type        = string
  default     = null
}

variable "stream" {
  description = "Whether to stream back partial progress. If set, tokens will be sent as data-only server-sent events as they become available."
  type        = bool
  default     = false
}

variable "logit_bias" {
  description = "Modify the likelihood of specified tokens appearing in the completion. Maps tokens (specified by their token ID) to an associated bias value from -100 to 100."
  type        = map(number)
  default     = null
}

variable "store" {
  description = "Whether to store the chat completion for later retrieval via API. Note: requires a compatible model (e.g., gpt-4o), this parameter set to true, and the Chat Completions Store feature enabled on your OpenAI account. Without these conditions, completions won't be retrievable through the API."
  type        = bool
  default     = false
}

variable "metadata" {
  description = "A map of key-value pairs that can be used to filter chat completions when listing them through the API. Only applicable when store is set to true."
  type        = map(string)
  default     = null
}

variable "imported" {
  description = "Whether this resource was imported from an existing chat completion and should ignore configuration changes"
  type        = bool
  default     = false
} 