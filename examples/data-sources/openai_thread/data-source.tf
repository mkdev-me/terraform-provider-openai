# Fetch a specific thread by ID
data "openai_thread" "customer_support" {
  thread_id = "thread_abc123"
}

# Output thread creation timestamp
output "thread_created_at" {
  value = data.openai_thread.customer_support.created_at
}

# Use thread data in a message creation
resource "openai_thread_message" "follow_up" {
  thread_id = data.openai_thread.customer_support.id
  role      = "assistant"
  content   = "Following up on your support request..."
}
