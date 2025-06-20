# OpenAI Threads Module Data Sources

# Data source to retrieve a specific thread by ID
data "openai_thread" "single" {
  count     = var.enable_thread_data_source && var.thread_id != null ? 1 : 0
  thread_id = var.thread_id
} 