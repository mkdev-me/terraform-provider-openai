variable "input" {
  description = "The input text to generate a response for"
  type        = string
}

variable "model" {
  description = "ID of the model to use (e.g., 'gpt-4o', 'gpt-4-turbo')"
  type        = string
}

variable "max_output_tokens" {
  description = "The maximum number of tokens to generate"
  type        = number
  default     = null
}

variable "temperature" {
  description = "Sampling temperature between 0 and 2. Higher values mean more randomness"
  type        = number
  default     = 0.7
}

variable "top_p" {
  description = "Nucleus sampling parameter. Top probability mass to consider"
  type        = number
  default     = null
}

variable "top_k" {
  description = "Top-k sampling parameter. Only consider top k tokens"
  type        = number
  default     = null
}

variable "include" {
  description = "Optional fields to include in the response"
  type        = list(string)
  default     = null
}

variable "instructions" {
  description = "Optional instructions to guide the model"
  type        = string
  default     = null
}

variable "stop_sequences" {
  description = "Optional list of sequences where the API will stop generating further tokens"
  type        = list(string)
  default     = null
}

variable "frequency_penalty" {
  description = "Penalty for token frequency between -2.0 and 2.0"
  type        = number
  default     = 0
}

variable "presence_penalty" {
  description = "Penalty for token presence between -2.0 and 2.0"
  type        = number
  default     = 0
}

variable "user" {
  description = "A unique identifier representing the end-user, to help track and detect abuse"
  type        = string
  default     = null
} 