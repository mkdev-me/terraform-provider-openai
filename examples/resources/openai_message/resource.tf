# First, create a thread to hold the conversation
resource "openai_thread" "conversation" {
  metadata = {
    topic = "technical-support"
    user  = "user-12345"
  }
}

# Add a simple text message to the thread
resource "openai_message" "user_question" {
  thread_id = openai_thread.conversation.id
  role      = "user"
  content   = "How do I authenticate with the OpenAI API using Python?"

  # Optional: Add metadata to the message
  metadata = {
    source    = "web-chat"
    timestamp = "2024-01-15T10:30:00Z"
    priority  = "normal"
  }
}

# Add a message with code content
resource "openai_message" "code_example" {
  thread_id = openai_thread.conversation.id
  role      = "user"
  content   = "Here's my current code that's not working:\n\n```python\nimport openai\n\nclient = openai.Client()\nresponse = client.completions.create(\n    model='gpt-3.5-turbo',\n    prompt='Hello world'\n)\n```\n\nWhat's wrong with it?"

  metadata = {
    contains_code = "true"
    language      = "python"
  }
}

# Add a message with file attachments (commented out due to schema compatibility issues)
# resource "openai_file" "error_log" {
#   file    = "error_log.txt"
#   purpose = "assistants"
# }
# 
# resource "openai_message" "with_attachment" {
#   thread_id = openai_thread.conversation.id
#   role      = "user"
#   content   = "I'm attaching the error log from my application. Can you help me understand what's going wrong?"
# 
#   # Attach files to the message
#   attachments = [
#     {
#       file_id = openai_file.error_log.id
#       tools = [
#         {
#           type = "file_search"
#         }
#       ]
#     }
#   ]
# 
#   metadata = {
#     attachment_type = "log"
#     urgency         = "high"
#   }
# }

# Simple message without attachments
resource "openai_message" "simple_message" {
  thread_id = openai_thread.conversation.id
  role      = "user"
  content   = "This is a simple message without file attachments."

  metadata = {
    message_type = "simple"
    urgency      = "low"
  }
}

# Add an assistant response
resource "openai_message" "assistant_response" {
  thread_id = openai_thread.conversation.id
  role      = "assistant"
  content   = "I can see the issue with your code. You're using an outdated API pattern. Here's the correct way to authenticate with the OpenAI API in Python:\n\n```python\nfrom openai import OpenAI\n\n# Initialize the client with your API key\nclient = OpenAI(api_key='your-api-key-here')\n\n# For chat completions, use the chat.completions endpoint\nresponse = client.chat.completions.create(\n    model='gpt-3.5-turbo',\n    messages=[\n        {'role': 'user', 'content': 'Hello world'}\n    ]\n)\n\nprint(response.choices[0].message.content)\n```\n\nThe main issues were:\n1. You need to import `OpenAI` class, not the module directly\n2. Use `chat.completions` instead of `completions` for chat models\n3. Pass messages array instead of a prompt string"

  metadata = {
    model_used    = "gpt-4"
    helpful_score = "5"
  }
}

# Output message ID
output "user_message_id" {
  value = openai_message.user_question.id
}
