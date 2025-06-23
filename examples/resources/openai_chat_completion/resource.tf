resource "openai_chat_completion" "example" {
  model = "gpt-4o-mini"

  messages {
    role    = "system"
    content = "You are a helpful assistant."
  }

  messages {
    role    = "user"
    content = "Hello! What's the weather like today?"
  }

  temperature = 0.7
  max_tokens  = 150
}

output "chat_response" {
  value = openai_chat_completion.example.choices[0].message[0].content
}
