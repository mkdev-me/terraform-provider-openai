module "basic_chat_with_store" {
  source = "../../modules/chat_completion"

  model       = "gpt-4o" # Compatible model that supports the Chat Completions Store feature
  max_tokens  = 300
  temperature = 0.7
  store       = true # Required for using the Chat Completions Store (if enabled on your account)

  messages = [
    {
      role    = "system"
      content = "You are a helpful assistant specialized in explaining complex topics in simple terms."
    },
    {
      role    = "user"
      content = "Explain the concept of cloud computing to me like I'm 10 years old."
    }
  ]
}

output "basic_chat_with_store_response" {
  description = "The assistant's response for the basic chat with store example"
  value       = module.basic_chat_with_store.content
} 