# Create an assistant for code review
resource "openai_assistant" "code_reviewer" {
  name        = "Code Review Assistant"
  description = "An AI assistant that reviews code and provides feedback"
  model       = "gpt-4-turbo-preview"

  instructions = "You are an expert code reviewer. Analyze code for bugs, security issues, performance problems, and suggest improvements. Be constructive and educational in your feedback."

  tools {
    type = "code_interpreter"
  }

  metadata = {
    version = "1.0"
    team    = "engineering"
  }
}

# Create a thread with code to review
resource "openai_thread" "code_review_thread" {
  messages {
    role    = "user"
    content = "Please review this Python function for security and performance:\n\n```python\ndef process_user_input(user_data):\n    query = f\"SELECT * FROM users WHERE id = {user_data['id']}\"\n    result = db.execute(query)\n    return result\n```"
  }

  metadata = {
    review_type = "security-performance"
    language    = "python"
  }
}

# Create a run to execute the assistant on the thread
resource "openai_run" "code_review_run" {
  thread_id    = openai_thread.code_review_thread.id
  assistant_id = openai_assistant.code_reviewer.id

  # Optional: Override assistant instructions for this specific run
  instructions = "Focus particularly on SQL injection vulnerabilities and suggest secure alternatives."

  # Optional: Override model for this run
  model = "gpt-4-turbo-preview"

  # Optional: Add metadata
  metadata = {
    priority     = "high"
    requested_by = "security-team"
    ticket_id    = "SEC-2024-001"
  }

  # Optional: Configure tools
  tools = [
    {
      type = "code_interpreter"
    }
  ]
}

# Create another run with different parameters
resource "openai_thread" "optimization_thread" {
  messages = [
    {
      role    = "user"
      content = "How can I optimize this sorting algorithm?\n\n```javascript\nfunction bubbleSort(arr) {\n  for (let i = 0; i < arr.length; i++) {\n    for (let j = 0; j < arr.length - 1; j++) {\n      if (arr[j] > arr[j + 1]) {\n        let temp = arr[j];\n        arr[j] = arr[j + 1];\n        arr[j + 1] = temp;\n      }\n    }\n  }\n  return arr;\n}\n```"
    }
  ]
}

resource "openai_run" "optimization_run" {
  thread_id    = openai_thread.optimization_thread.id
  assistant_id = openai_assistant.code_reviewer.id

  instructions = "Provide optimized versions of the code with explanations of the improvements."

  # Optional: Set temperature for more creative suggestions
  temperature = 0.7

  # Optional: Set maximum tokens
  max_tokens = 1000

  metadata = {
    optimization_type = "algorithm"
    complexity_focus  = "time-complexity"
  }
}

# Create a run with file search capability
resource "openai_vector_store" "docs_store" {
  name = "Documentation Store"
}

resource "openai_assistant" "docs_assistant" {
  name  = "Documentation Assistant"
  model = "gpt-4-turbo-preview"

  instructions = "You help users find and understand documentation."

  tools {
    type = "file_search"
  }

  tool_resources = {
    file_search = {
      vector_store_ids = [openai_vector_store.docs_store.id]
    }
  }
}

resource "openai_thread" "docs_thread" {
  messages = [
    {
      role    = "user"
      content = "How do I configure authentication in the system?"
    }
  ]
}

resource "openai_run" "docs_search_run" {
  thread_id    = openai_thread.docs_thread.id
  assistant_id = openai_assistant.docs_assistant.id

  tools {
    type = "file_search"
  }

  metadata = {
    search_type = "documentation"
    topic       = "authentication"
  }
}

# Output run ID
output "code_review_run_id" {
  value = openai_run.code_review_run.id
}
